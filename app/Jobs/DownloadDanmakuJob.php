<?php

namespace App\Jobs;

use App\Services\VideoManagerService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

class DownloadDanmakuJob implements ShouldQueue
{
    use Queueable;

    /**
     * Create a new job instance.
     */
    public function __construct(public int $avId)
    {
        //
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        $videoManagerService = app(VideoManagerService::class);
        $videoManagerService->downloadDanmaku($this->avId);
    }
}
