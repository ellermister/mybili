<?php
namespace App\Listeners;

use App\Events\AudioPartUpdated;
use App\Models\AudioPart;
use App\Services\VideoManager\Actions\Audio\CheckAudioPartFileToDownloadAction;
use Illuminate\Contracts\Queue\ShouldQueue;

class AudioPartFileDownload implements ShouldQueue
{
    public $queue = 'default';

    public function __construct()
    {
    }

    public function handle(AudioPartUpdated $event): void
    {
        if (isset($event->newAudioPart['video_id'])) {
            $audioPart = AudioPart::where('video_id', $event->newAudioPart['video_id'])->first();
            if ($audioPart) {
                app(CheckAudioPartFileToDownloadAction::class)->execute($audioPart, true);
            }
        }
    }
}
