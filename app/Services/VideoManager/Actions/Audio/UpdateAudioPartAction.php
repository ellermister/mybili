<?php
namespace App\Services\VideoManager\Actions\Audio;

use App\Events\AudioPartUpdated;
use App\Models\AudioPart;
use App\Models\Video;
use App\Services\BilibiliService;
use Illuminate\Support\Facades\Log;

class UpdateAudioPartAction
{
    public function __construct(
        public BilibiliService $bilibiliService
    ) {
    }

    public function execute(Video $video): void
    {
        $sid = (int) $video->id;

        try {
            $audioInfo = $this->bilibiliService->getAudioInfo($sid);
        } catch (\Exception $e) {
            Log::error('Get audio info failed', ['video_id' => $video->id, 'sid' => $sid, 'title' => $video->title]);
            return;
        }

        $existing = AudioPart::where('video_id', $video->id)->first();
        $oldData  = $existing ? $existing->toArray() : [];

        $audioPart = AudioPart::updateOrCreate(
            ['video_id' => $video->id],
            [
                'sid'      => $sid,
                'duration' => $audioInfo['duration'] ?? 0,
            ]
        );

        Log::info('Update audio part success', ['video_id' => $video->id, 'sid' => $sid, 'title' => $video->title]);

        event(new AudioPartUpdated($oldData, $audioPart->fresh()->toArray()));
    }
}
