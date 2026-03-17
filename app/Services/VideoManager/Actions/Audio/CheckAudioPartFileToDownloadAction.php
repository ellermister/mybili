<?php
namespace App\Services\VideoManager\Actions\Audio;

use App\Enums\SettingKey;
use App\Models\AudioPart;
use App\Models\Video;
use App\Services\DownloadFilterService;
use App\Services\DownloadQueueService;
use App\Services\SettingsService;
use Carbon\Carbon;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;

class CheckAudioPartFileToDownloadAction
{
    public function __construct(
        public SettingsService $settingsService,
        public DownloadFilterService $downloadFilterService,
        public DownloadQueueService $downloadQueueService
    ) {
    }

    public function execute(AudioPart $audioPart, bool $tryDownload = false): void
    {
        $video = $audioPart->video;
        if (! $video) {
            $video = Video::where('id', $audioPart->video_id)->first();
        }

        if ($video) {
            if ($this->downloadFilterService->shouldExcludeByVideo($video)) {
                Log::info('Audio download excluded by favorite and subscription', ['sid' => $audioPart->sid]);
                return;
            }

            if ($this->downloadFilterService->shouldExcludeByName((string) $video->title)) {
                Log::info('Audio download excluded by name', ['sid' => $audioPart->sid, 'title' => $video->title]);
                return;
            }

            if ($this->downloadFilterService->shouldExcludeByFavTime($video)) {
                Log::info('Audio download excluded by favorite time', ['sid' => $audioPart->sid]);
                return;
            }

            if ($this->downloadFilterService->shouldExcludeByDuration((int) $video->duration)) {
                Log::info('Audio download excluded by duration', ['sid' => $audioPart->sid, 'duration' => $video->duration]);
                return;
            }
        }

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
                $this->downloadQueueService->enqueueAudio($audioPart);
            } else {
                Log::info('Audio file not exists, download disabled', ['sid' => $audioPart->sid]);
            }
        } else {
            Log::info('Audio file already exists', ['sid' => $audioPart->sid]);

            $audioPart->audio_downloaded_at = Carbon::createFromTimestamp(filectime($savePath));
            $audioPart->audio_download_path = get_relative_path($savePath);
            $audioPart->save();
        }
    }
}
