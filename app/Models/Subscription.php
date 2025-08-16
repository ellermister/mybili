<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use App\Models\SubscriptionVideo;
use Storage;

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
        'cache_image_url',
    ];

    public function videos()
    {
        return $this->belongsToMany(Video::class, 'subscription_videos', 'subscription_id', 'video_id');
    }

    public function getCacheImageUrlAttribute()
    {
        return $this->cache_image ? Storage::url($this->cache_image) : null;
    }
}
