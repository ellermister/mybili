<?php
namespace App\Jobs;

use App\Models\VideoPart;
use App\Services\VideoManager\Actions\Video\DownloadVideoPartFileAction;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

class DownloadVideoJob extends BaseScheduledRateLimitedJob
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
     * 获取限流键名
     */
    protected function getRateLimitKey(): string
    {
        return 'download_video_job';
    }

    /**
     * 获取最大处理数量 - 每分钟最多5个视频下载
     */
    protected function getMaxProcessCount(): int
    {
        return config('services.bilibili.limit_download_video_job', 20);
    }

    /**
     * 获取时间窗口 - 1分钟
     */
    protected function getTimeWindow(): int
    {
        return 60;
    }

    /**
     * Execute the job.
     */
    public function process(): void
    {
        $downloadVideoPartFileAction = app(DownloadVideoPartFileAction::class);
        $downloadVideoPartFileAction->execute($this->videoPart);
    }

    public function displayName(): string
    {
        return sprintf('DownloadVideoJob %s-%s %s', $this->videoPart->video_id, $this->videoPart->page, $this->videoPart->part);
    }
}
