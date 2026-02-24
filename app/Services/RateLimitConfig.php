<?php
namespace App\Services;

class RateLimitConfig
{

    /**
     * Job 频率限制配置（用于 Laravel RateLimiter）
     * max_requests = 时间窗口内最多执行次数，window_seconds = 时间窗口（秒）
     * 窗口时间不易过大，避免爆发性消费
     */
    public static function getJobConfig(): array
    {
        return [
            'update_job'   => [
                'max_requests'   => 6,
                'window_seconds' => 6,
            ],
            'download_job' => [
                'max_requests'   => 2,
                'window_seconds' => 12,
            ],
        ];
    }

    /**
     * 获取 Job 限流配置（download 类可通过 config('services.bilibili.limit_download_video_job') 覆盖）
     */
    public static function getJobRateLimitConfig(string $jobType): array
    {
        $config = self::getJobConfig();
        $base   = $config[$jobType] ?? [
            'max_requests'   => 5,
            'window_seconds' => 60,
        ];

        if (in_array($jobType, ['download_job'], true)) {
            $override = config('services.bilibili.limit_download_video_job');
            if ($override !== null && $override !== '') {
                $base['max_requests'] = (int) $override;
            }
        }

        return $base;
    }
}
