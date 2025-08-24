<?php

namespace App\Jobs;

use App\Models\VideoPart;
use App\Services\VideoManager\Actions\Danmaku\DownloadDanmakuAction;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;

/**
 * 下载弹幕
 */
class DownloadDanmakuJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, SerializesModels;

    public $queue = 'slow';
    
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
    public function __construct(public VideoPart $videoPart)
    {
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        app(DownloadDanmakuAction::class)->execute($this->videoPart);
    }
}
