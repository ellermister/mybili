<?php

use App\Http\Controllers\CookieController;
use App\Http\Controllers\FavController;
use App\Http\Controllers\VideoController;
use Illuminate\Support\Facades\Route;

Route::apiResource('/fav', FavController::class)->only(['show', 'index']);
Route::get('/video/{id}', [VideoController::class, 'show']);
Route::get('/progress', [VideoController::class, 'progress']);
Route::get('/cookie/exist', [CookieController::class, 'checkFileExist']);
Route::get('/cookie/status', [CookieController::class, 'checkCookieValid']);
Route::post('/cookie/upload', [CookieController::class, 'uploadCookieFile']);
