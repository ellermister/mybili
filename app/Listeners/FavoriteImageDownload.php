<?php

namespace App\Listeners;

use App\Events\FavoriteUpdated;
use App\Models\FavoriteList;
use App\Services\CoverService;
use Log;

/**
 * 封面文件落地与缩略图：DownloadCoverImageJob → CoverService::downloadCover 会派发 CoverImageStored，由 GenerateCoverThumbnailListener 队列生成缩略图。
 */
class FavoriteImageDownload
{
    /**
     * Create the event listener.
     */
    public function __construct(public CoverService $coverService) {}

    public function handle(FavoriteUpdated $event): void
    {
        $oldCover = $event->oldFav['cover'] ?? '';
        $newCover = $event->newFav['cover'] ?? '';
        $resourceId = $event->newFav['id'] ?? '';

        if (! $resourceId) {
            Log::info('Favorite ID is empty, skip download', ['newFavorite' => $event->newFav]);

            return;
        }

        $resource = FavoriteList::find($resourceId);
        if ($oldCover != $newCover && $newCover != '' && $resource != null) {
            Log::info('Download fav image', ['cover' => $newCover, 'resourceId' => $resourceId]);
            if ($this->coverService->isCoverable($newCover, $resource)) {
                Log::info('Cover is already coverable, skip download', ['cover' => $newCover, 'resourceId' => $resourceId]);

                return;
            }
            $this->coverService->downloadCoverImageJob($newCover, 'favorite', $resource);
            Log::info('Trigger fav image download job success', ['cover' => $newCover, 'resourceId' => $resourceId]);
        }
    }
}
