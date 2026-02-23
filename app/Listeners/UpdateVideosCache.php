<?php
namespace App\Listeners;

use App\Events\VideoPartDownloaded;
use App\Events\VideoPartUpdated;
use App\Events\VideoUpdated;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Illuminate\Contracts\Queue\ShouldQueue;
use Laravel\Horizon\Contracts\Silenced;
use Log;
class UpdateVideosCache implements ShouldQueue, Silenced
{
    public $queue = 'fast';

    /**
     * Create the event listener.
     */
    public function __construct()
    {
        //
    }

    /**
     * Handle the event.
     */
    public function handle(VideoUpdated|VideoPartUpdated|VideoPartDownloaded $event): void
    {
        app(VideoServiceInterface::class)->updateVideosCache();
    }
}
