<?php
namespace App\Jobs;

use App\Contracts\VideoManagerServiceInterface;
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

        $videoManagerService = app(VideoManagerServiceInterface::class);

        $videoManagerService->updateFavList();

        Log::info('Update fav list job end');
    }

}
