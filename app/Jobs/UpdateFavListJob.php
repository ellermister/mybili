<?php

namespace App\Jobs;

use App\Console\Commands\DownloadVideo;
use App\Services\BilibiliService;
use App\Services\SettingsService;
use App\Services\VideoManagerService;
use Arr;
use Exception;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Log;

class UpdateFavListJob implements ShouldQueue
{
    use Queueable;

    protected $SESSDATA = '';
    protected $mid      = 0;

    /**
     * Create a new job instance.
     */
    public function __construct()
    {

    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        list($this->SESSDATA, $this->mid) = match_cookie_main();

        Log::info('Update fav list job start');
        $videoManagerService = new VideoManagerService(new SettingsService(), new BilibiliService());

        $videoManagerService->updateFavList();

        Log::info('Update fav list job end');
    }

}
