<?php
namespace App\Jobs;

use App\Contracts\VideoDownloadServiceInterface;
use App\Models\VideoPart;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

class DownloadVideoJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, SerializesModels;

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
    public function handle(): void
    {
        $videoDownloadService = app(VideoDownloadServiceInterface::class);
        $videoDownloadService->downloadVideoPartFileQueue($this->videoPart);
    }

    public function displayName(): string
    {
        return sprintf('DownloadVideoJob %s-%s %s', $this->videoPart->video_id, $this->videoPart->page, $this->videoPart->part);
    }
}
