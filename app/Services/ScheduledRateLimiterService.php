<?php
namespace App\Services;

use Carbon\Carbon;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;

/**
 * 预安排执行时间的频率限制器
 * 在任务触发时就安排好执行时间，避免重试
 */
class ScheduledRateLimiterService
{
    private const SCHEDULE_PREFIX = 'schedule:';
    private const AVAILABLE_SLOTS_PREFIX = 'available_slots:'; // 存储可用的空隙时间

    public function __construct(
        private BilibiliSuspendService $bilibiliSuspendService
    ) {}

    /**
     * 获取基准时间（考虑风控结束时间）
     */
    private function getBaseTime(): int
    {
        $now = time();
        $suspendEndTime = $this->bilibiliSuspendService->getSuspendEndTime();
        
        // 如果有风控结束时间且在当前时间之后，使用风控结束时间作为基准
        if ($suspendEndTime !== null && $suspendEndTime > $now) {
            return $suspendEndTime;
        }
        
        return $now;
    }

    /**
     * 调度下一个可用的执行时间（会实际锁定时间槽）
     *
     * @param string $key 限流键名
     * @param int $maxRequests 最大请求数
     * @param int $windowSeconds 时间窗口（秒）
     * @return int 下次执行的时间戳
     */
    public function scheduleNextAvailableTime(string $key, int $maxRequests, int $windowSeconds): int
    {
        $scheduleKey = self::SCHEDULE_PREFIX . $key;
        $slotsKey = self::AVAILABLE_SLOTS_PREFIX . $key;
        $now = time();
        $baseTime = $this->getBaseTime();

        Log::info('debug schedule next available time', ['baseTime' => $baseTime, 'date' => date('Y-m-d H:i:s', $baseTime)]);
        try {
            $retry = 0;
            while (true) {
                if (! Redis::setnx('lock:' . $scheduleKey, 1)) {
                    sleep(1);
                    $retry++;
                    if ($retry > 10) {
                        throw new \Exception('Redis lock failed');
                    }
                    continue;
                }

                Redis::expire('lock:' . $scheduleKey, 10);

                // 计算固定间隔
                $interval = $windowSeconds / $maxRequests;

                // 清理过期的执行计划和空隙
                $this->cleanupExpiredTasks($key, $now);

                // 优先检查是否有可用的空隙时间
                $availableSlot = $this->getEarliestAvailableSlot($key, $baseTime);
                if ($availableSlot !== null) {
                    // 有空隙可用，使用空隙时间
                    $nextTime = $availableSlot;
                    // 从可用空隙中移除这个时间
                    Redis::zrem($slotsKey, $availableSlot);
                } else {
                    // 没有空隙，获取最后一个任务的时间并加上间隔
                    $nextTime = $this->calculateNextTimeFromLastTask($key, $interval, $baseTime);
                }

                Log::info('debug schedule next available time', ['nextTime' => $nextTime, 'date' => date('Y-m-d H:i:s', $nextTime)]);

                // 记录这个执行计划
                Redis::zadd($scheduleKey, $nextTime, $nextTime . ':' . uniqid());

                // 设置过期时间
                Redis::expire($scheduleKey, $windowSeconds * 5 + 300);
                Redis::expire($slotsKey, $windowSeconds * 5 + 300);

                Redis::del('lock:' . $scheduleKey);
                return $nextTime;
            }
        } catch (\Exception $e) {
            Log::error("Scheduled rate limiter error: " . $e->getMessage() . ' ' . $e->getTraceAsString());
            return $baseTime + 1;
        }
    }

    /**
     * 获取下一个可用时间（纯查询，不锁定时间槽）
     *
     * @param string $key 限流键名
     * @param int $maxRequests 最大请求数
     * @param int $windowSeconds 时间窗口（秒）
     * @return int 下次执行的时间戳
     */
    public function getNextAvailableTime(string $key, int $maxRequests, int $windowSeconds): int
    {
        $now = time();
        $baseTime = $this->getBaseTime();

        try {
            // 计算固定间隔
            $interval = $windowSeconds / $maxRequests;

            // 清理过期数据
            $this->cleanupExpiredTasks($key, $now);

            // 检查空隙
            $availableSlot = $this->getEarliestAvailableSlot($key, $baseTime);
            if ($availableSlot !== null) {
                return $availableSlot;
            }

            // 没有空隙，计算下一个时间
            return $this->calculateNextTimeFromLastTask($key, $interval, $baseTime);
        } catch (\Exception $e) {
            Log::error("Get next available time error: " . $e->getMessage());
            return $baseTime + 1;
        }
    }

