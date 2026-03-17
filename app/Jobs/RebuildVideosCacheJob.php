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
        } finally {
            // 允许下一轮防抖排期
            Redis::del(self::DEBOUNCE_KEY);
        }
    }
}

