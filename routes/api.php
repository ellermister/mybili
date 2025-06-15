<?php

use App\Http\Controllers\CookieController;
use App\Http\Controllers\FavController;
use App\Http\Controllers\SettingsController;
use App\Http\Controllers\VideoController;
use Carbon\Carbon;
use Illuminate\Support\Facades\Route;

Route::apiResource('/fav', FavController::class)->only(['show', 'index']);
Route::get('/video/{id}', [VideoController::class, 'show']);
Route::get('/danmaku/v3', [VideoController::class, 'danmakuV3']);
Route::get('/danmaku/{id}', [VideoController::class, 'danmaku']);
Route::get('/progress', [VideoController::class, 'progress']);
Route::get('/cookie/exist', [CookieController::class, 'checkFileExist']);
Route::get('/cookie/status', [CookieController::class, 'checkCookieValid']);
Route::post('/cookie/upload', [CookieController::class, 'uploadCookieFile']);
Route::get('/settings', [SettingsController::class, 'getSettings']);
Route::post('/settings', [SettingsController::class, 'saveSettings']);

// 显示系统校准信息
Route::get('/system/info', function () {
    $info = [
        'php_version' => phpversion(),
        'laravel_version' => app()->version(),
        'database_version' => DB::connection()->getPdo()->getAttribute(PDO::ATTR_SERVER_VERSION),
        'timezone' => config('app.timezone'),
        'time_now' => Carbon::now()->toDateTimeString(),
    ];
    return response()->json($info);
});