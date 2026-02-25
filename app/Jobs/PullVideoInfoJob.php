<?php
namespace App\Jobs;

use App\Services\VideoManager\Actions\Video\PullVideoInfoAction;

class PullVideoInfoJob extends BaseScheduledRateLimitedJob
{
    public function __construct(public string $bvid)
    {
    }

    /**
     * 限流 release() 和异常重试都会消耗 attempts，用时间窗口代替固定次数上限，
     * 避免频繁限流时 attempts 耗尽导致 MaxAttemptsExceededException。
     */
    public function retryUntil(): \DateTimeInterface
    {
        return now()->addHour();
    }

    protected function getRateLimitKey(): string
    {
        return 'update_job';
    }

    public function process(): void
    {
        app(PullVideoInfoAction::class)->execute($this->bvid);
    }
}
