<?php
namespace App\Jobs;

use App\Contracts\VideoDownloadServiceInterface;
use App\Models\VideoPart;
use Illuminate\Contracts\Queue\ShouldQueue;

class DownloadVideoJob implements ShouldQueue
{

    public $queue = 'slow';

    /**
     * Create a new job instance.
     */
    public function __construct(public VideoPart $videoPart)
    {
    }

    public $tries = 3;

    /**
     * 任务失败前等待的时间（以秒为单位）
     *
     * @var array
     */
    public $backoff = [1800, 3600, 7200];

    /**
     * Execute the job.
     */
    public function handle(VideoDownloadServiceInterface $videoDownloadService): void
    {
        $videoDownloadService->downloadVideoPartFileQueue($this->videoPart);
    }
}
