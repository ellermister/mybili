<?php

namespace App\Jobs;

use App\Contracts\VideoManagerServiceInterface;
use App\Models\VideoPart;
use Illuminate\Contracts\Queue\ShouldQueue;

/**
 * 下载弹幕
 */
class DownloadDanmakuJob implements ShouldQueue
{
    public $queue = 'slow';

    /**
     * Create a new job instance.
     */
    public function __construct(public VideoPart $videoPart, public VideoManagerServiceInterface $videoManagerService)
    {
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        $this->videoManagerService->downloadDanmaku($this->videoPart);
    }
}
