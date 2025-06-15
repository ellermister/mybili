<?php

namespace App\Providers;

use App\Contracts\DownloadImageServiceInterface;
use App\Contracts\VideoManagerServiceInterface;
use App\Services\DownloadImageService;
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

    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        //
    }
}
