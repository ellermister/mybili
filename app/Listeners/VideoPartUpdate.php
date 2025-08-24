<?php
namespace App\Listeners;

use App\Events\VideoUpdated;
use App\Models\Video;
use App\Services\VideoManager\Actions\Video\UpdateVideoPartsAction;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Log;

class VideoPartUpdate implements ShouldQueue
{
    use Queueable;

    /**
     * Create the event listener.
     */
    public function __construct(public UpdateVideoPartsAction $updateVideoPartsAction)
    {

    }

    /**
     * Handle the event.
     */
    public function handle(VideoUpdated $event): void
    {
        if (isset($event->newVideo)) {
            if ($event->newVideo['invalid']) {
                Log::info('Video is invalid, skip update video parts', ['id' => $event->newVideo['id'], 'bvid' => $event->newVideo['bvid'], 'title' => $event->newVideo['title']]);
                return;
            }

            $video = Video::where('id', $event->newVideo['id'])->first();
            if ($video) {
                $this->updateVideoPartsAction->execute($video);
            }
        }
    }
}
