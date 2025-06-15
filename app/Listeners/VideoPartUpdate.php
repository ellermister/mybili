<?php

namespace App\Listeners;

use App\Contracts\VideoManagerServiceInterface;
use App\Events\VideoUpdated;
use App\Models\Video;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

class VideoPartUpdate implements ShouldQueue
{
    use Queueable;

    /**
     * Create the event listener.
     */
    public function __construct(public VideoManagerServiceInterface $videoManagerService)
    {
        
    }

    /**
     * Handle the event.
     */
    public function handle(VideoUpdated $event): void
    {
        if(isset($event->newVideo)){
            $video = Video::where('id', $event->newVideo['id'])->first();
            if($video){
                $this->videoManagerService->updateVideoParts($video);
            }
        }
    }
}
