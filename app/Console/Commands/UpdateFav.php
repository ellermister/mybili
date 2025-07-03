<?php
namespace App\Console\Commands;

use App\Contracts\VideoDownloadServiceInterface;
use App\Contracts\VideoManagerServiceInterface;
use App\Jobs\UpdateFavListJob;
use App\Jobs\UpdateFavVideosJob;
use App\Models\Video;
use App\Models\VideoPart;
use Illuminate\Console\Command;

class UpdateFav extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:update-fav 
    {--update-fav} 
    {--update-fav-videos} 
    {--update-video-parts} 
    {--download-video-part}
    {--update-fav-videos-page= : 更新指定收藏夹的指定页码}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'update bilibili fav list';

    /**
     * Execute the console command.
     */
    public function handle(VideoManagerServiceInterface $videoManagerService, VideoDownloadServiceInterface $videoDownloadService)
    {
        if ($this->option('update-fav')) {
            $job = new UpdateFavListJob();
            dispatch($job);
        }

        if ($this->option('update-fav-videos')) {
            $page = $this->option('update-fav-videos-page');
            $favList = $videoManagerService->getFavList();
            foreach ($favList as $item) {
                $this->info('dispatch update fav videos: ' . $item['title'] . ' id: ' . $item['id']);
                $this->dispatchUpdateFavVideosJob($item, $page);
            }
        }

        if ($this->option('update-video-parts')) {
            Video::chunk(100, function ($videos) use ($videoManagerService) {
                foreach ($videos as $video) {
                    $videoManagerService->updateVideoParts($video);
                }
            });
        }

        if ($this->option('download-video-part')) {
            VideoPart::chunk(100, function ($videoParts) use ($videoDownloadService) {
                foreach ($videoParts as $videoPart) {
                    $videoDownloadService->downloadVideoPartFile($videoPart, true);
                }
            });
        }
    }

    protected function dispatchUpdateFavVideosJob(array $fav, ?int $page = null){
        $pageSize = intval(config('services.bilibili.fav_videos_page_size'));
        if($page === null){
            $maxPage = ceil($fav['media_count'] / $pageSize);
            $targetPage = 1;
            while($targetPage <= $maxPage){
                UpdateFavVideosJob::dispatchWithRateLimit($fav, $targetPage);
                $targetPage++;
            }
        }else{
            UpdateFavVideosJob::dispatchWithRateLimit($fav, intval($page));
        }
    }

}
