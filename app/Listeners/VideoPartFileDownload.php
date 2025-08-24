<?php
namespace App\Listeners;

use App\Events\VideoPartUpdated;
use App\Models\VideoPart;
use App\Services\VideoManager\Actions\Video\CheckVideoPartFileToDownloadAction;
use Illuminate\Contracts\Queue\ShouldQueue;

class VideoPartFileDownload implements ShouldQueue
{

    public $queue = 'default';
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
    public function handle(VideoPartUpdated $event): void
    {
        if (isset($event->newVideoPart)) {
            $videoPart = VideoPart::where('cid', $event->newVideoPart['cid'])->first();
            if ($videoPart) {
                app(CheckVideoPartFileToDownloadAction::class)->execute($videoPart, true);
            }
        }
    }
}
