<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class FavoriteListVideo extends Model
{
    protected $table = 'favorite_list_videos';
    protected $fillable = ['favorite_list_id', 'video_id', 'created_at'];
    protected $primaryKey = ['favorite_list_id', 'video_id'];
    public $incrementing = false;
    public $timestamps = true;
}
