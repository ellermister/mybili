<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Storage;

class Video extends Model
{
    protected $table      = 'videos';
    protected $fillable   = ['id', 'link', 'title', 'intro', 'cover', 'bvid', 'pubtime', 'attr', 'invalid', 'frozen', 'cache_image', 'page', 'fav_time', 'danmaku_downloaded_at', 'video_downloaded_at'];
    protected $primaryKey = 'id';

    protected $casts = [
        'pubtime'  => 'timestamp',
        'fav_time' => 'timestamp',
    ];

    protected $appends = [
        'cache_image_url',
    ];

    protected $attributes = [
        'cache_image' => '',
    ];

    public function parts()
    {
        return $this->hasMany(VideoPart::class, 'video_id', 'id');
    }

    public function getCacheImageUrlAttribute()
    {
        return $this->cache_image ? Storage::url($this->cache_image) : null;
    }

    public function favorite()
    {
        return $this->belongsToMany(FavoriteList::class, 'favorite_list_videos', 'video_id', 'favorite_list_id');
    }

    public function subscriptions()
    {
        return $this->belongsToMany(Subscription::class, 'subscription_videos', 'video_id', 'subscription_id');
    }
}
