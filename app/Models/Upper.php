<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Upper extends Model
{
    protected $table = 'upper';
    protected $fillable = ['mid', 'name', 'face',];
    protected $casts = [
        'mid' => 'integer',
        'name' => 'string',
        'face' => 'string',
    ];

    protected $appends = [
        'cover_info',
    ];
    /**
     * 获取UP主封面（通过多态关联）
     */
    public function coverImage()
    {
        return $this->morphToMany(Cover::class, 'coverable', 'coverables')
                    ->withTimestamps();
    }
    /**
     * 获取单个封面（由于唯一约束，一个UP主只有一个封面）
     */
    public function getCoverInfoAttribute()
    {
        return $this->coverImage()->first();
    }
}

