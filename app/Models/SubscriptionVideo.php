<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use App\Models\Subscription;
use App\Models\Video;

class SubscriptionVideo extends Model
{
    protected $table = 'subscription_videos';
    protected $primaryKey = ['subscription_id', 'video_id'];
    public $incrementing = false;

    protected $fillable = ['subscription_id', 'video_id', 'bvid'];

    public function subscription()
    {
        return $this->belongsTo(Subscription::class, 'subscription_id', 'id');
    }

    public function video()
    {
        return $this->belongsTo(Video::class, 'video_id', 'id');
    }
}
