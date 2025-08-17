<?php
namespace App\Listeners;

use App\Contracts\DownloadImageServiceInterface;
use App\Events\SubscriptionUpdated;
use App\Models\Subscription;
use Log;

class SubscriptionImageDownload
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
    public function handle(SubscriptionUpdated $event): void
    {
        $oldSubscription = $event->oldSubscription;
        $newSubscription = $event->newSubscription;

        $oldCover = $oldSubscription['cover'] ?? '';
        $newCover = $newSubscription['cover'] ?? '';
        if ($oldCover != $newCover  || ($newCover != '' && $newSubscription['cache_image'] == '')) {
            Log::info('Download subscription image', ['cover' => $newSubscription['cover']]);
            if (empty($newCover)) {
                Log::info('Cover is empty, skip download');
                return;
            }
            try {
                $savePath = $this->downloadImageService->getImageLocalPath($newSubscription['cover']);
                $this->downloadImageService->downloadImage($newSubscription['cover'], $savePath);
                Subscription::where('id', $newSubscription['id'])->update(['cache_image' => get_relative_path($savePath)]);
            } catch (\Exception $e) {
                Log::error('Download fav image failed', ['error' => $e->getMessage()]);
            }
        }
    }
}
