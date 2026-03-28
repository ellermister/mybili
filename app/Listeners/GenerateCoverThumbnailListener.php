<?php

namespace App\Listeners;

use App\Events\CoverImageStored;
use App\Jobs\GenerateCoverThumbnailJob;

class GenerateCoverThumbnailListener
{
    public function handle(CoverImageStored $event): void
    {
        GenerateCoverThumbnailJob::dispatch($event->cover->id)->onQueue('fast');
    }
}
