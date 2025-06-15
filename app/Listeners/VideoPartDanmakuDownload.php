<?php
namespace App\Listeners;

use App\Contracts\VideoManagerServiceInterface;
use App\Events\VideoPartUpdated;
use App\Models\VideoPart;
use Illuminate\Contracts\Queue\ShouldQueue;

class VideoPartDanmakuDownload implements ShouldQueue
{
    /**
     * 指定队列名称
     */
    public $queue = 'slow';

    /**
     * Create the event listener.
     */
    public function __construct(
        public VideoManagerServiceInterface $videoManagerService
    ) {
    }

    /**
     * Handle the event.
     */
    public function handle(VideoPartUpdated $event): void
    {
        if (isset($event->newVideoPart)) {
            $videoPart = VideoPart::where('cid', $event->newVideoPart['cid'])->first();
            if ($videoPart) {
                //模拟测试
                sleep(10);
                return;
                // $this->videoManagerService->downloadDanmaku($videoPart);
            }
        }
    }
}
