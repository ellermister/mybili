<?php
namespace App\Services\VideoManager\Actions\Audio;

use App\Enums\SettingKey;
use App\Models\AudioPart;
use App\Models\Video;
use App\Services\BilibiliSuspendService;
use App\Services\DownloadVideoService;
use App\Services\SettingsService;
use Carbon\Carbon;
use Illuminate\Support\Facades\Log;

class DownloadAudioPartFileAction
{
    public function __construct(
        public SettingsService $settingsService,
        public BilibiliSuspendService $bilibiliSuspendService,
        public DownloadVideoService $downloadVideoService
    ) {
    }

    public function execute(AudioPart $audioPart): void
    {
        $url      = sprintf('https://www.bilibili.com/audio/au%s', $audioPart->sid);
        $savePath = $this->buildAudioFilePath($audioPart);

        $filePath = $this->getValidFilePath($audioPart);
        if ($filePath) {
            $this->updateAudioPartDownloaded($audioPart, $filePath);
            Log::info('Audio file already exists', ['sid' => $audioPart->sid, 'filePath' => $filePath]);
            return;
        }

        $this->downloadVideoService->createDownloadDirectory();

        Log::info('Download audio', [
            'sid'      => $audioPart->sid,
            'url'      => $url,
            'savePath' => $savePath,
        ]);

        [$output, $result] = $this->execDownloadAudio($url, $savePath);

        if ($result != 0) {
            $msg = implode('', $output);
            throw new \Exception("音频下载异常:\n" . $msg);
        }

        Log::info('Download audio success', ['sid' => $audioPart->sid, 'savePath' => $savePath]);
        $this->updateAudioPartDownloaded($audioPart, $savePath);
    }

    protected function execDownloadAudio(string $url, string $savePath): array
    {
        if (config('services.bilibili.ignore_cookies')) {
            $command = sprintf(
                'yt-dlp_linux -f "bestaudio/best" --playlist-items 1 -o %s %s',
                escapeshellarg($savePath),
                escapeshellarg($url)
            );
        } else {
            $cookiePath = storage_path('app/cookie.txt');
            file_put_contents($cookiePath, $this->settingsService->get(SettingKey::COOKIES_CONTENT));
            $command = sprintf(
                'yt-dlp_linux -f "bestaudio/best" --playlist-items 1 --cookies=%s -o %s %s',
                escapeshellarg($cookiePath),
                escapeshellarg($savePath),
                escapeshellarg($url)
            );
        }

        exec($command, $output, $result);

        if ($result != 0) {
            $msg = implode('', $output);
            if (strpos($msg, '412') !== false) {
                Log::error('Check 412 error, set suspend for bilibili high rate limit', ['msg' => $msg, 'url' => $url]);
                $this->bilibiliSuspendService->setSuspend();
            }
        }

        return [$output, $result];
    }

    public function buildAudioFilePath(AudioPart $audioPart): string
    {
        $path = storage_path('app/public/' . $this->downloadVideoService->downloadFolder);
        return sprintf('%s/au%s.mp3', $path, $audioPart->sid);
    }

    public function getValidFilePath(AudioPart $audioPart): ?string
    {
        $savePath = $this->buildAudioFilePath($audioPart);
        if (is_file($savePath)) {
            return $savePath;
        }
        return null;
    }

    public function updateAudioPartDownloaded(AudioPart $audioPart, string $savePath): void
    {
        $audioDownloadedNum = 0;
        if (is_file($savePath)) {
            $audioPart->audio_downloaded_at = Carbon::createFromTimestamp(filectime($savePath));
            $audioPart->audio_download_path = get_relative_path($savePath);
            $audioPart->save();
            $audioDownloadedNum = 1;
        } else {
            $audioPart->audio_downloaded_at = null;
            $audioPart->audio_download_path = null;
            $audioPart->save();
            $audioDownloadedNum = 0;
        }

        $video = Video::where('id', $audioPart->video_id)->first();
        if ($video) {
            $video->audio_downloaded_num = $audioDownloadedNum;
            $video->save();
        }
    }
}
