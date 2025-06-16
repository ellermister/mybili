<?php
namespace App\Listeners;

use App\Contracts\DownloadImageServiceInterface;
use App\Events\VideoUpdated;
use App\Jobs\DownloadVideoImage;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
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

        if ($newVideo['invalid']) {
            Log::info('Video is invalid, skip download video image', ['id' => $newVideo['id'], 'bvid' => $newVideo['bvid'], 'title' => $newVideo['title']]);
            return;
        }

        // 如果封面有变化，或者封面不为空且缓存封面为空，则下载封面
        if (
            ($oldCover != $newCover) || ($newCover != '' && $newVideo['cache_image'] == '')
        ) {
            Log::info('Download video image', ['cover' => $newVideo['cover']]);
            DownloadVideoImage::dispatch($newVideo, $this->downloadImageService);
        }
    }
}
