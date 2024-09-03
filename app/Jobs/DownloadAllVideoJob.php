<?php

namespace App\Jobs;

use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

class DownloadAllVideoJob implements ShouldQueue
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
        $iterator = null;
        $keys     = [
        ];
        do {
            $result = redis()->scan($iterator, 'video:*', 50);
            $keys   = array_merge($keys, $result);
        } while ($iterator != 0);

        foreach ($keys as $videoKey) {
            $result = redis()->get($videoKey);
            $data   = json_decode($result, true);
            if ($data) {
                if(!video_has_invalid($data)){
                    $job = new DownloadVideoJob($data);
                    // $job->handle();
                    dispatch($job);
                }
            }
        }
    }
}
