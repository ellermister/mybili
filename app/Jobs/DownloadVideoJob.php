<?php
namespace App\Jobs;

use App\Models\VideoPart;
use App\Services\DownloadQueueService;
use App\Services\VideoManager\Actions\Video\DownloadVideoPartFileAction;
use Illuminate\Support\Facades\Log;

class DownloadVideoJob extends BaseScheduledRateLimitedJob
{

    public $tries = 1;

    public function __construct(public VideoPart $videoPart)
    {
    }

    protected function getRateLimitKey(): string
    {
        return 'download_job';
    }

    protected function onRateLimited(int $availableIn): void
    {
        app(DownloadQueueService::class)->markPendingByVideoPart($this->videoPart->id);
    }

    public function process(): void
    {
        try{
            app(DownloadVideoPartFileAction::class)->execute($this->videoPart);
        } catch (\App\Exceptions\ApiGetVideoStatusException $e) {
            // 稿件状态异常，跳过下载
            Log::error('video manuscript status abnormal', ['video_id' => $this->videoPart->video_id, 'part' => $this->videoPart->part, 'message' => $e->getMessage(), 'code' => $e->getCode()]);
            // 标记视频失败不下载
            app(DownloadQueueService::class)->markFailedByVideoPart($this->videoPart->id, sprintf('video manuscript status abnormal: %s', $e->getMessage()));
            return;
        }
        app(DownloadQueueService::class)->markDoneByVideoPart($this->videoPart->id);
    }

    public function failed(\Throwable $exception): void
    {
        parent::failed($exception);
        app(DownloadQueueService::class)->markRetryOrFailedByVideoPart(
            $this->videoPart->id,
            $exception->getMessage()
        );
    }

    public function displayName(): string
    {
        if ($this->videoPart->video->title && $this->videoPart->video->title != $this->videoPart->part) {
            return sprintf('DownloadVideoJob %s-%s %s-%s', $this->videoPart->video_id, $this->videoPart->page, $this->videoPart->video->title, $this->videoPart->part);
        } else {
            return sprintf('DownloadVideoJob %s-%s %s', $this->videoPart->video_id, $this->videoPart->page, $this->videoPart->part);
        }
    }
}
