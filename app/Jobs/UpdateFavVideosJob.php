<?php
namespace App\Jobs;

use App\Services\VideoManager\Actions\Favorite\UpdateFavoriteVideosAction;
use Log;

class UpdateFavVideosJob extends BaseScheduledRateLimitedJob
{
    public function __construct(public array $fav, public ?int $page = null)
    {
    }

    /**
     * 限流 release() 和异常重试都会消耗 attempts，用时间窗口代替固定次数上限，
     * 避免频繁限流时 attempts 耗尽导致 MaxAttemptsExceededException。
     */
    public function retryUntil(): \DateTimeInterface
    {
        return now()->addHour();
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
