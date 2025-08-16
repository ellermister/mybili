<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use App\Models\SubscriptionVideo;

class Subscription extends Model
{
    protected $table = 'subscriptions';
    protected $primaryKey = 'id';

    protected $fillable = ['type', 'mid', 'season_id', 'name', 'description', 'cover', 'url', 'total', 'status', 'last_check_at'];
    
    const STATUS_ACTIVE = 1;
    const STATUS_DISABLED = 0;

    public function videos()
    {
        return $this->hasMany(SubscriptionVideo::class, 'subscription_id', 'id');
    }
}
