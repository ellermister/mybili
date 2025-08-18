<?php
namespace App\Services;

use App\Contracts\VideoDownloadServiceInterface;
use App\Enums\SettingKey;
use App\Events\VideoPartDownloaded;
use App\Jobs\DownloadVideoJob;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\DownloadFilterService;
use Carbon\Carbon;
use Log;
use Storage;
use Str;

class VideoDownloadService implements VideoDownloadServiceInterface
{

    public function __construct(public DownloadFilterService $downloadFilterService, public SettingsService $settingsService, public BilibiliService $bilibiliService)
    {
    }

    /**
     * 视频文件和hash同时存在才认为有效
     */
    public function hasVideoFile(VideoPart $videoPart): bool
    {
        $result = $this->getVideoPartValidFilePath($videoPart);
        if ($result) {
            return true;
        }
        return false;
    }

    public function getVideoPartValidFilePath(VideoPart $videoPart): ?string
    {

        $savePath = $this->getVideoDownloadPath($videoPart->video_id, $videoPart->page);
        $hashPath = $this->getVideoDownloadHashPath($videoPart->video_id, $videoPart->page);
        if (is_file($savePath) && is_file($hashPath)) {
            return $savePath;
        }

        $savePath = $this->getVideoDownloadPath($videoPart->video_id, $videoPart->cid);
        $hashPath = $this->getVideoDownloadHashPath($videoPart->video_id, $videoPart->cid);
        if (is_file($savePath) && is_file($hashPath)) {
            return $savePath;
        }

        return null;
    }

    public function buildVideoPartFilePath(VideoPart $videoPart): string
    {
        return $this->getVideoDownloadPath($videoPart->video_id, $videoPart->cid);
    }

    protected function createVideoDirectory(): void
    {
        $videoPath = storage_path('app/public/videos');
        if (! is_dir($videoPath)) {
            mkdir($videoPath, 0644);
            if (! is_dir($videoPath)) {
                throw new \Exception("下载路径不存在, 创建失败");
            }
        }
    }

    protected function getVideoDownloadPath(string $id, int $part = 1): string
    {
        $videoPath = storage_path('app/public/videos');
        if ($part === 1) {
            return sprintf('%s/%s.mp4', $videoPath, $id);
        }
        return sprintf('%s/%s.part%d.mp4', $videoPath, $id, $part);
    }

    protected function getVideoDownloadHashPath(string $id, int $part = 1): string
    {
        $videoPath = storage_path('app/public/videos');
        if ($part === 1) {
            return sprintf('%s/%s.mp4.hash', $videoPath, $id);
        }
        return sprintf('%s/%s.part%d.mp4.hash', $videoPath, $id, $part);
    }

    protected function getVideoDownloadHash(string $savePath): ?string
    {
        $hashPath = sprintf('%s.hash', $savePath);
        if (is_file($hashPath)) {
            return trim(file_get_contents($hashPath));
        }
        return null;
    }

    protected function saveVideoDownloadHash(string $savePath, string $hash): void
    {
        $hashPath = sprintf('%s.hash', $savePath);
        file_put_contents($hashPath, $hash);
    }

