<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Coverables extends Model
{
    protected $fillable = ['cover_id', 'coverable_id', 'coverable_type'];

    public function cover()
    {
        return $this->belongsTo(Cover::class);
    }

    public function coverable()
    {
        return $this->morphTo();
    }
}
