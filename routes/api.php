<?php

use App\Http\Controllers\CookieController;
use App\Http\Controllers\FavController;
use App\Http\Controllers\SettingsController;
use App\Http\Controllers\SubscriptionController;
use App\Http\Controllers\VideoController;
use Illuminate\Support\Facades\Route;
use App\Http\Controllers\SystemController;

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
Route::post('/settings/test-telegram', [SettingsController::class, 'testTelegramConnection']);

Route::apiResource('/subscription', SubscriptionController::class)->only(['index', 'store', 'update', 'destroy', 'show']);

// 显示系统校准信息
Route::get('/system/info', [SystemController::class, 'getSystemInfo']);