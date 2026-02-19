<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Facades\Storage;

class AudioPart extends Model
{
    protected $table    = 'audio_parts';
    protected $fillable = ['video_id', 'sid', 'duration', 'audio_downloaded_at', 'audio_download_path'];

    protected $appends = [
        'audio_download_url',
    ];

    public function video()
    {
        return $this->belongsTo(Video::class, 'video_id', 'id');
    }

    public function getAudioDownloadUrlAttribute(): ?string
    {
        return $this->audio_download_path ? Storage::url($this->audio_download_path) : null;
    }
}
