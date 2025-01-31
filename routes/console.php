<?php
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Schedule::call(function () {
    Artisan::call('app:update-fav');
})->withoutOverlapping()
  ->everyFiveMinutes();

Schedule::call(function () {
    Artisan::call('app:download-video');
})->withoutOverlapping()->everyFiveMinutes();

Schedule::call(function () {
    Artisan::call('app:download-danmaku');
})->withoutOverlapping()->everyOddHour();