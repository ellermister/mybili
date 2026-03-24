<?php

namespace App\Jobs;

use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Redis;
use Laravel\Horizon\Contracts\Silenced;

class RebuildVideosCacheJob implements ShouldQueue, Silenced
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    /**
     * 防抖排期锁 key
     */
    public const DEBOUNCE_KEY = 'video_list:rebuild_scheduled';

    public function handle(VideoServiceInterface $videoService): void
    {
        try {
            $videoService->updateVideosCache();
            $this->invalidateFavVideosCaches();
        } finally {
            // 允许下一轮防抖排期
            Redis::del(self::DEBOUNCE_KEY);
        }
    }

    /**
     * 清除所有 per-fav 和 per-sub 视频列表缓存，下次请求时重新生成
     */
    private function invalidateFavVideosCaches(): void
    {
        foreach (['fav_videos:*', 'sub_videos:*'] as $pattern) {
            $cursor = '0';
            do {
                [$cursor, $keys] = Redis::scan($cursor, ['match' => $pattern, 'count' => 100]);
                if (! empty($keys)) {
                    Redis::del(...$keys);
                }
            } while ($cursor !== '0');
        }
    }
}

