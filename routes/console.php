<?php

use App\Services\VideoManager\Contracts\VideoServiceInterface;
use App\Enums\SettingKey;
use App\Services\SettingsService;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

// 每10分钟更新收藏夹列表
Schedule::call(function () {
    $updateFavEnable = app(SettingsService::class)->get(SettingKey::FAVORITE_SYNC_ENABLED);
    if ($updateFavEnable == "on") {
        Artisan::call('app:sync-media', ['--fav-list' => true]);
    }
})
->name('update-fav')
->withoutOverlapping()
->everyTenMinutes();

// 每10分钟更新收藏夹视频，只更新第一页
Schedule::call(function () {
    $updateFavEnable = app(SettingsService::class)->get(SettingKey::FAVORITE_SYNC_ENABLED);
    if ($updateFavEnable == "on") {
        if (app(VideoServiceInterface::class)->count() == 0) {
            Artisan::call('app:sync-media', ['--fav-videos' => true]);
            return;
        }
        if (now()->format('H') === '04') {
            return;
        }
        Artisan::call('app:sync-media', ['--fav-videos' => true, '--fav-page' => 1]);
    }
})
->name('update-fav-videos-page-1')
->withoutOverlapping()
->everyTenMinutes();

// 每天凌晨全量更新收藏夹视频
Schedule::call(function () {
    $updateFavEnable = app(SettingsService::class)->get(SettingKey::FAVORITE_SYNC_ENABLED);
    if ($updateFavEnable == "on") {
        Artisan::call('app:sync-media', ['--fav-videos' => true]);
    }
})
->name('update-fav-videos-all')
->withoutOverlapping()
->dailyAt('04:00');

// 每天凌晨修复收藏夹无效视频
Schedule::call(function () {
    Artisan::call('app:sync-media', ['--fix-invalid' => true]);
})
->name('fix-invalid-fav-videos')
->withoutOverlapping()
->dailyAt('04:00');


Schedule::command('stats:send')->daily();

// Horizon metrics snapshot - 每分钟收集队列指标数据
Schedule::command('horizon:snapshot')->everyMinute();


Schedule::call(function () {
    $humanReadableNameEnable = app(SettingsService::class)->get(SettingKey::HUMAN_READABLE_NAME_ENABLED);
    if ($humanReadableNameEnable == "on") {
        Artisan::call('app:make-human-readable-names');
    }
})
->name('make-human-readable-names')
->withoutOverlapping()
->everyTenMinutes();


// 每10分钟更新订阅
Schedule::call(function () {
    Artisan::call('app:sync-media', ['--subscriptions' => true]);
})
->name('update-subscription')
->withoutOverlapping()
->everyTenMinutes();


Schedule::command('app:update-no-parts-valid-video')->hourly();

// 每分钟从下载队列取出任务并派发 Job
Schedule::command('app:process-download-queue')
    ->everyMinute()
    ->withoutOverlapping();