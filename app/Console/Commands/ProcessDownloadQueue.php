<?php
namespace App\Console\Commands;

use App\Jobs\DownloadAudioJob;
use App\Jobs\DownloadVideoJob;
use App\Models\AudioPart;
use App\Models\DownloadQueue;
use App\Models\VideoPart;
use App\Services\BilibiliSuspendService;
use App\Services\DownloadQueueService;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class ProcessDownloadQueue extends Command
{
    protected $signature   = 'app:process-download-queue';
    protected $description = '从下载队列中取出待下载任务并派发 Job';

    public function __construct(
        private DownloadQueueService $downloadQueueService,
        private BilibiliSuspendService $bilibiliSuspendService
    ) {
        parent::__construct();
    }

    public function handle(): int
    {
        // 1. 重置卡死任务（running 超过 6 小时）
        $reset = $this->downloadQueueService->resetStuckTasks();
        if ($reset > 0) {
            Log::warning("Reset {$reset} stuck download tasks back to pending");
        }

        // 2. 检查风控暂停
        $suspendEndTime = $this->bilibiliSuspendService->getSuspendEndTime();
        if ($suspendEndTime !== null && $suspendEndTime > time()) {
            $this->info('Bilibili suspended, skip processing. Resume at: ' . date('Y-m-d H:i:s', $suspendEndTime));
            return self::SUCCESS;
        }

        // 3. 计算可用槽位
        $concurrency = (int) config('services.bilibili.download_concurrency', 3);
        $running     = DownloadQueue::where('status', DownloadQueue::STATUS_RUNNING)->count();
        $slots       = $concurrency - $running;

        if ($slots <= 0) {
            $this->info("All {$concurrency} download slots busy, skipping.");
            return self::SUCCESS;
        }

        // 4. 在事务内取出并标记 running（防止调度命令重叠执行时重复分配）
        $dispatched = 0;
        DB::transaction(function () use ($slots, &$dispatched) {
            $tasks = $this->downloadQueueService->getNextBatch($slots);

            foreach ($tasks as $task) {
                $this->downloadQueueService->markRunning($task);

                if ($task->type === DownloadQueue::TYPE_AUDIO) {
                    $audioPart = AudioPart::where('video_id', $task->video_id)->first();
                    if ($audioPart) {
                        DownloadAudioJob::dispatch($audioPart);
                        $dispatched++;
                        Log::info("Dispatched DownloadAudioJob", ['video_id' => $task->video_id]);
                    } else {
                        $task->update([
                            'status'    => DownloadQueue::STATUS_FAILED,
                            'error_msg' => 'AudioPart not found',
                        ]);
                    }
                } else {
                    $videoPart = VideoPart::find($task->video_part_id);
                    if ($videoPart) {
                        DownloadVideoJob::dispatch($videoPart);
                        $dispatched++;
                        Log::info("Dispatched DownloadVideoJob", ['video_part_id' => $task->video_part_id]);
                    } else {
                        $task->update([
                            'status'    => DownloadQueue::STATUS_FAILED,
                            'error_msg' => 'VideoPart not found',
                        ]);
                    }
                }
            }
        });

        $this->info("Dispatched {$dispatched} download job(s). Running: {$running}/{$concurrency}");

        // 5. 定期清理历史记录（概率触发，约每 100 次执行一次）
        if (rand(1, 100) === 1) {
            $cleaned = $this->downloadQueueService->cleanupHistory(7);
            if ($cleaned > 0) {
                Log::info("Cleaned up {$cleaned} old download queue records");
            }
        }

        return self::SUCCESS;
    }
}
