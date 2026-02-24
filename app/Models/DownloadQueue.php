<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

/**
 * 下载队列
 *
 * status 状态流转：
 *   pending → running（ProcessDownloadQueueCommand pop 时）
 *   running → done   （下载成功后）
 *   running → failed （Job 所有重试耗尽后）
 *   pending → cancelled（用户主动取消）
 */
class DownloadQueue extends Model
{
    protected $table    = 'download_queue';
    protected $fillable = [
        'type', 'video_id', 'video_part_id',
        'status', 'priority', 'retry_count', 'error_msg',
        'unique_key', 'scheduled_at', 'completed_at',
    ];

    public const TYPE_VIDEO = 'video';
    public const TYPE_AUDIO = 'audio';

    public const STATUS_PENDING   = 'pending';
    public const STATUS_RUNNING   = 'running';
    public const STATUS_DONE      = 'done';
    public const STATUS_FAILED    = 'failed';
    public const STATUS_CANCELLED = 'cancelled';

    // Job 失败后最多重新派发次数（超过后标记为永久失败）
    public const MAX_RETRIES = 3;

    // 超过此时间仍处于 running 状态视为卡死，重置回 pending（秒）
    public const STUCK_TIMEOUT = 600;

    public static function buildUniqueKey(string $type, int $id): string
    {
        return "{$type}:{$id}";
    }

    public function video()
    {
        return $this->belongsTo(Video::class, 'video_id', 'id');
    }

    public function videoPart()
    {
        return $this->belongsTo(VideoPart::class, 'video_part_id', 'id');
    }

    public function isPending(): bool
    {
        return $this->status === self::STATUS_PENDING;
    }

    public function isRunning(): bool
    {
        return $this->status === self::STATUS_RUNNING;
    }
}
