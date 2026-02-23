<?php
namespace App\Jobs;

use App\Services\VideoManager\Actions\Video\PullVideoInfoAction;

class PullVideoInfoJob extends BaseScheduledRateLimitedJob
{
    public $backoff = [60, 300, 600];

    public function __construct(public string $bvid)
    {
    }

    protected function getRateLimitKey(): string
    {
        return 'update_job';
    }

    /**
     * Execute the job.
     */
    public function process(): void
    {
        app(PullVideoInfoAction::class)->execute($this->bvid);
    }
}
