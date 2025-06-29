<?php

use App\Enums\SettingKey;
use App\Services\SettingsService;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Schedule::call(function () {
    $updateFavEnable = app(SettingsService::class)->get(SettingKey::FAVORITE_SYNC_ENABLED);
    if ($updateFavEnable == "on") {
        Artisan::call('app:update-fav', ['--update-fav' => true]);
    }
})
->name('update-fav')
->withoutOverlapping()
->everyTenMinutes();


Schedule::call(function () {
    $updateFavEnable = app(SettingsService::class)->get(SettingKey::FAVORITE_SYNC_ENABLED);
    if ($updateFavEnable == "on") {
        Artisan::call('app:update-fav', ['--update-fav-videos' => true]);
    }
})
->name('update-fav-videos')
->withoutOverlapping()
->everyTenMinutes();


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