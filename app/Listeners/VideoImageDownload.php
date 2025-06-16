<?php

namespace App\Listeners;

use App\Contracts\DownloadImageServiceInterface;
use App\Events\VideoUpdated;
use App\Jobs\DownloadVideoImage;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\InteractsWithQueue;
use Log;

class VideoImageDownload implements ShouldQueue
{
    use Queueable;
    /**
     * Create the event listener.
     */
    public function __construct(public DownloadImageServiceInterface $downloadImageService)
    {
    }

    /**
     * Handle the event.
     */
    public function handle(VideoUpdated $event): void
    {
        $oldVideo = $event->oldVideo;
        $newVideo = $event->newVideo;

        $oldCover = $oldVideo['cover'] ?? '';
        $newCover = $newVideo['cover'] ?? '';

        if ($oldCover != $newCover) {
            Log::info('Download video image', ['cover' => $newVideo['cover']]);
            DownloadVideoImage::dispatch($newVideo, $this->downloadImageService);
        }
    }
}
