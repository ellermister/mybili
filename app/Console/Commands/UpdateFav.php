<?php

namespace App\Console\Commands;

use App\Jobs\UpdateFavListJob;
use Illuminate\Console\Command;

class UpdateFav extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:update-fav';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'update bilibili fav list';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $job = new UpdateFavListJob();
        dispatch($job);
    }

}
