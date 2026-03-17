<?php

namespace App\Jobs;

use App\Services\RateLimitConfig;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\PendingDispatch;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Support\Facades\Redis;

/**
 * 基于 Laravel RateLimiter 的限流 Job 基类
 * 执行时检查限流与风控挂起，超限则 release 延迟重试，不再在派发时排期到未来时间
 * 通过滑动窗口有一个弊端，就是当排队任务过多时，会被窗口过期时瞬时爆发执行多个任务，有可能会触发风控，先运行一段时间看看
 */
abstract class BaseScheduledRateLimitedJob implements ShouldQueue
{
    use Queueable;

    /**
     * 限流键前缀，与 getRateLimitKey() 组合
     */
    private const RATE_LIMIT_KEY_PREFIX = 'job:';

    // 重复检查 TTL, 设为较长时间（如2小时），配合任务执行完毕后清理，避免堆积任务时重复派发
    public const DUPLICATE_CHECK_TTL = 7200;

    public string $duplicateCheckKey = '';
    public string $recentSuccessKey = '';

    abstract protected function getRateLimitKey(): string;

    /**
     * 覆盖 Dispatchable::dispatch()，使所有子类任何调用方式都自动进入限流队列
     * 无论是 Job::dispatch(...) 还是 dispatch(new Job(...))，队列都由此统一设定
     */
    public static function dispatch(...$arguments)
    {
        return (new PendingDispatch(new static(...$arguments)))
            ->onQueue('bilibili-rate-limit');
    }

    /**
     * 子类可覆盖以自定义限流参数，默认从 RateLimitConfig 读取
     */
    protected function getMaxAttempts(): int
    {
        $config = RateLimitConfig::getJobRateLimitConfig($this->getRateLimitKey());
        return $config['max_requests'];
    }

    protected function getDecaySeconds(): int
    {
        $config = RateLimitConfig::getJobRateLimitConfig($this->getRateLimitKey());
        return $config['window_seconds'];
    }

    private static function generateDuplicateCheckKey(array $args): string
    {
        $className = static::class;
        $argsHash  = md5(serialize($args));
        return "job_duplicate_check:{$className}:{$argsHash}";
    }

    private static function generateRecentSuccessKey(array $args): string
    {
        $className = static::class;
        $argsHash  = md5(serialize($args));
        return "job_recent_success:{$className}:{$argsHash}";
    }

    private static function isDuplicateJob(array $args): bool
    {
        return Redis::exists(self::generateDuplicateCheckKey($args));
    }

    private static function markJobAsCreated(array $args): void
    {
        Redis::setex(self::generateDuplicateCheckKey($args), self::DUPLICATE_CHECK_TTL, (string) time());
    }

    public function setJobArgs(array $args): void
    {
        $this->duplicateCheckKey = self::generateDuplicateCheckKey($args);
        $this->recentSuccessKey  = self::generateRecentSuccessKey($args);
    }

    public function clearDuplicateCheck(): void
    {
        if ($this->duplicateCheckKey !== '') {
            Redis::del($this->duplicateCheckKey);
            Log::info('Cleared duplicate check key', ['job' => static::class, 'key' => $this->duplicateCheckKey]);
        }
    }

    public function handle(): void
    {
        $limitKey     = self::RATE_LIMIT_KEY_PREFIX . $this->getRateLimitKey();
        $maxAttempts  = $this->getMaxAttempts();
        $decaySeconds = $this->getDecaySeconds();

        if (RateLimiter::tooManyAttempts($limitKey, $maxAttempts)) {
            $availableIn = RateLimiter::availableIn($limitKey);
            Log::info('Job rate limited', [
                'job'          => static::class,
                'limit_key'    => $limitKey,
                'available_in' => $availableIn,
            ]);
            $this->onRateLimited($availableIn);
            return;
        }

        RateLimiter::hit($limitKey, $decaySeconds);

        try {
            $this->process();
            $this->markRecentSuccessIfEnabled();
            $this->clearDuplicateCheck();
        } catch (\Throwable $e) {
            Log::error('Job processing failed: ' . $e->getMessage(), [
                'job'       => static::class,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    public function failed(\Throwable $exception): void
    {
        $this->clearDuplicateCheck();
        Log::error('Job finally failed after all retries', [
            'job'       => static::class,
            'exception' => $exception->getMessage(),
        ]);
    }

    /**
     * 限流时的处理策略，子类可覆盖。
     * 默认调用 release() 重排，会消耗 attempts（需配合 retryUntil() 使用）。
     * 有 DB 槽位的下载 Job 应覆盖此方法，直接归还槽位后干净退出，不消耗 attempts。
     */
    protected function onRateLimited(int $availableIn): void
    {
        $this->release($availableIn);
    }

    abstract protected function process(): void;

    /**
     * 近 1 小时成功执行过则跳过派发（默认关闭，子类可开启）
     */
    protected static function enableSkipIfSucceededRecently(): bool
    {
        return false;
    }

    protected static function skipIfSucceededRecentlyTtlSeconds(): int
    {
        return 3600;
    }

    private function markRecentSuccessIfEnabled(): void
    {
        if (! static::enableSkipIfSucceededRecently()) {
            return;
        }
        if ($this->recentSuccessKey === '') {
            return;
        }
        $ttl = static::skipIfSucceededRecentlyTtlSeconds();
        if ($ttl <= 0) {
            return;
        }
        Redis::setex($this->recentSuccessKey, $ttl, (string) time());
    }

    /**
     * 创建并派发 Job（立即入队，如果重复则跳过）
     */
    public static function dispatchWithRateLimit(...$args): void
    {
        if (static::enableSkipIfSucceededRecently() && Redis::exists(self::generateRecentSuccessKey($args))) {
            Log::info('Job succeeded recently, skipping dispatch', ['job' => static::class, 'args' => $args]);
            return;
        }

        if (self::isDuplicateJob($args)) {
            Log::info('Duplicate job detected, skipping dispatch', ['job' => static::class, 'args' => $args]);
            return;
        }

        self::markJobAsCreated($args);

        $job = new static(...$args);
        $job->setJobArgs($args);

        dispatch($job)->onQueue('bilibili-rate-limit');
        Log::info('Job dispatched (rate limit applied in handle)', ['job' => static::class]);
    }
}
