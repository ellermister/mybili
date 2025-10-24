<?php
namespace App\Listeners;

use App\Events\SubscriptionUpdated;
use App\Models\Subscription;
use App\Services\CoverService;
use Log;

class SubscriptionImageDownload
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
    public function handle(SubscriptionUpdated $event): void
    {
        $oldSubscription = $event->oldSubscription;
        $newSubscription = $event->newSubscription;

        $oldCover   = $oldSubscription['cover'] ?? '';
        $newCover   = $newSubscription['cover'] ?? '';
        $resourceId = $newSubscription['id'] ?? '';
        if (! $resourceId) {
            Log::info('Subscription ID is empty, skip download', ['newSubscription' => $newSubscription]);
            return;
        }
        $resource = Subscription::find($resourceId);
        if ($oldCover != $newCover && $newCover != '' && $resource != null) {
            Log::info('Download subscription image', ['cover' => $newCover, 'resourceId' => $resourceId]);
            if ($this->coverService->isCoverable($newCover, $resource)) {
                Log::info('Cover is already coverable, skip download', ['cover' => $newCover, 'resourceId' => $resourceId]);
                return;
            }

            $this->coverService->downloadCoverImageJob($newCover, 'subscription', $resource);
            Log::info('Trigger subscription image download job success', ['cover' => $newCover, 'resourceId' => $resourceId]);
        }
    }
}
