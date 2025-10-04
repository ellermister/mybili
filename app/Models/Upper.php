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
}

