<?php
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Schedule::call(function () {
    Artisan::call('app:update-fav');
})->hourly()->onSuccess(function () {
    Artisan::call('app:download-video');
});
