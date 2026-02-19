<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;
use App\Models\AudioPart;

class Video extends Model
{

    use SoftDeletes;

    protected $table      = 'videos';
    protected $fillable   = ['id', 'link', 'title', 'intro', 'cover', 'bvid', 'pubtime', 'duration', 'attr', 'invalid', 'frozen', 'page', 'fav_time', 'danmaku_downloaded_at', 'video_downloaded_at', 'upper_id', 'type'];
    protected $primaryKey = 'id';

    protected $casts = [
        // 移除时间字段的 cast，使用访问器代替，不能使用cast，因为写入时格式无法匹配，错误写入
    ];

    protected $appends = [
        'cover_info',
    ];

    protected $attributes = [
        'cache_image' => '',
        'attr'        => 0,
        'invalid'     => false,
        'frozen'      => false,
        'page'        => 1,
        'fav_time'    => null,
        'upper_id'    => null,
        'type'        => 2,
    ];

    public function parts()
    {
        return $this->hasMany(VideoPart::class, 'video_id', 'id');
    }

    public function audioPart()
    {
        return $this->hasOne(AudioPart::class, 'video_id', 'id');
    }

    public function isAudio(): bool
    {
        return (int) $this->type === 12;
    }

    // 访问器：读取时将时间戳转换为 Carbon 对象
    public function getPubtimeAttribute($value)
    {
        return $value ? strtotime($value) : null;
    }

    public function getFavTimeAttribute($value)
    {
        return $value ? strtotime($value) : null;
    }

    public function favorite()
    {
        return $this->belongsToMany(FavoriteList::class, 'favorite_list_videos', 'video_id', 'favorite_list_id');
    }

    public function subscriptions()
    {
        return $this->belongsToMany(Subscription::class, 'subscription_videos', 'video_id', 'subscription_id');
    }

    public function upper()
    {
        return $this->belongsTo(Upper::class, 'upper_id', 'mid');
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
