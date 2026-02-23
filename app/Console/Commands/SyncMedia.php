<?php

namespace App\Console\Commands;

use App\Jobs\FixInvalidFavVideosJob;
use App\Jobs\UpdateFavListJob;
use App\Jobs\UpdateFavVideosJob;
use App\Models\Subscription;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\DownloadFilterService;
use App\Services\SubscriptionService;
use App\Services\VideoManager\Actions\Video\CheckVideoPartFileToDownloadAction;
use App\Services\VideoManager\Actions\Video\UpdateVideoPartsAction;
use App\Services\VideoManager\Contracts\FavoriteServiceInterface;
use Illuminate\Console\Command;
use Illuminate\Database\Eloquent\Collection;
use Log;

class SyncMedia extends Command
{
    protected $signature = 'app:sync-media
        {--fav-list : 更新收藏夹列表}
        {--fav-videos : 更新收藏夹内视频（发现新视频）}
        {--fav= : 仅处理指定收藏夹 ID，不传则处理全部}
        {--fav-page= : 仅更新指定收藏夹的页码，不传则全量}
        {--subscriptions : 更新订阅}
        {--subscription= : 仅更新指定订阅 ID，不传则按策略更新全部}
        {--pull-all : 更新订阅时，拉取全部视频}
        {--video-parts : 更新现有视频分P 资料}
        {--video-id= : 仅处理指定视频 ID，不传则处理全部}
        {--download : 触发下载检查并排队下载}
        {--fix-invalid : 修复收藏夹无效视频}
        ';

    protected $description = '统一调度：更新收藏夹、更新订阅、触发下载（支持按收藏夹/订阅/视频过滤）';

    public function handle(
        FavoriteServiceInterface $favoriteService,
        UpdateVideoPartsAction $updateVideoPartsAction,
        CheckVideoPartFileToDownloadAction $checkVideoPartFileToDownloadAction
    ): int {
        $did = false;

        if ($this->option('fav-list')) {
            $this->runFavList();
            $did = true;
        }

        if ($this->option('fav-videos')) {
            $this->runFavVideos($favoriteService);
            $did = true;
        }

        if ($this->option('subscriptions')) {
            $this->runSubscriptions();
            $did = true;
        }

        if ($this->option('video-parts')) {
            $this->runVideoParts($updateVideoPartsAction);
            $did = true;
        }

        if ($this->option('download')) {
            $this->runDownload($checkVideoPartFileToDownloadAction);
            $did = true;
        }

        if ($this->option('fix-invalid')) {
            $this->runFixInvalid($favoriteService);
            $did = true;
        }

        if (! $did) {
            $this->warn('请至少指定一个动作，例如: --fav-list, --fav-videos, --subscriptions, --video-parts, --download, --fix-invalid');
            $this->line('示例: php artisan app:sync-media --fav-videos --fav=123456');
            $this->line('       php artisan app:sync-media --subscriptions --subscription=1');
            $this->line('       php artisan app:sync-media --download --video-id=100');
            return self::FAILURE;
        }

        return self::SUCCESS;
    }

    protected function runFavList(): void
    {
        $this->info('派发: 更新收藏夹列表');
        dispatch(new UpdateFavListJob());
    }

    protected function runFavVideos(FavoriteServiceInterface $favoriteService): void
    {
        $favId = $this->option('fav') ? (int) $this->option('fav') : null;
        $page  = $this->option('fav-page') !== null ? (int) $this->option('fav-page') : null;

        $favList = $favoriteService->getFavorites();
        if ($favId !== null) {
            $favList = $favList->where('id', $favId);
            if ($favList->isEmpty()) {
                $this->warn("未找到收藏夹 id: {$favId}");
                return;
            }
        }

        foreach ($favList as $item) {
            $this->info('派发更新收藏夹视频: ' . $item->title . ' id: ' . $item->id);
            $this->dispatchUpdateFavVideosJob($item->toArray(), $page);
        }
    }

    protected function runSubscriptions(): void
    {
        $subId = $this->option('subscription') ? (int) $this->option('subscription') : null;
        $pullAll = $this->option('pull-all') ? true : false;
        $svc   = app(SubscriptionService::class);

        if ($subId !== null) {
            $sub = Subscription::find($subId);
            if (! $sub) {
                $this->warn("未找到订阅 id: {$subId}");
                return;
            }
            $this->info('更新订阅: ' . $sub->name . ' id: ' . $sub->id);
            $svc->updateSubscription($sub, $pullAll);
        } else {
            $this->info('更新全部订阅（按策略）');
            $svc->updateSubscriptions($pullAll);
        }
    }

