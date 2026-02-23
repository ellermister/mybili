<?php
namespace App\Jobs;

use App\Models\AudioPart;
use App\Services\DownloadQueueService;
use App\Services\VideoManager\Actions\Audio\DownloadAudioPartFileAction;

class DownloadAudioJob extends BaseScheduledRateLimitedJob
{

    public $tries = 3;
    public $backoff = [1800, 3600, 7200];

    public function __construct(public AudioPart $audioPart)
    {
    }

    protected function getRateLimitKey(): string
    {
        return 'download_job';
    }

    public function process(): void
    {
        app(DownloadAudioPartFileAction::class)->execute($this->audioPart);
        app(DownloadQueueService::class)->markDoneByAudio($this->audioPart->video_id);
    }

    public function failed(\Throwable $exception): void
    {
        parent::failed($exception);
        app(DownloadQueueService::class)->markFailedByAudio(
            $this->audioPart->video_id,
            $exception->getMessage()
        );
    }

    public function displayName(): string
    {
        return sprintf('DownloadAudioJob au%s', $this->audioPart->sid);
    }
}
