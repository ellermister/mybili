<?php

namespace App\Console\Commands;

use App\Jobs\DownloadVideoJob;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Redis;

class DownloadVideo extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:download-video';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'download videos';

    /**
     * Execute the console command.
     */
    public function handle()
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
                $job = new DownloadVideoJob($data);
                dispatch($job);
                // $job->handle();
            }
        }
    }
}
