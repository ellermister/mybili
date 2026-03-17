<?php

namespace App\Jobs;

use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Database\Eloquent\Model;
use App\Services\CoverService;
use Illuminate\Support\Facades\Redis;
use Laravel\Horizon\Contracts\Silenced;
/**
 * 新的下载封面任务处理
 */
class DownloadCoverImageJob implements ShouldQueue, Silenced
{

    public $queue = 'fast';
    
    /**
     * 最大重试次数
     */
    public $tries = 3;
    
    /**
     * 任务失败前等待的时间（以秒为单位）
     */
    public $backoff = [60, 300, 600];
    /**
     * Create a new job instance.
     */
    public function __construct(
        private string $url,
        private string $type,
        private Model $model
    )
    {
        //
    }

    /**
     * Execute the job.
     */
    public function handle(CoverService $coverService): void
    {
        if(empty($this->url)){
            return;
        }
        $coverService->downloadCover($this->url, $this->type, $this->model);

        // 防抖触发 progress 页缓存重建，避免封面批量下载时事件风暴
        $ttlSeconds = 10;
        $locked = Redis::setnx(RebuildVideosCacheJob::DEBOUNCE_KEY, 1);
        if (! $locked) {
            return;
        }
        Redis::expire(RebuildVideosCacheJob::DEBOUNCE_KEY, $ttlSeconds);

        RebuildVideosCacheJob::dispatch()
            ->delay(now()->addSeconds(2))
            ->onQueue('fast');
    }
}
