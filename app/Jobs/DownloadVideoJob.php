<?php
namespace App\Jobs;

use App\Models\VideoPart;
use App\Services\DownloadQueueService;
use App\Services\VideoManager\Actions\Video\DownloadVideoPartFileAction;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

class DownloadVideoJob extends BaseScheduledRateLimitedJob
{
    use Dispatchable, InteractsWithQueue, SerializesModels;

    public $queue = 'slow';
    public $tries = 3;
    public $backoff = [1800, 3600, 7200];

    public function __construct(public VideoPart $videoPart)
    {
    }

    protected function getRateLimitKey(): string
    {
        return 'download_video_job';
    }

    protected function getMaxProcessCount(): int
    {
        return config('services.bilibili.limit_download_video_job', 20);
    }

    protected function getTimeWindow(): int
    {
        return 60;
    }

    public function process(): void
    {
        app(DownloadVideoPartFileAction::class)->execute($this->videoPart);
        app(DownloadQueueService::class)->markDoneByVideoPart($this->videoPart->id);
    }

    public function failed(\Throwable $exception): void
    {
        parent::failed($exception);
        app(DownloadQueueService::class)->markFailedByVideoPart(
            $this->videoPart->id,
            $exception->getMessage()
        );
    }

    public function displayName(): string
    {
        return sprintf('DownloadVideoJob %s-%s %s', $this->videoPart->video_id, $this->videoPart->page, $this->videoPart->part);
    }
}
