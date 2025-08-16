<?php

namespace App\Listeners;

use App\Events\VideoUpdated;
use App\Jobs\SendVideoFreezeMessageJob;
use Log;

class SendVideoFreezeMessage
{
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
        if(isset($event->oldVideo['freeze']) && isset($event->newVideo['freeze'])){
            if(!$event->oldVideo['freeze'] && $event->newVideo['freeze']){
                Log::info('Video frozen listener', ['video' => $event->newVideo]);

                dispatch(new SendVideoFreezeMessageJob($event->newVideo));
            }
        }
    }
}
