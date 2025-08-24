<?php
namespace App\Console\Commands;

use App\Jobs\FixInvalidFavVideosJob;
use App\Jobs\UpdateFavListJob;
use App\Jobs\UpdateFavVideosJob;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\DownloadFilterService;
use App\Services\VideoManager\Actions\Video\CheckVideoPartFileToDownloadAction;
use App\Services\VideoManager\Actions\Video\UpdateVideoPartsAction;
use App\Services\VideoManager\Contracts\FavoriteServiceInterface;
use Illuminate\Console\Command;
use Illuminate\Database\Eloquent\Collection;
use Log;

class UpdateFav extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:update-fav
    {--update-fav= : 触发收藏夹自身更新}
    {--update-fav-videos= : 触发收藏夹列表中视频更新，发现新视频}
    {--update-fav-videos-page= : 更新指定收藏夹的指定页码}
    {--update-video-parts= : 触发拉取更新现有视频分P资料}
    {--download-video-part= : 触发下载现有视频分P}
    {--fix-invalid-fav-videos= : 修复指定收藏夹的无效视频}
    {--fix-invalid-fav-videos-page= : 修复指定收藏夹的指定页码}
    ';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'update bilibili fav list';

    /**
     * Execute the console command.
     */
    public function handle(
        FavoriteServiceInterface $favoriteService,
        UpdateVideoPartsAction $updateVideoPartsAction,
        CheckVideoPartFileToDownloadAction $checkVideoPartFileToDownloadAction
    ) {
        if ($this->option('update-fav')) {
            $job = new UpdateFavListJob();
            dispatch($job);
        }

        if ($this->option('update-fav-videos')) {
            $page    = $this->option('update-fav-videos-page');
            $favList = $favoriteService->getFavorites();
            foreach ($favList as $item) {
                $this->info('dispatch update fav videos: ' . $item['title'] . ' id: ' . $item['id']);
                $this->dispatchUpdateFavVideosJob($item->toArray(), $page);
            }
        }

        if ($this->option('update-video-parts')) {
            Video::chunk(100, function ($videos) use ($updateVideoPartsAction) {
                foreach ($videos as $video) {
                    $currentFav = $video->favorite;
                    if ($this->shouldExcludeByFavForMultiFav($currentFav)) {
                        $message = sprintf('in update video parts command, exclude fav: %s id: %s', $video['title'], $video['id']);
                        $this->info(
                            $message
                        );
                        Log::info($message, ['favs' => collect($currentFav->pluck('id'))->toArray()]);
                        continue;
                    }
                    $updateVideoPartsAction->execute($video);
                }
            });
        }

        if ($this->option('download-video-part')) {
            VideoPart::chunk(100, function ($videoParts) use ($checkVideoPartFileToDownloadAction) {
                foreach ($videoParts as $videoPart) {
                    if ($this->shouldExcludeByFavForMultiFav($videoPart->video->favorite)) {
                        $message = sprintf('in download video part command, exclude fav: %s id: %s', $videoPart->video['title'], $videoPart->video['id']);
                        $this->info($message);
                        Log::info($message, ['favs' => collect($videoPart->video->favorite->pluck('id'))->toArray()]);
                        continue;
                    }
                    $checkVideoPartFileToDownloadAction->execute($videoPart, true);
                }
            });
        }

        if ($this->option('fix-invalid-fav-videos')) {
            $page    = $this->option('fix-invalid-fav-videos-page');
            $favList = $favoriteService->getFavorites();
            foreach ($favList as $item) {
                $this->info('dispatch fix invalid fav videos: ' . $item['title'] . ' id: ' . $item['id']);
                $this->dispatchFixInvalidFavVideosJob($item->toArray(), $page);
            }
        }
    }

    private function shouldExcludeByFavForMultiFav(Collection $favs)
    {
        $downloadFilterService = app(DownloadFilterService::class);

        $favIds = $favs->pluck('id')->unique();
        foreach ($favIds as $favId) {
            if (! $downloadFilterService->shouldExcludeByFav($favId)) {
                return false;
            }
        }
        return true;
    }

    protected function dispatchUpdateFavVideosJob(array $fav, ?int $page = null)
    {

        $downloadFilterService = app(DownloadFilterService::class);
        if ($downloadFilterService->shouldExcludeByFav($fav['id'])) {
            $message = sprintf('in update fav command, exclude fav: %s id: %s', $fav['title'], $fav['id']);
            $this->info($message);
            Log::info($message, ['favs' => collect($fav['id'])]);
            return;
        }

        $pageSize = intval(config('services.bilibili.fav_videos_page_size'));
        if ($page === null) {
            $maxPage    = ceil($fav['media_count'] / $pageSize);
            $targetPage = 1;
            while ($maxPage > 0 && $targetPage <= $maxPage) {
                UpdateFavVideosJob::dispatchWithRateLimit($fav, $targetPage);
                $targetPage++;
            }
        } else {
            UpdateFavVideosJob::dispatchWithRateLimit($fav, intval($page));
        }
    }

    protected function dispatchFixInvalidFavVideosJob(array $fav, ?int $page = null)
    {
        $pageSize = intval(config('services.bilibili.fav_videos_page_size'));
        if ($page === null) {
            $maxPage    = ceil($fav['media_count'] / $pageSize);
            $targetPage = 1;
            while ($maxPage > 0 && $targetPage <= $maxPage) {
                FixInvalidFavVideosJob::dispatchWithRateLimit($fav, $targetPage);
                $targetPage++;
            }
        } else {
            FixInvalidFavVideosJob::dispatchWithRateLimit($fav, intval($page));
        }
    }

}
