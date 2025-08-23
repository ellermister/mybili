<?php

namespace App\Jobs;

use App\Contracts\VideoManagerServiceInterface;
use Log;

class FixInvalidFavVideosJob extends BaseScheduledRateLimitedJob
{

    public $queue = 'default';

    /**
     * Create a new job instance.
     */
    public function __construct(public array $fav, public int $page)
    {
        //
    }

    /**
     * 获取限流键名
     */
    protected function getRateLimitKey(): string
    {
        return 'fix_invalid_fav_videos_job';
    }

    /**
     * 获取最大处理数量 - 每分钟最多2个收藏夹更新
     */
    protected function getMaxProcessCount(): int
    {
        return 2;
    }

    /**
     * 获取时间窗口
     */
    protected function getTimeWindow(): int
    {
        return 60;
    }


    /**
     * Execute the job.
     */
    protected function process(): void
    {
        Log::info('Fix invalid fav videos job start');
        $videoManagerService = app(VideoManagerServiceInterface::class);

        $videoManagerService->fixFavInvalidVideos($this->fav['id'], $this->page);
        Log::info('Fix invalid fav videos job end', ['fav_title' => $this->fav['title'], 'page' => $this->page]);
    }

    public function displayName(): string
    {
        return __CLASS__ . ' ' . $this->fav['title'] . ' page: ' . $this->page;
    }
}
