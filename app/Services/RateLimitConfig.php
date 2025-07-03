<?php
namespace App\Services;

class RateLimitConfig
{

    /**
     * Job频率限制配置
     */
    public static function getJobConfig(): array
    {
        // max_requests 最大请求数
        // window_seconds 时间窗口（秒）
        // max_wait_seconds 最大等待时间（秒）
        return [
            'update_fav_videos_job' => [
                'max_requests'     => 12,
                'window_seconds'   => 60,
                'max_wait_seconds' => 600,
            ],
            'update_fav_list_job'   => [
                'max_requests'     => 12,
                'window_seconds'   => 60,
                'max_wait_seconds' => 600,
            ],
        ];
    }

    /**
     * 获取Job配置
     */
    public static function getJobRateLimitConfig(string $jobType): array
    {
        $config = self::getJobConfig();
        return $config[$jobType] ?? [
            'max_requests'     => 5,
            'window_seconds'   => 60,
            'max_wait_seconds' => 300,
        ];
    }
}
