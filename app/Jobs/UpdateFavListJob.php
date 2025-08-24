<?php
namespace App\Jobs;

use App\Services\VideoManager\Actions\Favorite\UpdateFavoritesAction;
use Log;

class UpdateFavListJob extends BaseScheduledRateLimitedJob
{
    public $queue = 'default';

    /**
     * Create a new job instance.
     */
    public function __construct()
    {

    }

    /**
     * 获取限流键名
     */
    protected function getRateLimitKey(): string
    {
        return 'update_fav_list_job';
    }

    /**
     * 获取最大处理数量 - 每分钟最多3个收藏列表更新
     */
    protected function getMaxProcessCount(): int
    {
        return 3;
    }

    /**
     * 获取时间窗口 - 3分钟
     */
    protected function getTimeWindow(): int
    {
        return 180;
    }

    /**
     * 具体的处理逻辑
     */
    protected function process(): void
    {
        Log::info('Update favorites job start');

        app(UpdateFavoritesAction::class)->execute();

        Log::info('Update favorites job end');
    }
}
