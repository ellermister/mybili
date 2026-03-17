<?php

namespace App\Jobs;

use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;
use Laravel\Horizon\Contracts\Silenced;

class TriggerProcessDownloadQueueJob implements ShouldQueue, Silenced
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public const LOCK_KEY = 'process_download_queue:trigger';

    public function handle(): void
    {
        $locked = Redis::setnx(self::LOCK_KEY, 1);
        if (! $locked) {
            return;
        }

        Redis::expire(self::LOCK_KEY, 5);

        try {
            Artisan::call('app:process-download-queue');
        } catch (\Throwable $e) {
            Log::warning('Trigger process-download-queue failed', [
                'error' => $e->getMessage(),
            ]);
        } finally {
            Redis::del(self::LOCK_KEY);
        }
    }
}

