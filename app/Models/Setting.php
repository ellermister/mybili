<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Setting extends Model
{
    protected $table      = 'settings';
    protected $fillable   = ['name', 'value'];
    protected $primaryKey = 'id';

    protected $casts = [
        'value' => 'json',
    ];
}
