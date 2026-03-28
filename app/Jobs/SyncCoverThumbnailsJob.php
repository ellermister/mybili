<?php

namespace App\Jobs;

use App\Models\Cover;
use App\Services\CoverThumbnailService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Laravel\Horizon\Contracts\Silenced;

class SyncCoverThumbnailsJob implements ShouldQueue, Silenced
{
    public $queue = 'fast';

    public function handle(CoverThumbnailService $thumbnailService): void
    {
        $limit = (int) config('cover_thumbnail.batch_limit', 500);
        $thumbnailService->syncMissingThumbnails(max(1, $limit));
        // 剩余未生成缩略图的封面数量
        $remainingCount = Cover::query()->whereNull('thumbnail_generated_at')->count();
        if ($remainingCount > 0) {
            dispatch(new SyncCoverThumbnailsJob());
        }
    }
}
