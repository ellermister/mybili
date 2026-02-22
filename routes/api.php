<?php

use App\Http\Controllers\CookieController;
use App\Http\Controllers\DownloadQueueController;
use App\Http\Controllers\FavController;
use App\Http\Controllers\SettingsController;
use App\Http\Controllers\SubscriptionController;
use App\Http\Controllers\VideoController;
use Illuminate\Support\Facades\Route;
use App\Http\Controllers\SystemController;

Route::apiResource('/fav', FavController::class)->only(['show', 'index']);
Route::get('/videos/{id}', [VideoController::class, 'show']);
Route::get('/videos', [VideoController::class, 'index']);
Route::delete('/videos/{id}', [VideoController::class, 'destroy']);
Route::get('/danmaku', [VideoController::class, 'danmaku']);
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

// 下载队列管理
Route::get('/download-queue', [DownloadQueueController::class, 'index']);
Route::get('/download-queue/stat', [DownloadQueueController::class, 'stat']);
Route::post('/download-queue/{id}/cancel', [DownloadQueueController::class, 'cancel']);
Route::post('/download-queue/{id}/retry', [DownloadQueueController::class, 'retry']);
Route::post('/download-queue/{id}/priority', [DownloadQueueController::class, 'priority']);