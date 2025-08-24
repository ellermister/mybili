<?php
namespace App\Jobs;

use App\Services\VideoManager\Actions\Video\PullVideoInfoAction;

class PullVideoInfoJob extends BaseScheduledRateLimitedJob
{
    public $queue = 'default';

    /**
     * Create a new job instance.
     */
    public function __construct(public string $bvid)
    {
        //
    }

    /**
     * 获取限流键名
     */
    protected function getRateLimitKey(): string
    {
        return 'pull_video_info_job';
    }

    /**
     * 获取最大处理数量 - 每3分钟最多5个视频信息更新
     */
    protected function getMaxProcessCount(): int
    {
        return 5;
    }

    /**
     * 获取时间窗口 - 3分钟
     */
    protected function getTimeWindow(): int
    {
        return 180;
    }
    /**
     * Execute the job.
     */
    public function process(): void
    {
        app(PullVideoInfoAction::class)->execute($this->bvid);
    }
}
