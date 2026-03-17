<?php
namespace App\Listeners;

use App\Events\VideoPartDownloaded;
use App\Events\VideoPartUpdated;
use App\Events\VideoUpdated;
use App\Jobs\RebuildVideosCacheJob;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Support\Facades\Redis;
use Laravel\Horizon\Contracts\Silenced;
class UpdateVideosCache implements ShouldQueue, Silenced
{
    public $queue = 'fast';

    /**
     * Create the event listener.
     */
    public function __construct()
    {
        //
    }

    /**
     * Handle the event.
     */
    public function handle(VideoUpdated|VideoPartUpdated|VideoPartDownloaded $event): void
    {
        // 防抖：短时间内多次事件只触发一次重建
        $ttlSeconds = 10;
        $locked = Redis::setnx(RebuildVideosCacheJob::DEBOUNCE_KEY, 1);
        if (! $locked) {
            return;
        }
        Redis::expire(RebuildVideosCacheJob::DEBOUNCE_KEY, $ttlSeconds);

        // 延迟一点点，让事件在短窗口内合并
        RebuildVideosCacheJob::dispatch()
            ->delay(now()->addSeconds(2))
            ->onQueue('fast');
    }
}
