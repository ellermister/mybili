<?php

namespace App\Listeners;

use App\Events\VideoUpdated;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Illuminate\Contracts\Queue\ShouldQueue;

class UpdateVideosCache implements ShouldQueue
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
    public function handle(VideoUpdated $event): void
    {
        app(VideoServiceInterface::class)->updateVideosCache();
    }
}
