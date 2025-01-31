<?php
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Schedule::call(function () {
    Artisan::call('app:update-fav');
})->everyFiveMinutes();

Schedule::call(function () {
    Artisan::call('app:download-video');
})->everyFiveMinutes();

Schedule::call(function () {
    Artisan::call('app:download-danmaku');
})->everyOddHour();