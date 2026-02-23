<?php

namespace App\Jobs;

use App\Services\VideoManager\Actions\Video\FixFavoriteInvalidVideoAction;
use Log;

class FixInvalidFavVideosJob extends BaseScheduledRateLimitedJob
{
    public function __construct(public array $fav, public int $page)
    {
    }

    protected function getRateLimitKey(): string
    {
        return 'update_job';
    }

    /**
     * Execute the job.
     */
    protected function process(): void
    {
        Log::info('Fix invalid fav videos job start');
        app(FixFavoriteInvalidVideoAction::class)->execute($this->fav['id'], $this->page);
        Log::info('Fix invalid fav videos job end', ['fav_title' => $this->fav['title'], 'page' => $this->page]);
    }

    public function displayName(): string
    {
        return __CLASS__ . ' ' . $this->fav['title'] . ' page: ' . $this->page;
    }
}
