<?php
namespace App\Services\VideoManager\Actions\Audio;

use App\Enums\SettingKey;
use App\Jobs\DownloadAudioJob;
use App\Models\AudioPart;
use App\Models\Video;
use App\Services\SettingsService;
use Carbon\Carbon;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;

class CheckAudioPartFileToDownloadAction
{
    public function __construct(
        public SettingsService $settingsService
    ) {
    }

    public function execute(AudioPart $audioPart, bool $tryDownload = false): void
    {
        $isDownloaded = false;

        if ($audioPart->audio_download_path) {
            $savePath = Storage::disk('public')->path(Str::after($audioPart->audio_download_path, '/storage/'));
            if (is_file($savePath)) {
                $isDownloaded = true;
            }
        }

        if (! $isDownloaded) {
            if (! $tryDownload) {
                return;
            }

            if ($this->settingsService->get(SettingKey::VIDEO_DOWNLOAD_ENABLED) == 'on') {
                $lock = redis()->setnx(sprintf('audio_downloading:%s', $audioPart->sid), 1);
                if (! $lock) {
                    Log::info('Audio is being downloaded', ['sid' => $audioPart->sid]);
                    return;
                }
                redis()->expire(sprintf('audio_downloading:%s', $audioPart->sid), 3600 * 8);

                DownloadAudioJob::dispatchWithRateLimit($audioPart);
            } else {
                Log::info('Audio file not exists, download disabled', ['sid' => $audioPart->sid]);
            }
        } else {
            Log::info('Audio file already exists', ['sid' => $audioPart->sid]);

            $audioPart->audio_downloaded_at = Carbon::createFromTimestamp(filectime($savePath));
            $audioPart->audio_download_path = get_relative_path($savePath);
            $audioPart->save();

            $video = Video::where('id', $audioPart->video_id)->first();
            if ($video) {
                $video->audio_downloaded_num = 1;
                $video->save();
            }
        }
    }
}
