<?php
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Schedule::call(function () {
    Artisan::call('app:update-fav');
})
->name('update-fav')
->withoutOverlapping()
->everyTenMinutes();

Schedule::call(function () {
    Artisan::call('app:download-video');
})
->name('download-video')
->withoutOverlapping()
->everyTenMinutes();

Schedule::call(function () {
    Artisan::call('app:download-danmaku');
})
->name('download-danmaku')
->withoutOverlapping()
->dailyAt('00:00');