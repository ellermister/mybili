<?php

namespace App\Services;

use Illuminate\Support\Facades\Redis;
use Log;

class BilibiliSuspendService
{
    private const SUSPEND_KEY_PREFIX = 'bilibili:suspend';
    private const DEFAULT_SUSPEND_HOURS = 2;

    /**
     * 设置风控
     * @param int $hours 风控时长（小时），默认2小时
     * @return bool
     */
    public function setSuspend(int $hours = self::DEFAULT_SUSPEND_HOURS): bool
    {
        $key = self::SUSPEND_KEY_PREFIX;
        $expireTime = now()->addHours($hours)->timestamp;

        return Redis::setex($key, $hours * 3600, $expireTime);
    }

    /**
     * 获取风控结束时间
     * @return int|null 返回风控结束时间戳，如果未设置则返回null
     */
    public function getSuspendEndTime(): ?int
    {
        $key = self::SUSPEND_KEY_PREFIX;
        $endTime = Redis::get($key);
        Log::info('get suspend end time', ['end_time' => $endTime]);
        return $endTime ? (int) $endTime : null;
    }
}
