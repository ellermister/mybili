<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class FavoriteList extends Model
{
    protected $table = 'favorite_lists';
    protected $fillable = ['id', 'title', 'cover', 'ctime', 'mtime', 'media_count', 'cache_image'];
    protected $primaryKey = 'id';

    protected $casts = [
        'ctime' => 'timestamp',
        'mtime' => 'timestamp',
    ];

    public function videos()
    {
        return $this->belongsToMany(Video::class, 'favorite_list_videos', 'favorite_list_id', 'video_id');
    }
}
