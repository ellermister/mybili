<?php
namespace App\Console\Commands;

use App\Models\Video;
use App\Services\DownloadFilterService;
use App\Services\VideoManager\Actions\Video\UpdateVideoPartsAction;
use Illuminate\Console\Command;
use Illuminate\Database\Eloquent\Collection;
use Log;

class UpdateNoPartsValidVideo extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:update-no-parts-valid-video {--id=} {--force}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = '更新没有分P的视频';

    public function handle(UpdateVideoPartsAction $updateVideoPartsAction)
    {
        $startTime = microtime(true);
        $count     = 0;
        $builder   = Video::query()->leftJoin('video_parts as vp', 'vp.video_id', '=', 'videos.id')
            ->whereNull('vp.id')
            ->where('video_downloaded_num', 0)
            ->where('videos.frozen', 0)
            ->where('videos.invalid', 0)->select('videos.*');
        if ($this->option('id')) {
            $builder->where('videos.id', $this->option('id'));
        }
        $builder->chunk(100, function ($videos) use ($updateVideoPartsAction, &$count) {
            foreach ($videos as $video) {
                if (!$this->option('force') && $this->checkCacheAndSetCache($video)) {
                    continue;
                }

                if ($video->favorite->count() > 0 && $this->shouldExcludeByFavForMultiFav($video->favorite)) {
                    $message = sprintf('in update no parts valid video command, exclude fav: %s id: %s', $video['title'], $video['id']);
                    $this->info($message);
                    Log::info($message, ['favs' => collect($video->favorite->pluck('id'))->toArray()]);
                    continue;
                }
                $updateVideoPartsAction->execute($video);
                $count++;
            }
        });
        $endTime = microtime(true);
        $this->info(sprintf('update no parts valid video time: %s seconds, count: %s', $endTime - $startTime, $count));
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

    private function checkCacheAndSetCache(Video $video)
    {
        $cacheKey = 'update-no-parts-valid-video:' . $video->id;
        if (redis()->exists($cacheKey)) {
            return true;
        }
        redis()->set($cacheKey, true);
        redis()->expire($cacheKey, 86400);
        return false;
    }
}
