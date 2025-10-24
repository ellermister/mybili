<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Subscription extends Model
{
    protected $table = 'subscriptions';
    protected $primaryKey = 'id';

    protected $fillable = ['type', 'mid', 'season_id', 'name', 'description', 'cover', 'url', 'total', 'status', 'last_check_at'];
    
    const STATUS_ACTIVE = 1;
    const STATUS_DISABLED = 0;

    protected $casts = [
        'last_check_at' => 'datetime:Y-m-d H:i',
    ];

    protected $appends = [
        'cover_info',
    ];

    public function videos()
    {
        return $this->belongsToMany(Video::class, 'subscription_videos', 'subscription_id', 'video_id');
    }

    /**
     * 获取视频封面（通过多态关联）
     * 返回 Cover 模型，可以访问封面的完整信息
     */
    public function coverImage()
    {
        return $this->morphToMany(Cover::class, 'coverable', 'coverables')
                    ->withTimestamps();
    }
    /**
     * 获取单个封面（由于唯一约束，一个视频只有一个封面）
     * 使用示例: $video->coverImage()->first()
     * 或者使用访问器: $video->cover_info
     */
    public function getCoverInfoAttribute()
    {
        return $this->coverImage()->first();
    }
}
