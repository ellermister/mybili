<?php
namespace App\Jobs;

use App\Contracts\VideoManagerServiceInterface;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Log;

class UpdateFavVideosJob implements ShouldQueue
{
    use Queueable;

    /**
     * Create a new job instance.
     */
    public function __construct()
    {
        //
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        Log::info('Update fav videos job start');

        $videoManagerService = app(VideoManagerServiceInterface::class);

        $favList = $videoManagerService->getFavList();
        foreach ($favList as $item) {
            $videoManagerService->updateFavVideos($item['id']);
        }

        Log::info('Update fav videos job end');
    }
}
