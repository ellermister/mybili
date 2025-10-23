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
                    // WAL模式配置
                    DB::statement('PRAGMA journal_mode=WAL');
                    DB::statement('PRAGMA synchronous=NORMAL');
                    
                    // 锁定和并发优化
                    DB::statement('PRAGMA busy_timeout=60000'); // 60秒超时
                    DB::statement('PRAGMA locking_mode=NORMAL'); // 允许多连接
                    
                    // WAL checkpoint优化
                    DB::statement('PRAGMA wal_autocheckpoint=2000'); // 增加checkpoint阈值
                    DB::statement('PRAGMA journal_size_limit=67108864'); // 64MB WAL大小限制
                    
                    // 内存和缓存优化
                    DB::statement('PRAGMA cache_size=-20000'); // 20MB缓存
                    DB::statement('PRAGMA temp_store=MEMORY'); // 临时表存内存
                    DB::statement('PRAGMA mmap_size=268435456'); // 256MB内存映射
                    
                    // 性能优化
                    DB::statement('PRAGMA page_size=4096'); // 4KB页大小
                    DB::statement('PRAGMA auto_vacuum=INCREMENTAL'); // 增量清理
                    DB::statement('PRAGMA incremental_vacuum(100)'); // 清理100页
                    
                    // 查询优化
                    DB::statement('PRAGMA optimize'); // 优化查询计划
                }
            } catch (\Exception $e) {
                // 在构建阶段或数据库不可用时静默处理
                // 可以记录日志但不抛出异常
                logger()->debug('SQLite optimization skipped: ' . $e->getMessage());
            }
        }
    }
}