    /**
     * 检查是否可以立即执行
     *
     * @param string $key 限流键名
     * @param int $maxRequests 最大请求数
     * @param int $windowSeconds 时间窗口（秒）
     * @return bool
     */
    public function canExecuteNow(string $key, int $maxRequests, int $windowSeconds): bool
    {
        $nextTime = $this->getNextAvailableTime($key, $maxRequests, $windowSeconds);
        $baseTime = $this->getBaseTime();
        return $nextTime <= $baseTime;
    }

    /**
     * 调度并获取延迟执行时间（会实际锁定时间槽）
     *
     * @param string $key 限流键名
     * @param int $maxRequests 最大请求数
     * @param int $windowSeconds 时间窗口（秒）
     * @return int 需要延迟的秒数
     */
    public function scheduleAndGetDelaySeconds(string $key, int $maxRequests, int $windowSeconds): int
    {
        $nextTime = $this->scheduleNextAvailableTime($key, $maxRequests, $windowSeconds);

        return max(0, $nextTime - time());
    }

    /**
     * 获取延迟执行时间（纯查询，不锁定时间槽）
     *
     * @param string $key 限流键名
     * @param int $maxRequests 最大请求数
     * @param int $windowSeconds 时间窗口（秒）
     * @return int 需要延迟的秒数
     */
    public function getDelaySeconds(string $key, int $maxRequests, int $windowSeconds): int
    {
        $nextTime = $this->getNextAvailableTime($key, $maxRequests, $windowSeconds);

        return max(0, $nextTime - time());
    }

    /**
     * 清理过期的任务和空隙
     */
    private function cleanupExpiredTasks(string $key, int $now): void
    {
        $scheduleKey = self::SCHEDULE_PREFIX . $key;
        $slotsKey = self::AVAILABLE_SLOTS_PREFIX . $key;

        // 清理过期的执行计划
        Redis::zremrangebyscore($scheduleKey, 0, $now - 1);
        
        // 清理过期的空隙
        Redis::zremrangebyscore($slotsKey, 0, $now - 1);
    }

    /**
     * 获取最早的可用空隙时间
     */
    private function getEarliestAvailableSlot(string $key, int $now): ?int
    {
        $slotsKey = self::AVAILABLE_SLOTS_PREFIX . $key;
        
        // 获取最早的可用空隙（大于等于当前时间）
        $slots = Redis::zrangebyscore($slotsKey, $now, '+inf', ['limit' => ['offset' => 0, 'count' => 1]]);
        
        return !empty($slots) ? (int)array_values($slots)[0] : null;
    }

    /**
     * 根据最后一个任务计算下一个执行时间
     */
    private function calculateNextTimeFromLastTask(string $key, float $interval, int $now): int
    {
        $scheduleKey = self::SCHEDULE_PREFIX . $key;
        
        // 获取最后一个任务的时间
        $lastTasks = Redis::zrevrangebyscore($scheduleKey, '+inf', $now, ['limit' => ['offset' => 0, 'count' => 1], 'withscores' => true]);
        
        if (!empty($lastTasks)) {
            // 有任务存在，取最后一个任务时间 + 间隔
            $lastTime = array_values($lastTasks)[0];
            return (int)($lastTime + $interval);
        } else {
            // 没有任务，从当前时间开始
            return $now;
        }
    }

    /**
     * 获取当前计划执行数量
     *
     * @param string $key 限流键名
     * @param int $windowSeconds 时间窗口（秒）
     * @return int
     */
    public function getScheduledCount(string $key, int $windowSeconds): int
    {
        $scheduleKey = self::SCHEDULE_PREFIX . $key;
        $now = time();

        try {
            // 清理过期的执行计划
            Redis::zremrangebyscore($scheduleKey, 0, $now - 1);
            
            // 获取未来的所有计划数量
            return Redis::zcount($scheduleKey, $now, '+inf');
        } catch (\Exception $e) {
            Log::error("Get scheduled count error: " . $e->getMessage());
            return 0;
        }
    }

