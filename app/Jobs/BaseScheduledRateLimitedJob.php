<?php

namespace App\Jobs;

use App\Services\ScheduledRateLimiterService;
use App\Services\RateLimitConfig;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;

/**
 * 基于预安排执行时间的频率限制Job基类
 * 在任务触发时就安排好执行时间，避免重试和阻塞
 */
abstract class BaseScheduledRateLimitedJob implements ShouldQueue
{
    use Queueable;

    /**
     * 最大重试次数（通常不需要重试）
     */
    public $tries = 1;

    /**
     * 重复检查的过期时间（秒）
     */
    public const DUPLICATE_CHECK_TTL = 3600; // 1小时

    /**
     * Job的重复检查键，用于清理
     */
    public string $duplicateCheckKey = '';

    /**
     * 获取限流键名
     * 
     * @return string
     */
    abstract protected function getRateLimitKey(): string;

    /**
     * 获取最大处理数量
     * 
     * @return int
     */
    protected function getMaxProcessCount(): int
    {
        $config = RateLimitConfig::getJobRateLimitConfig($this->getRateLimitKey());
        return $config['max_requests'];
    }

    /**
     * 获取时间窗口（秒）
     * 
     * @return int
     */
    protected function getTimeWindow(): int
    {
        $config = RateLimitConfig::getJobRateLimitConfig($this->getRateLimitKey());
        return $config['window_seconds'];
    }

    /**
     * 生成重复检查的唯一键
     * 
     * @param array $args Job构造函数参数
     * @return string
     */
    private static function generateDuplicateCheckKey(array $args): string
    {
        $className = static::class;
        $argsHash = md5(serialize($args));
        return "job_duplicate_check:{$className}:{$argsHash}";
    }

    /**
     * 检查是否已存在相同的Job
     * 
     * @param array $args Job构造函数参数
     * @return bool
     */
    private static function isDuplicateJob(array $args): bool
    {
        $key = self::generateDuplicateCheckKey($args);
        return Redis::exists($key);
    }

    /**
     * 标记Job为已创建
     * 
     * @param array $args Job构造函数参数
     * @return void
     */
    private static function markJobAsCreated(array $args): void
    {
        $key = self::generateDuplicateCheckKey($args);
        Redis::setex($key, self::DUPLICATE_CHECK_TTL, time());
    }

    /**
     * 获取延迟执行时间（秒）
     * 
     * @return int
     */
    public function getDelaySeconds()
    {
        $scheduledRateLimiter = app(ScheduledRateLimiterService::class);
        $key = $this->getRateLimitKey();
        $maxCount = $this->getMaxProcessCount();
        $window = $this->getTimeWindow();

        $delaySeconds = $scheduledRateLimiter->scheduleAndGetDelaySeconds($key, $maxCount, $window);

        Log::info("Job delay calculation", [
            'job' => static::class,
            'rate_limit_key' => $key,
            'max_count' => $maxCount,
            'window' => $window,
            'delay_seconds' => $delaySeconds,
            'scheduled_time' => $delaySeconds > 0 ? date('Y-m-d H:i:s', time() + $delaySeconds) : 'immediate'
        ]);

        return $delaySeconds;
    }

    /**
     * 设置Job的重复检查键
     * 
     * @param array $args
     * @return void
     */
    public function setJobArgs(array $args): void
    {
        $this->duplicateCheckKey = self::generateDuplicateCheckKey($args);
    }

    /**
     * 清理重复检查标记
     * 
     * @return void
     */
    public function clearDuplicateCheck(): void
    {
        if (!empty($this->duplicateCheckKey)) {
            Redis::del($this->duplicateCheckKey);
            
            Log::info("Cleared duplicate check key", [
                'job' => static::class,
                'key' => $this->duplicateCheckKey
            ]);
        }
    }

    /**
     * 执行Job
     */
    public function handle(): void
    {
        try {
            $this->process();
            // 任务处理成功，清理重复检查标记
            $this->clearDuplicateCheck();
        } catch (\Exception $e) {
            Log::error("Job processing failed: " . $e->getMessage(), [
                'job' => static::class,
                'exception' => $e
            ]);
            throw $e;
        }
    }

    /**
     * 任务最终失败时的处理（重试次数用完后调用）
     * 
     * @param \Throwable $exception
     * @return void
     */
    public function failed(\Throwable $exception): void
    {
        // 任务最终失败，清理重复检查标记
        $this->clearDuplicateCheck();
        
        Log::error("Job finally failed after all retries", [
            'job' => static::class,
            'exception' => $exception->getMessage()
        ]);
    }


    /**
     * 具体的处理逻辑，由子类实现
     */
    abstract protected function process(): void;

    /**
     * 创建并分发Job，自动处理频率限制
     * 
     * @param mixed ...$args Job构造函数参数
     * @return void
     */
    public static function dispatchWithRateLimit(...$args): void
    {
        // 检查是否已存在相同的Job
        if (self::isDuplicateJob($args)) {
            Log::info("Duplicate job detected, skipping dispatch", [
                'job' => static::class,
                'args' => $args
            ]);
            return;
        }

        // 标记Job为已创建
        self::markJobAsCreated($args);

        $job = new static(...$args);
        
        // 将参数存储到Job实例中，用于后续清理
        $job->setJobArgs($args);
        
        $delaySeconds = $job->getDelaySeconds();
        
        if ($delaySeconds > 0) {
            // 使用延迟执行
            dispatch($job)->delay(now()->addSeconds($delaySeconds));
            Log::info("Job dispatched with delay", [
                'job' => static::class,
                'delay_seconds' => $delaySeconds,
                'scheduled_time' => now()->addSeconds($delaySeconds)->format('Y-m-d H:i:s')
            ]);
        } else {
            // 立即执行
            dispatch($job);
            Log::info("Job dispatched immediately", [
                'job' => static::class
            ]);
        }
    }
} 