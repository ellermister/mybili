<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Storage;

class Video extends Model
{
    protected $table      = 'videos';
    protected $fillable   = ['id', 'link', 'title', 'intro', 'cover', 'bvid', 'pubtime', 'duration', 'attr', 'invalid', 'frozen', 'cache_image', 'page', 'fav_time', 'danmaku_downloaded_at', 'video_downloaded_at','upper_id'];
    protected $primaryKey = 'id';

    protected $casts = [
        // 移除时间字段的 cast，使用访问器代替，不能使用cast，因为写入时格式无法匹配，错误写入
    ];

    protected $appends = [
        'cache_image_url',
    ];

    protected $attributes = [
        'cache_image' => '',
        'attr'        => 0,
        'invalid'     => false,
        'frozen'      => false,
        'page'        => 1,
        'fav_time'    => null,
        'upper_id'    => null,
    ];

    public function parts()
    {
        return $this->hasMany(VideoPart::class, 'video_id', 'id');
    }

    public function getCacheImageUrlAttribute()
    {
        return $this->cache_image ? Storage::url($this->cache_image) : null;
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
}
