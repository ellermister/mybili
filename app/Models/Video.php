<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Video extends Model
{
    protected $table      = 'videos';
    protected $fillable   = ['id', 'link', 'title', 'intro', 'cover', 'bvid', 'pubtime', 'attr', 'invalid', 'frozen', 'cache_image', 'page', 'fav_time', 'danmaku_downloaded_at', 'video_downloaded_at'];
    protected $primaryKey = 'id';

    protected $casts = [
        'pubtime'  => 'timestamp',
        'fav_time' => 'timestamp',
    ];

    public function parts()
    {
        return $this->hasMany(VideoPart::class, 'video_id', 'id');
    }
}
