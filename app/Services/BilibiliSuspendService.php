<?php

namespace App\Services;

use App\Jobs\BilibiliSuspendResumeJob;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Str;
use Laravel\Horizon\Contracts\HorizonCommandQueue;
use Laravel\Horizon\Contracts\SupervisorRepository;
use Laravel\Horizon\SupervisorCommands\ContinueWorking;
use Laravel\Horizon\SupervisorCommands\Pause;
use Log;

class BilibiliSuspendService
{
    private const SUSPEND_KEY_PREFIX = 'bilibili:suspend';
    private const DEFAULT_SUSPEND_HOURS = 2;

    /**
     * 需要被暂停/恢复的 supervisor config key（与 horizon.php 中配置的 key 一致）
     */
    private const RATE_LIMIT_SUPERVISOR = 'supervisor-bilibili-rate-limit';

    /**
     * 设置风控：写入 Redis key + 通过 HorizonCommandQueue 暂停限流 supervisor
     * + 派发延迟恢复 Job 到 default 队列（由非限流 supervisor 执行，确保恢复能跑到）
     *
     * @param int $hours 风控时长（小时），默认 2 小时
     * @return bool
     */
    public function setSuspend(int $hours = self::DEFAULT_SUSPEND_HOURS): bool
    {
        $key        = self::SUSPEND_KEY_PREFIX;
        $expireTime = now()->addHours($hours)->timestamp;

        $ok = Redis::setex($key, $hours * 3600, (string) $expireTime);

        $this->pauseRateLimitSupervisor();

        BilibiliSuspendResumeJob::dispatch()
            ->delay(now()->addHours($hours))
            ->onQueue('default');

        return $ok;
    }

    /**
     * 通过 HorizonCommandQueue 向限流 supervisor 推送 Pause 指令
     * Supervisor 每秒 loop() 时会读取并调用 pause()，worker 停止拉取新 Job
     */
    public function pauseRateLimitSupervisor(): void
    {
        $name = $this->findSupervisorFullName();
        if ($name === null) {
            Log::warning('BilibiliSuspendService: rate-limit supervisor not found, skip pause');
            return;
        }

        app(HorizonCommandQueue::class)->push($name, Pause::class);
        Log::info('BilibiliSuspendService: pushed Pause command to supervisor', ['supervisor' => $name]);
    }

    /**
     * 通过 HorizonCommandQueue 向限流 supervisor 推送 ContinueWorking 指令
     */
    public function continueRateLimitSupervisor(): void
    {
        $name = $this->findSupervisorFullName();
        if ($name === null) {
            Log::warning('BilibiliSuspendService: rate-limit supervisor not found, skip continue');
            return;
        }

        app(HorizonCommandQueue::class)->push($name, ContinueWorking::class);
        Log::info('BilibiliSuspendService: pushed ContinueWorking command to supervisor', ['supervisor' => $name]);
    }

    /**
     * 从 SupervisorRepository 中找到限流 supervisor 的完整名称（含 master 随机后缀）
     * supervisor 名称格式为 "{master_name}:{config_key}"，例如 "myhost-abc1:supervisor-bilibili-rate-limit"
     */
    private function findSupervisorFullName(): ?string
    {
        $supervisors = app(SupervisorRepository::class)->all();

        $target = collect($supervisors)->first(
            fn($s) => Str::endsWith($s->name, self::RATE_LIMIT_SUPERVISOR)
        );

        return $target?->name;
    }

    /**
     * 获取风控结束时间戳
     * @return int|null 风控结束时间戳，未设置则返回 null
     */
    public function getSuspendEndTime(): ?int
    {
        $endTime = Redis::get(self::SUSPEND_KEY_PREFIX);
        return $endTime ? (int) $endTime : null;
    }
}
