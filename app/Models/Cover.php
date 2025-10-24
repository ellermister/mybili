<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Storage;

class Cover extends Model
{
    protected $fillable = ['url', 'type', 'filename', 'path', 'mime_type', 'size', 'width', 'height'];

    const UPDATED_AT = null;

    protected $appends = [
        'image_url',
    ];

    /**
     * 获取所有使用此封面的关联记录
     */
    public function coverables()
    {
        return $this->hasMany(Coverables::class, 'cover_id');
    }

    /**
     * 获取所有使用此封面的视频
     */
    public function videos()
    {
        return $this->morphedByMany(Video::class, 'coverable', 'coverables');
    }

    /**
     * 获取所有使用此封面的收藏夹
     */
    public function favoriteLists()
    {
        return $this->morphedByMany(FavoriteList::class, 'coverable', 'coverables');
    }

    /**
     * 获取所有使用此封面的UP主
     */
    public function uppers()
    {
        return $this->morphedByMany(Upper::class, 'coverable', 'coverables');
    }

    public function getImageUrlAttribute()
    {
        return $this->path ? Storage::url($this->path) : null;
    }
}
