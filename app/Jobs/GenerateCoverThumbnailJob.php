<?php

namespace App\Jobs;

use App\Models\Cover;
use App\Services\CoverThumbnailService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Laravel\Horizon\Contracts\Silenced;

class GenerateCoverThumbnailJob implements ShouldQueue, Silenced
{
    use Dispatchable, InteractsWithQueue, SerializesModels;

    public $queue = 'fast';

    public $tries = 3;

    public $backoff = [30, 120, 300];

    public function __construct(public int $coverId) {}

    public function handle(CoverThumbnailService $thumbnailService): void
    {
        $cover = Cover::find($this->coverId);
        if ($cover === null) {
            return;
        }

        $thumbnailService->generateForCover($cover);
    }
}
