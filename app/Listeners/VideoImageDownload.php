<?php
namespace App\Listeners;

use App\Events\VideoUpdated;
use App\Models\Video;
use App\Services\CoverService;
use Log;

class VideoImageDownload
{
    /**
     * Create the event listener.
     */
    public function __construct(public CoverService $coverService)
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

        $resourceId = $newVideo['id'] ?? '';
        if (! $resourceId) {
            Log::info('Video ID is empty, skip download', ['newVideo' => $newVideo]);
            return;
        }
        $resource = Video::find($resourceId);
        if ($oldCover != $newCover && $newCover != '' && $resource != null) {
            Log::info('Download video image', ['cover' => $newCover, 'resourceId' => $resourceId]);
            if ($this->coverService->isCoverable($newCover, $resource)) {
                Log::info('Cover is already coverable, skip download', ['cover' => $newCover, 'resourceId' => $resourceId]);
                return;
            }

            $this->coverService->downloadCoverImageJob($newCover, 'video', $resource);
            Log::info('Trigger video image download job success', ['cover' => $newCover, 'resourceId' => $resourceId]);
        }
    }
}
