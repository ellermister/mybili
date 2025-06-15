<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class VideoPart extends Model
{
    protected $table    = 'video_parts';
    protected $fillable = ['video_id', 'cid', 'page', 'from', 'part', 'duration', 'vid', 'weblink', 'width', 'height', 'rotate', 'first_frame', 'created_at', 'video_downloaded_at','danmaku_downloaded_at'];

    public function video()
    {
        return $this->belongsTo(Video::class, 'video_id', 'id');
    }

    public function danmakus()
    {
        return $this->hasMany(Danmaku::class, 'cid', 'cid');
    }
}