    public function updateVideoPartDownloaded(VideoPart $videoPart, string $savePath): void
    {
        if (is_file($savePath)) {
            $calcHash = hash_file('sha256', $savePath);
            if ($calcHash !== $this->getVideoDownloadHash($savePath)) {
                $this->saveVideoDownloadHash($savePath, $calcHash);
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

    public function downloadVideoPartFile(VideoPart $videoPart, bool $tryDownload = false): void
    {
        $video = Video::where('id', $videoPart->video_id)->first();

        // 获取当前分片在视频第几个索引
        $videoParts = VideoPart::where('video_id', $video->id)->select(['video_id', 'page', 'cid'])->get();

        // 如果是全数字命名的，则数字对应的就是文件名索引
        $isNumberNamed = false;
        foreach ($videoParts as $item) {
            if (preg_match('/^\d+$/', $item['page']) && intval($item['page']) == $item['page']) {
                $isNumberNamed = true;
            } else {
                $isNumberNamed = false;
                break;
            }
        }

        $currentIndex = 1;
        if ($isNumberNamed) {
            $currentIndex = $videoPart->page;
        } else {
            // 如果不是数字索引的，则按照排序，取当前分片在视频第几个索引
            $videoParts   = $videoParts->sortBy('video_id');
            $currentIndex = $videoParts->search(function ($item) use ($videoPart) {
                return $item['cid'] == $videoPart->cid;
            }) + 1;
        }

        if ($video) {
            $isDownloaded = false;
            if ($videoPart->video_download_path) {
                $savePath     = Storage::disk('public')->path(Str::after($videoPart->video_download_path, '/storage/'));
                if (is_file($savePath)) {
                    $isDownloaded = true;
                }
            } else {
                $savePath = $this->getVideoDownloadPath($video->id, $currentIndex);
                if (is_file($savePath)) {
                    $isDownloaded = true;
                }
            }

            if (! $isDownloaded) {
                if (! $tryDownload) {
                    return;
                }

                // 下载视频分P文件
                if ($this->settingsService->get(SettingKey::VIDEO_DOWNLOAD_ENABLED) == 'on') {
                    // 加锁， 控流
                    $lock = redis()->setnx(sprintf('video_downloading:%s', $videoPart->cid), 1);
                    if (! $lock) {
                        Log::info('Video is being downloaded', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                        return;
                    }
                    redis()->expire(sprintf('video_downloading:%s', $videoPart->cid), 3600 * 8);

                    DownloadVideoJob::dispatchWithRateLimit($videoPart);
                } else {
                    Log::info('Video part file not exists, download video part file disabled', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                }
            } else {
                // 如果文件存在，则更新记录到数据库并写入日志
                Log::info('Video part file already exists', ['id' => $videoPart->cid, 'title' => $videoPart->part]);

                // 时间取文件创建时间
                $videoPart->video_downloaded_at = Carbon::createFromTimestamp(filectime($savePath));

                // 获取相对路径
                $videoPart->video_download_path = get_relative_path($savePath);
                $videoPart->save();
            }

            // 更新視頻已經下載的文件數量
            $video->video_downloaded_num = VideoPart::where('video_id', $video->id)->whereNotNull('video_download_path')->count();
            $video->save();
        }
    }

    public function downloadVideoPartFileQueue(VideoPart $videoPart): void
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

        // 写入cookie到文件
        $cookiePath = storage_path('app/cookie.txt');
        file_put_contents($cookiePath, $this->settingsService->get(SettingKey::COOKIES_CONTENT));

        $url = config('services.bilibili.id_type') == 'bv' ? sprintf('https://www.bilibili.com/video/%s/', $video->bvid) : sprintf('https://www.bilibili.com/video/av%s/', $video->id);

        // 获取音乐节特辑类的链接
        $festivalJumpUrl = $this->bilibiliService->getVideoFestivalJumpUrl($video->id);
        if($festivalJumpUrl){
            $url = $festivalJumpUrl;
        }

        $filePath = $this->getVideoPartValidFilePath($videoPart);
        if ($filePath) {
            $this->updateVideoPartDownloaded($videoPart, $filePath);
            Log::info('video file already exists', ['video_id' => $video->id, 'part' => $videoPart->page, 'filePath' => $filePath]);
            return;
        } else {
            // 如果不存在就去下载
            $this->createVideoDirectory();
            $binPath = base_path('download-video.sh');

            $savePath = $this->buildVideoPartFilePath($videoPart);
            Log::info('download video', [
                'video_id' => $video->id,
                'part'     => $videoPart->page,
                'url'      => escapeshellarg($url),
                'savePath' => escapeshellarg($savePath),
                'binPath'  => escapeshellarg($binPath),
            ]);

            $command = sprintf('%s %s %s %s', $binPath, escapeshellarg($url), escapeshellarg($savePath), escapeshellarg($videoPart->page));
            exec($command, $output, $result);
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
}
