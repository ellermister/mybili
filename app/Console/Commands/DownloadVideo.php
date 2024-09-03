<?php

namespace App\Console\Commands;

use App\Jobs\DownloadAllVideoJob;
use Illuminate\Console\Command;

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
        $job = new DownloadAllVideoJob();
        $job->handle();
    }
}
