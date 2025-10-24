<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class FavoriteList extends Model
{
    protected $table      = 'favorite_lists';
    protected $fillable   = ['id', 'title', 'cover', 'ctime', 'mtime', 'media_count'];
    protected $primaryKey = 'id';

    protected $casts = [
        'ctime' => 'timestamp',
        'mtime' => 'timestamp',
    ];

    protected $appends = [
        'cover_info',
    ];

    public function videos()
    {
        return $this->belongsToMany(Video::class, 'favorite_list_videos', 'favorite_list_id', 'video_id')
            ->orderBy('fav_time', 'desc');
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
