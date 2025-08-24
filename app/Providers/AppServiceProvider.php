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
use Illuminate\Support\Facades\DB;

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
        // SQLite WAL模式优化
        if (config('database.default') === 'sqlite') {
            try {
                // 检查数据库文件是否存在
                $databasePath = config('database.connections.sqlite.database');
                if (file_exists($databasePath)) {
                    DB::statement('PRAGMA journal_mode=WAL');
                    DB::statement('PRAGMA synchronous=NORMAL');
                    DB::statement('PRAGMA wal_autocheckpoint=1000');
                    DB::statement('PRAGMA busy_timeout=30000');
                    DB::statement('PRAGMA cache_size=10000');
                    DB::statement('PRAGMA temp_store=MEMORY');
                }
            } catch (\Exception $e) {
                // 在构建阶段或数据库不可用时静默处理
                // 可以记录日志但不抛出异常
                logger()->debug('SQLite optimization skipped: ' . $e->getMessage());
            }
        }
    }
}
