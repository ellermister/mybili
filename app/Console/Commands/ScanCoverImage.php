<?php
namespace App\Console\Commands;

use App\Jobs\DownloadCoverImageJob;
use App\Models\FavoriteList;
use App\Models\Subscription;
use App\Models\Upper;
use App\Models\Video;
use Illuminate\Console\Command;

class ScanCoverImage extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:scan-cover-image {--target=} {--id=}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'scan cover image';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $target = $this->option('target');
        $id     = $this->option('id');
        if ($target) {
            if ($target == 'favorite') {
                $builder = FavoriteList::query();
                if ($id) {
                    $builder->where('id', $id);
                }
                $builder->each(function ($favoriteList) {
                    if ($favoriteList->cover) {
                        dispatch(new DownloadCoverImageJob($favoriteList->cover, 'favorite', $favoriteList));
                    } else {
                        $this->warn("Favorite list cover is empty (ID: {$favoriteList->id})");
                    }
                });
            } else if ($target == 'subscription') {
                $builder = Subscription::query();
                if ($id) {
                    $builder->where('id', $id);
                }
                $builder->each(function ($subscription) {
                    if ($subscription->cover) {
                        if ($subscription->type == 'up') {
                            dispatch(new DownloadCoverImageJob($subscription->cover, 'avatar', $subscription));
                        } else {
                            dispatch(new DownloadCoverImageJob($subscription->cover, 'video', $subscription));
                        }
                    } else {
                        $this->warn("Subscription cover is empty (ID: {$subscription->id})");
                    }
                });
            } else if ($target == 'video') {
                $builder = Video::query();
                if ($id) {
                    $builder->where('id', $id);
                }
                $builder->each(function ($video) {
                    if ($video->cover) {
                        dispatch(new DownloadCoverImageJob($video->cover, 'video', $video));
                    } else {
                        $this->warn("Video cover is empty (ID: {$video->id})");
                    }
                });
            } else if ($target == 'upper') {
                $builder = Upper::query();
                if ($id) {
                    $builder->where('mid', $id);
                }
                $builder->each(function ($upper) {
                    if ($upper->face) {
                        dispatch(new DownloadCoverImageJob($upper->face, 'avatar', $upper));
                    } else {
                        $this->warn("Upper cover is empty (MID: {$upper->mid})");
                    }
                });
            } else {
                $this->error('Invalid target');
            }
        } else {
            $this->error('Invalid target');
        }
    }
}
