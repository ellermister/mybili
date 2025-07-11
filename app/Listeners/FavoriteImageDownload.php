<?php
namespace App\Listeners;

use App\Contracts\DownloadImageServiceInterface;
use App\Events\FavoriteUpdated;
use App\Models\FavoriteList;
use Log;

class FavoriteImageDownload
{
    /**
     * Create the event listener.
     */
    public function __construct(public DownloadImageServiceInterface $downloadImageService)
    {
    }

    /**
     * Handle the event.
     */
    public function handle(FavoriteUpdated $event): void
    {
        $oldFav = $event->oldFav;
        $newFav = $event->newFav;

        $oldCover = $oldFav['cover'] ?? '';
        $newCover = $newFav['cover'] ?? '';
        if ($oldCover != $newCover ) {
            Log::info('Download fav image', ['cover' => $newFav['cover']]);
            if (empty($newCover)) {
                Log::info('Cover is empty, skip download');
                return;
            }
            try {
                $savePath = $this->downloadImageService->getImageLocalPath($newFav['cover']);
                $this->downloadImageService->downloadImage($newFav['cover'], $savePath);
                FavoriteList::where('id', $newFav['id'])->update(['cache_image' => get_relative_path($savePath)]);
            } catch (\Exception $e) {
                Log::error('Download fav image failed', ['error' => $e->getMessage()]);
            }
        }
    }
}
