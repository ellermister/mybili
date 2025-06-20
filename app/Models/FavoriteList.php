<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Storage;

class FavoriteList extends Model
{
    protected $table = 'favorite_lists';
    protected $fillable = ['id', 'title', 'cover', 'ctime', 'mtime', 'media_count', 'cache_image'];
    protected $primaryKey = 'id';

    protected $casts = [
        'ctime' => 'timestamp',
        'mtime' => 'timestamp',
    ];

    protected $appends = [
        'cache_image_url',
    ];

    public function videos()
    {
        return $this->belongsToMany(Video::class, 'favorite_list_videos', 'favorite_list_id', 'video_id')
                    ->orderBy('fav_time', 'desc');
    }

    public function getCacheImageUrlAttribute()
    {
        return $this->cache_image ? Storage::url($this->cache_image) : null;
    }
}
