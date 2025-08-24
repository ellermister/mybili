<?php

namespace App\Providers;

use App\Contracts\DownloadImageServiceInterface;
use App\Contracts\TelegramBotServiceInterface;
use App\Contracts\VideoDownloadServiceInterface;
use App\Services\DownloadImageService;
use App\Services\TelegramBotService;
use App\Services\VideoDownloadService;
use App\Services\VideoManager\Contracts\DanmakuServiceInterface;
use App\Services\VideoManager\Contracts\FavoriteServiceInterface;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use App\Services\VideoManager\Services\DanmakuService;
use App\Services\VideoManager\Services\FavoriteService;
use App\Services\VideoManager\Services\VideoService;
use Illuminate\Support\ServiceProvider;


class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        $this->app->singleton(DownloadImageServiceInterface::class, DownloadImageService::class);
        $this->app->singleton(VideoDownloadServiceInterface::class, VideoDownloadService::class);
        $this->app->singleton(TelegramBotServiceInterface::class, TelegramBotService::class);
        
        $this->app->singleton(DanmakuServiceInterface::class, DanmakuService::class);
        $this->app->singleton(FavoriteServiceInterface::class, FavoriteService::class);
        $this->app->singleton(VideoServiceInterface::class, VideoService::class);

    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        //
    }
}