    /**
     * 获取未来时间窗口内的执行计划
     *
     * @param string $key 限流键名
     * @param int $windowSeconds 时间窗口（秒）
     * @return array 执行时间列表
     */
    public function getScheduledTimes(string $key, int $windowSeconds): array
    {
        $scheduleKey = self::SCHEDULE_PREFIX . $key;
        $now = time();

        try {
            // 清理过期的执行计划
            Redis::zremrangebyscore($scheduleKey, 0, $now - 1);

            // 获取所有未来的计划
            $scheduledTimes = Redis::zrangebyscore($scheduleKey, $now, '+inf', ['withscores' => true]);

            $times = [];
            foreach ($scheduledTimes as $member => $score) {
                $times[] = [
                    'time' => $score,
                    'formatted' => Carbon::createFromTimestamp($score)->format('Y-m-d H:i:s'),
                ];
            }

            return $times;
        } catch (\Exception $e) {
            Log::error("Get scheduled times error: " . $e->getMessage());
            return [];
        }
    }

    /**
     * 取消执行计划（将时间加入可用空隙）
     *
     * @param string $key 限流键名
     * @param int $scheduledTime 计划执行时间
     * @return bool
     */
    public function cancelScheduledExecution(string $key, int $scheduledTime): bool
    {
        $scheduleKey = self::SCHEDULE_PREFIX . $key;
        $slotsKey = self::AVAILABLE_SLOTS_PREFIX . $key;

        try {
            // 移除执行计划
            $removed = Redis::zremrangebyscore($scheduleKey, $scheduledTime, $scheduledTime);
            
            if ($removed > 0) {
                // 如果时间在未来，将其加入可用空隙
                if ($scheduledTime > time()) {
                    Redis::zadd($slotsKey, $scheduledTime, $scheduledTime);
                }
                return true;
            }
            return false;
        } catch (\Exception $e) {
            Log::error("Cancel scheduled execution error: " . $e->getMessage());
            return false;
        }
    }

    /**
     * 重置执行计划
     *
     * @param string $key 限流键名
     * @return bool
     */
    public function resetSchedule(string $key): bool
    {
        $scheduleKey = self::SCHEDULE_PREFIX . $key;

        try {
            Redis::del($scheduleKey);
            return true;
        } catch (\Exception $e) {
            Log::error("Reset schedule error: " . $e->getMessage());
            return false;
        }
    }

    /**
     * 获取执行计划统计信息
     *
     * @param string $key 限流键名
     * @param int $maxRequests 最大请求数
     * @param int $windowSeconds 时间窗口（秒）
     * @return array
     */
    public function getScheduleStats(string $key, int $maxRequests, int $windowSeconds): array
    {
        $now = time();
        $interval = $windowSeconds / $maxRequests;
        
        // 获取当前总的计划数量（未来所有任务）
        $totalScheduledCount = $this->getScheduledCount($key, $windowSeconds);
        
        // 计算下一个可用时间和延迟
        $nextTime = $this->getNextAvailableTime($key, $maxRequests, $windowSeconds);
        $delaySeconds = max(0, $nextTime - $now);
        
        // 计算近期窗口内的任务数量（用于利用率计算）
        $nearTermEnd = $now + $windowSeconds;
        $scheduleKey = self::SCHEDULE_PREFIX . $key;
        $nearTermCount = Redis::zcount($scheduleKey, $now, $nearTermEnd);

        return [
            'key' => $key,
            'scheduled_count' => $nearTermCount, // 显示近期窗口内的数量
            'total_scheduled' => $totalScheduledCount, // 总的未来任务数量
            'max_requests' => $maxRequests,
            'window_seconds' => $windowSeconds,
            'interval_seconds' => (int)$interval,
            'next_available_time' => $nextTime,
            'delay_seconds' => $delaySeconds,
            'can_execute_now' => $delaySeconds === 0,
            'remaining_slots' => max(0, $maxRequests - $nearTermCount),
            'utilization_percent' => round(($nearTermCount / $maxRequests) * 100, 2),
            'base_time' => $now, // 添加基准时间信息
            'suspend_end_time' => $this->bilibiliSuspendService->getSuspendEndTime(), // 添加风控结束时间信息
        ];
    }
}

