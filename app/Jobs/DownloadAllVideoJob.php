<?php

namespace App\Jobs;

use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

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
                if (!video_has_invalid($data)) {

                    $key = sprintf('download_lock:%s', $data['id']);
                    if (redis()->setnx($key, 1)) {
                        redis()->expire($key, 60 * 60 * 24);
                        $job = new DownloadVideoJob($data);
                        dispatch($job);
                    }
                }
            }
        }
    }
}