    protected function runVideoParts(UpdateVideoPartsAction $updateVideoPartsAction): void
    {
        $videoId = $this->option('video-id') ? (int) $this->option('video-id') : null;
        $builder = Video::query();
        if ($videoId !== null) {
            $builder->where('id', $videoId);
        }

        $count = 0;
        $start = microtime(true);
        $builder->chunk(100, function ($videos) use ($updateVideoPartsAction, &$count) {
            foreach ($videos as $video) {
                $video->load('favorite', 'subscriptions');
                if ($this->shouldExcludeByFavForMultiFav($video->favorite)) {
                    $msg = sprintf('跳过(过滤): %s id: %s', $video->title, $video->id);
                    $this->line($msg);
                    Log::info($msg, ['favs' => $video->favorite->pluck('id')->toArray()]);
                    continue;
                }
                $this->line(sprintf('更新分P: %s id: %s', $video->title, $video->id));
                $updateVideoPartsAction->execute($video);
                $count++;
            }
        });
        $this->info(sprintf('更新视频分P 完成: %d 条, 耗时 %.2f 秒', $count, microtime(true) - $start));
    }

    protected function runDownload(CheckVideoPartFileToDownloadAction $checkVideoPartFileToDownloadAction): void
    {
        $videoId = $this->option('video-id') ? (int) $this->option('video-id') : null;

        $builder = VideoPart::query();
        if ($videoId !== null) {
            $builder->where('video_id', $videoId);
        }

        $count = 0;
        $builder->chunk(100, function ($videoParts) use ($checkVideoPartFileToDownloadAction, &$count) {
            foreach ($videoParts as $videoPart) {
                $checkVideoPartFileToDownloadAction->execute($videoPart, true);
                $count++;
            }
        });
        $this->info("下载检查完成，已处理 {$count} 个分P");
    }

    protected function runFixInvalid(FavoriteServiceInterface $favoriteService): void
    {
        $favId = $this->option('fav') ? (int) $this->option('fav') : null;
        $page  = $this->option('fav-page') !== null ? (int) $this->option('fav-page') : null;

        $favList = $favoriteService->getFavorites();
        if ($favId !== null) {
            $favList = $favList->where('id', $favId);
            if ($favList->isEmpty()) {
                $this->warn("未找到收藏夹 id: {$favId}");
                return;
            }
        }

        foreach ($favList as $item) {
            $this->info('派发修复无效视频: ' . $item->title . ' id: ' . $item->id);
            $this->dispatchFixInvalidFavVideosJob($item->toArray(), $page);
        }
    }

    protected function shouldExcludeByFavForMultiFav(Collection $favs): bool
    {
        $svc   = app(DownloadFilterService::class);
        $ids   = $favs->pluck('id')->unique();
        foreach ($ids as $id) {
            if (! $svc->shouldExcludeByFav($id)) {
                return false;
            }
        }
        return $ids->isNotEmpty();
    }

    protected function dispatchUpdateFavVideosJob(array $fav, ?int $page = null): void
    {
        $svc = app(DownloadFilterService::class);
        if ($svc->shouldExcludeByFav($fav['id'])) {
            $this->line(sprintf('跳过(过滤): %s id: %s', $fav['title'], $fav['id']));
            Log::info('sync-media exclude fav', ['fav_id' => $fav['id'], 'title' => $fav['title']]);
            return;
        }

        $pageSize  = (int) config('services.bilibili.fav_videos_page_size');
        $maxPage   = (int) ceil($fav['media_count'] / $pageSize);
        if ($page !== null) {
            UpdateFavVideosJob::dispatchWithRateLimit($fav, $page);
            return;
        }
        for ($p = 1; $p <= $maxPage; $p++) {
            UpdateFavVideosJob::dispatchWithRateLimit($fav, $p);
        }
    }

    protected function dispatchFixInvalidFavVideosJob(array $fav, ?int $page = null): void
    {
        $pageSize = (int) config('services.bilibili.fav_videos_page_size');
        $maxPage  = (int) ceil($fav['media_count'] / $pageSize);
        if ($page !== null) {
            FixInvalidFavVideosJob::dispatchWithRateLimit($fav, $page);
            return;
        }
        for ($p = 1; $p <= $maxPage; $p++) {
            FixInvalidFavVideosJob::dispatchWithRateLimit($fav, $p);
        }
    }
}
