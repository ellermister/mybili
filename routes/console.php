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

