<?php

namespace App\Providers;

use App\Contracts\DownloadImageServiceInterface;
use App\Contracts\TelegramBotServiceInterface;
use App\Contracts\VideoDownloadServiceInterface;
use App\Contracts\VideoManagerServiceInterface;
use App\Services\DownloadImageService;
use App\Services\TelegramBotService;
use App\Services\VideoDownloadService;
use App\Services\VideoManagerDBService;
use Illuminate\Support\ServiceProvider;


class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        $this->app->singleton(VideoManagerServiceInterface::class, VideoManagerDBService::class);
        $this->app->singleton(DownloadImageServiceInterface::class, DownloadImageService::class);
        $this->app->singleton(VideoDownloadServiceInterface::class, VideoDownloadService::class);
        $this->app->singleton(TelegramBotServiceInterface::class, TelegramBotService::class);

    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        //
    }
}
