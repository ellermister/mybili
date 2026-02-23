<?php
namespace App\Jobs;

use App\Services\VideoManager\Actions\Favorite\UpdateFavoriteVideosAction;
use Log;

class UpdateFavVideosJob extends BaseScheduledRateLimitedJob
{
    public $backoff = [60, 300, 600];

    public function __construct(public array $fav, public ?int $page = null)
    {
    }

    protected function getRateLimitKey(): string
    {
        return 'update_job';
    }

    /**
     * 具体的处理逻辑
     */
    protected function process(): void
    {
        Log::info('Update favorite videos job start');
        app(UpdateFavoriteVideosAction::class)->execute($this->fav, $this->page);
        Log::info('Update favorite videos job end', ['fav_title' => $this->fav['title'], 'page' => $this->page]);
    }

    public function displayName(): string
    {
        return __CLASS__ . ' ' . $this->fav['title'] . ' page: ' . $this->page;
    }
}
