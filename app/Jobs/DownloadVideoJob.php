<?php
namespace App\Jobs;

use App\Models\VideoPart;
use App\Services\DownloadQueueService;
use App\Services\VideoManager\Actions\Video\DownloadVideoPartFileAction;

class DownloadVideoJob extends BaseScheduledRateLimitedJob
{

    public $tries   = 3;
    public $backoff = [1800, 3600, 7200];

    public function __construct(public VideoPart $videoPart)
    {
    }

    protected function getRateLimitKey(): string
    {
        return 'download_job';
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
        if($this->videoPart->video->title && $this->videoPart->video->title != $this->videoPart->part){
            return sprintf('DownloadVideoJob %s-%s %s-%s', $this->videoPart->video_id, $this->videoPart->page, $this->videoPart->video->title, $this->videoPart->part);
        }else{
            return sprintf('DownloadVideoJob %s-%s %s', $this->videoPart->video_id, $this->videoPart->page, $this->videoPart->part);
        }
    }
}
