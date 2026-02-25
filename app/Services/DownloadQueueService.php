<?php
namespace App\Services;

use App\Models\AudioPart;
use App\Models\DownloadQueue;
use App\Models\VideoPart;
use Illuminate\Database\Eloquent\Collection;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class DownloadQueueService
{
    /**
     * 将视频分P加入下载队列，已存在则忽略
     */
    public function enqueueVideo(VideoPart $videoPart, int $priority = 0): ?DownloadQueue
    {
        $uniqueKey = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_VIDEO, $videoPart->id);

        $existing = DownloadQueue::where('unique_key', $uniqueKey)
            ->whereIn('status', [DownloadQueue::STATUS_PENDING, DownloadQueue::STATUS_RUNNING])
            ->first();

        if ($existing) {
            return null;
        }

        $item = DownloadQueue::updateOrCreate(
            ['unique_key' => $uniqueKey],
            [
                'type'          => DownloadQueue::TYPE_VIDEO,
                'video_id'      => $videoPart->video_id,
                'video_part_id' => $videoPart->id,
                'status'        => DownloadQueue::STATUS_PENDING,
                'priority'      => $priority,
                'retry_count'   => 0,
                'error_msg'     => null,
                'scheduled_at'  => null,
                'completed_at'  => null,
            ]
        );

        Log::info('Enqueued video download', ['video_part_id' => $videoPart->id, 'video_id' => $videoPart->video_id]);
        return $item;
    }

    /**
     * 将音频加入下载队列，已存在则忽略
     */
    public function enqueueAudio(AudioPart $audioPart, int $priority = 0): ?DownloadQueue
    {
        $uniqueKey = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_AUDIO, $audioPart->video_id);

        $existing = DownloadQueue::where('unique_key', $uniqueKey)
            ->whereIn('status', [DownloadQueue::STATUS_PENDING, DownloadQueue::STATUS_RUNNING])
            ->first();

        if ($existing) {
            return null;
        }

        $item = DownloadQueue::updateOrCreate(
            ['unique_key' => $uniqueKey],
            [
                'type'          => DownloadQueue::TYPE_AUDIO,
                'video_id'      => $audioPart->video_id,
                'video_part_id' => null,
                'status'        => DownloadQueue::STATUS_PENDING,
                'priority'      => $priority,
                'retry_count'   => 0,
                'error_msg'     => null,
                'scheduled_at'  => null,
                'completed_at'  => null,
            ]
        );

        Log::info('Enqueued audio download', ['video_id' => $audioPart->video_id, 'sid' => $audioPart->sid]);
        return $item;
    }

    /**
     * 获取下一批 pending 任务（在事务内调用）
     */
    public function getNextBatch(int $limit): Collection
    {
        return DownloadQueue::where('status', DownloadQueue::STATUS_PENDING)
            ->orderByDesc('priority')
            ->orderBy('id')
            ->limit($limit)
            ->get();
    }

    public function markRunning(DownloadQueue $item): void
    {
        $item->update([
            'status'       => DownloadQueue::STATUS_RUNNING,
            'scheduled_at' => now(),
        ]);
    }

    /**
     * 限流时归还槽位：running → pending，不消耗 retry_count。
     * Job 干净退出，由 ProcessDownloadQueue 下次轮询重新派发。
     */
    public function markPendingByVideoPart(int $videoPartId): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_VIDEO, $videoPartId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if ($item && $item->status === DownloadQueue::STATUS_RUNNING) {
            $item->update(['status' => DownloadQueue::STATUS_PENDING]);
        }
    }

    public function markPendingByAudio(int $videoId): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_AUDIO, $videoId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if ($item && $item->status === DownloadQueue::STATUS_RUNNING) {
            $item->update(['status' => DownloadQueue::STATUS_PENDING]);
        }
    }

    public function markDoneByVideoPart(int $videoPartId): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_VIDEO, $videoPartId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if ($item) {
            $item->update([
                'status'       => DownloadQueue::STATUS_DONE,
                'completed_at' => now(),
                'error_msg'    => null,
            ]);
        }
    }

    public function markDoneByAudio(int $videoId): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_AUDIO, $videoId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if ($item) {
            $item->update([
                'status'       => DownloadQueue::STATUS_DONE,
                'completed_at' => now(),
                'error_msg'    => null,
            ]);
        }
    }

    public function markFailedByVideoPart(int $videoPartId, string $error): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_VIDEO, $videoPartId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if ($item) {
            $item->update([
                'status'       => DownloadQueue::STATUS_FAILED,
                'completed_at' => now(),
                'error_msg'    => mb_substr($error, 0, 500),
            ]);
        }
    }

    public function markFailedByAudio(int $videoId, string $error): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_AUDIO, $videoId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if ($item) {
            $item->update([
                'status'       => DownloadQueue::STATUS_FAILED,
                'completed_at' => now(),
                'error_msg'    => mb_substr($error, 0, 500),
            ]);
        }
    }

    /**
     * Job 失败后：未超过最大重试次数则重回 pending，等待 ProcessDownloadQueue 重新派发；
     * 超过次数后标记为永久 failed，释放槽位且不再重试。
     */
    public function markRetryOrFailedByVideoPart(int $videoPartId, string $error): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_VIDEO, $videoPartId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if (! $item) {
            return;
        }

        $retryCount = $item->retry_count + 1;

        if ($retryCount >= DownloadQueue::MAX_RETRIES) {
            $item->update([
                'status'       => DownloadQueue::STATUS_FAILED,
                'retry_count'  => $retryCount,
                'completed_at' => now(),
                'error_msg'    => mb_substr($error, 0, 500),
            ]);
            Log::warning('DownloadVideoJob permanently failed', ['video_part_id' => $videoPartId, 'retries' => $retryCount, 'error' => $error]);
        } else {
            $item->update([
                'status'      => DownloadQueue::STATUS_PENDING,
                'retry_count' => $retryCount,
                'error_msg'   => mb_substr("[retry {$retryCount}] " . $error, 0, 500),
            ]);
            Log::info('DownloadVideoJob failed, reset to pending for retry', ['video_part_id' => $videoPartId, 'retry_count' => $retryCount]);
        }
    }

    public function markRetryOrFailedByAudio(int $videoId, string $error): void
    {
        $key  = DownloadQueue::buildUniqueKey(DownloadQueue::TYPE_AUDIO, $videoId);
        $item = DownloadQueue::where('unique_key', $key)->first();
        if (! $item) {
            return;
        }

        $retryCount = $item->retry_count + 1;

        if ($retryCount >= DownloadQueue::MAX_RETRIES) {
            $item->update([
                'status'       => DownloadQueue::STATUS_FAILED,
                'retry_count'  => $retryCount,
                'completed_at' => now(),
                'error_msg'    => mb_substr($error, 0, 500),
            ]);
            Log::warning('DownloadAudioJob permanently failed', ['video_id' => $videoId, 'retries' => $retryCount, 'error' => $error]);
        } else {
            $item->update([
                'status'      => DownloadQueue::STATUS_PENDING,
                'retry_count' => $retryCount,
                'error_msg'   => mb_substr("[retry {$retryCount}] " . $error, 0, 500),
            ]);
            Log::info('DownloadAudioJob failed, reset to pending for retry', ['video_id' => $videoId, 'retry_count' => $retryCount]);
        }
    }

    /**
     * 取消任务（只有 pending 状态可取消）
     */
    public function cancel(int $id): bool
    {
        $item = DownloadQueue::find($id);
        if (! $item || $item->status !== DownloadQueue::STATUS_PENDING) {
            return false;
        }
        $item->update(['status' => DownloadQueue::STATUS_CANCELLED]);
        return true;
    }

    /**
     * 将失败/取消的任务重新加入待下载
     */
    public function retry(int $id): bool
    {
        $item = DownloadQueue::find($id);
        if (! $item || ! in_array($item->status, [DownloadQueue::STATUS_FAILED, DownloadQueue::STATUS_CANCELLED])) {
            return false;
        }
        $item->update([
            'status'       => DownloadQueue::STATUS_PENDING,
            'retry_count'  => 0,
            'error_msg'    => null,
            'scheduled_at' => null,
            'completed_at' => null,
        ]);
        return true;
    }

    /**
     * 更新优先级
     */
    public function setPriority(int $id, int $priority): bool
    {
        $item = DownloadQueue::find($id);
        if (! $item) {
            return false;
        }
        $item->update(['priority' => $priority]);
        return true;
    }

    /**
     * 重置卡死任务：running 超过 STUCK_TIMEOUT 秒的重置为 pending
     */
    public function resetStuckTasks(): int
    {
        $cutoff = now()->subSeconds(DownloadQueue::STUCK_TIMEOUT);
        return DownloadQueue::where('status', DownloadQueue::STATUS_RUNNING)
            ->where('scheduled_at', '<', $cutoff)
            ->update([
                'status'    => DownloadQueue::STATUS_PENDING,
                'error_msg' => 'Reset: stuck timeout',
            ]);
    }

    /**
     * 清理已完成/已取消的历史记录（保留 N 天）
     */
    public function cleanupHistory(int $days = 7): int
    {
        return DownloadQueue::whereIn('status', [DownloadQueue::STATUS_DONE, DownloadQueue::STATUS_CANCELLED])
            ->where('updated_at', '<', now()->subDays($days))
            ->delete();
    }

    /**
     * 获取队列统计信息
     */
    public function getStat(): array
    {
        $rows = DownloadQueue::selectRaw(
            "status, COUNT(*) as cnt"
        )->groupBy('status')->pluck('cnt', 'status')->toArray();

        return [
            'pending'   => (int) ($rows[DownloadQueue::STATUS_PENDING]   ?? 0),
            'running'   => (int) ($rows[DownloadQueue::STATUS_RUNNING]   ?? 0),
            'done'      => (int) ($rows[DownloadQueue::STATUS_DONE]      ?? 0),
            'failed'    => (int) ($rows[DownloadQueue::STATUS_FAILED]    ?? 0),
            'cancelled' => (int) ($rows[DownloadQueue::STATUS_CANCELLED] ?? 0),
        ];
    }
}
