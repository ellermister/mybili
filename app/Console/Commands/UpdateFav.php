<?php
namespace App\Console\Commands;

use App\Contracts\VideoManagerServiceInterface;
use App\Contracts\VideoDownloadServiceInterface;
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
    protected $signature = 'app:update-fav {--update-fav} {--update-fav-videos} {--update-video-parts} {--download-video-part}';

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
            $job = new UpdateFavVideosJob();
            dispatch($job);
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

}
