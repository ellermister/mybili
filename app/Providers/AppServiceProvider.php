<?php

namespace App\Providers;

use App\Contracts\DownloadImageServiceInterface;
use App\Contracts\TelegramBotServiceInterface;
use App\Services\DownloadImageService;
use App\Services\TelegramBotService;
use App\Services\VideoManager\Contracts\DanmakuServiceInterface;
use App\Services\VideoManager\Contracts\FavoriteServiceInterface;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use App\Services\VideoManager\DanmakuService;
use App\Services\VideoManager\FavoriteService;
use App\Services\VideoManager\VideoService;
use Illuminate\Support\ServiceProvider;


class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        $this->app->singleton(DownloadImageServiceInterface::class, DownloadImageService::class);
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
