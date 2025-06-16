<?php
namespace App\Listeners;

use App\Contracts\VideoDownloadServiceInterface;
use App\Events\VideoPartUpdated;
use App\Models\VideoPart;
use Illuminate\Contracts\Queue\ShouldQueue;

class VideoPartFileDownload implements ShouldQueue
{

    public $queue = 'slow';
    /**
     * Create the event listener.
     */
    public function __construct(public VideoDownloadServiceInterface $videoDownloadService)
    {
        //
    }

    /**
     * Handle the event.
     */
    public function handle(VideoPartUpdated $event): void
    {
        if (isset($event->newVideoPart)) {
            $videoPart = VideoPart::where('cid', $event->newVideoPart['cid'])->first();
            if ($videoPart) {
                $this->videoDownloadService->downloadVideoPartFile($videoPart, true);
            }
        }
    }
}
