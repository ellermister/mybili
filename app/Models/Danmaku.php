<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Danmaku extends Model
{
    protected $table      = 'danmaku';
    protected $fillable   = ['id', 'video_id', 'cid', 'progress', 'mode', 'color', 'content', 'created_at'];
    protected $primaryKey = 'id';
    const UPDATED_AT = null;
    protected $createdAt  = 'created_at';
}
