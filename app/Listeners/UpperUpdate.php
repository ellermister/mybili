<?php

namespace App\Listeners;

use App\Events\UpperTryUpdated;
use App\Services\UpperService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Queue\InteractsWithQueue;
use Log;

class UpperUpdate
{
    /**
     * Create the event listener.
     */
    public function __construct(public UpperService $upperService)
    {
        //
    }

    /**
     * Handle the event.
     */
    public function handle(UpperTryUpdated $event): void
    {
        try{
            $upper = $event->upper;
            if(isset($upper['mid'])){
                $this->upperService->saveUpperInfo($upper['mid'], $upper['name'] ?? '', $upper['face'] ?? '');
            }
        }catch(\Exception $e){
            Log::error('UpperUpdate failed: ' . $e->getMessage(), ['upper' => $upper]);
        }

    }
}
