<?php
namespace App\Services\VideoManager\Actions\Video;

use App\Enums\SettingKey;
use App\Events\VideoPartDownloaded;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\BilibiliService;
use App\Services\BilibiliSuspendService;
use App\Services\DownloadFilterService;
use App\Services\DownloadVideoService;
use App\Services\SettingsService;
use Carbon\Carbon;
use Log;

class DownloadVideoPartFileAction
{

    public function __construct(
        public BilibiliService $bilibiliService,
        public DownloadFilterService $downloadFilterService,
        public SettingsService $settingsService,
        public BilibiliSuspendService $bilibiliSuspendService,
        public DownloadVideoService $downloadVideoService
    ) {

    }

    public function execute(VideoPart $videoPart): void
    {
        $video = $videoPart->video;
        if (video_has_invalid($video->toArray())) {
            Log::info('exclude video by invalid', ['video_id' => $video->id, 'title' => $video->title]);
            return;
        }

        if ($this->downloadFilterService->shouldExcludeByName($video->title)) {
            Log::info('exclude video by name', ['video_id' => $video->id, 'title' => $video->title]);
            return;
        }

        // 判断是否开启多P缓存
        if (! $this->downloadFilterService->isMultiPEnabled() && $videoPart->page > 1) {
            Log::info('multiP is not enabled', ['video_id' => $video->id, 'part' => $videoPart->page]);
            return;
        }

        $url = config('services.bilibili.id_type') == 'bv' ? sprintf('https://www.bilibili.com/video/%s/', $video->bvid) : sprintf('https://www.bilibili.com/video/av%s/', $video->id);

        // 获取音乐节特辑类的链接
        $festivalJumpUrl = $this->bilibiliService->getVideoFestivalJumpUrl($video->id);
        if ($festivalJumpUrl) {
            $url = $festivalJumpUrl;
        }

        $filePath = $this->downloadVideoService->getVideoPartValidFilePath($videoPart);
        if ($filePath) {
            $this->updateVideoPartDownloaded($videoPart, $filePath);
            Log::info('video file already exists', ['video_id' => $video->id, 'part' => $videoPart->page, 'filePath' => $filePath]);
            return;
        } else {
            // 如果不存在就去下载
            $this->downloadVideoService->createDownloadDirectory();
            $binPath = base_path('download-video.sh');

            $savePath = $this->downloadVideoService->buildVideoPartFilePath($videoPart);
            Log::info('download video', [
                'video_id' => $video->id,
                'part'     => $videoPart->page,
                'url'      => escapeshellarg($url),
                'savePath' => escapeshellarg($savePath),
                'binPath'  => escapeshellarg($binPath),
            ]);

            list($output, $result) = $this->execDownloadVideo($url, $savePath, $videoPart->page);
            if ($result != 0) {
                $msg = implode('', $output);
                throw new \Exception("下载异常:\n" . $msg);
            }
            Log::info('download video output', ['video_id' => $video->id, 'part' => $videoPart->page, 'output' => $output]);
            Log::info('download video success', ['video_id' => $video->id, 'part' => $videoPart->page, 'savePath' => $savePath]);
            $this->updateVideoPartDownloaded($videoPart, $savePath);

            event(new VideoPartDownloaded($videoPart));
        }
    }

    protected function execDownloadVideo(string $url, string $savePath, int $page): array
    {
        if (config('services.bilibili.ignore_cookies')) {
            $command = sprintf('yt-dlp_linux -f "bestvideo+bestaudio/best" --playlist-items %s -o %s %s', escapeshellarg($page), escapeshellarg($savePath), escapeshellarg($url));
        } else {
            // 写入cookie到文件
            $cookiePath = storage_path('app/cookie.txt');
            file_put_contents($cookiePath, $this->settingsService->get(SettingKey::COOKIES_CONTENT));
            $command = sprintf('yt-dlp_linux -f "bestvideo+bestaudio/best" --playlist-items %s --cookies=%s -o %s %s', escapeshellarg($page), escapeshellarg($cookiePath), escapeshellarg($savePath), escapeshellarg($url));
        }
        exec($command, $output, $result);

        if ($result != 0) {
            $msg = implode('', $output);
            if (strpos($msg, "412") !== false) {
                Log::error("Check 412 error, set suspend for bilibili high rate limit", ['msg' => $msg, 'url' => $url, 'page' => $page]);
                $this->bilibiliSuspendService->setSuspend();
            }
        }
        return [$output, $result];
    }

    public function updateVideoPartDownloaded(VideoPart $videoPart, string $savePath): void
    {
        if (is_file($savePath)) {
            $calcHash = hash_file('sha256', $savePath);
            if ($calcHash !== $this->downloadVideoService->getDownloadHash($savePath)) {
                $this->downloadVideoService->saveDownloadHash($savePath, $calcHash);
            }

            $videoPart->video_downloaded_at = Carbon::createFromTimestamp(filectime($savePath));
            $videoPart->video_download_path = get_relative_path($savePath);
            $videoPart->save();
        } else {
            $videoPart->video_downloaded_at = null;
            $videoPart->video_download_path = null;
            $videoPart->save();
        }

        $video                       = Video::where('id', $videoPart->video_id)->first();
        $video->video_downloaded_num = VideoPart::where('video_id', $video->id)->whereNotNull('video_download_path')->count();
        $video->save();
    }

}
