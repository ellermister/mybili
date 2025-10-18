<?php
namespace App\Services;

use App\Models\VideoPart;
use App\Services\BilibiliService;
use App\Services\BilibiliSuspendService;
use App\Services\DownloadFilterService;
use App\Services\SettingsService;

class DownloadVideoService
{
    public string $downloadFolder = 'videos';

    public function __construct(
        public DownloadFilterService $downloadFilterService,
        public SettingsService $settingsService,
        public BilibiliService $bilibiliService,
        public BilibiliSuspendService $bilibiliSuspendService
    ) {
        $this->createDownloadDirectory();
    }

    public function createDownloadDirectory(): void
    {
        $path = storage_path('app/public/' . $this->downloadFolder);
        if (! is_dir($path)) {
            mkdir($path, 0644);
            if (! is_dir($path)) {
                throw new \Exception("下载路径不存在, 创建失败");
            }
        }
    }

    protected function getDownloadPath(string $id, int $part = 1): string
    {

        $path = storage_path('app/public/' . $this->downloadFolder);
        if ($part === 1) {
            return sprintf('%s/%s.mp4', $path, $id);
        }
        return sprintf('%s/%s.part%d.mp4', $path, $id, $part);
    }

    public function getDownloadHashPath(string $id, int $part = 1): string
    {
        $path = storage_path('app/public/' . $this->downloadFolder);
        if ($part === 1) {
            return sprintf('%s/%s.mp4.hash', $path, $id);
        }
        return sprintf('%s/%s.part%d.mp4.hash', $path, $id, $part);
    }

    public function getDownloadHash(string $savePath): ?string
    {
        $hashPath = sprintf('%s.hash', $savePath);
        if (is_file($hashPath)) {
            return trim(file_get_contents($hashPath));
        }
        return null;
    }

    public function saveDownloadHash(string $savePath, string $hash): void
    {
        $hashPath = sprintf('%s.hash', $savePath);
        file_put_contents($hashPath, $hash);
    }

    public function getVideoPartValidFilePath(VideoPart $videoPart): ?string
    {

        $savePath = $this->getDownloadPath($videoPart->video_id, $videoPart->page);
        $hashPath = $this->getDownloadHashPath($videoPart->video_id, $videoPart->page);
        if (is_file($savePath) && is_file($hashPath)) {
            return $savePath;
        }

        $savePath = $this->getDownloadPath($videoPart->video_id, $videoPart->cid);
        $hashPath = $this->getDownloadHashPath($videoPart->video_id, $videoPart->cid);
        if (is_file($savePath) && is_file($hashPath)) {
            return $savePath;
        }

        return null;
    }

    public function getOldDownloadPath(string $id, int $part = 1): string
    {

        $path = storage_path('app/public/' . $this->downloadFolder);
        if ($part === 1) {
            return sprintf('%s/%s.mp4', $path, $id);
        }
        return sprintf('%s/%s.part%d.mp4', $path, $id, $part);
    }

    public function buildVideoPartFilePath(VideoPart $videoPart): string
    {
        return $this->getDownloadPath($videoPart->video_id, $videoPart->cid);
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

    public function deleteVideoPartFile(VideoPart $videoPart): void
    {
        $savePath = $this->getDownloadPath($videoPart->video_id, $videoPart->page);
        $hashPath = $this->getDownloadHashPath($videoPart->video_id, $videoPart->page);
        if (is_file($savePath)) {
            unlink($savePath);
        }
        if (is_file($hashPath)) {
            unlink($hashPath);
        }

        $savePath = $this->getDownloadPath($videoPart->video_id, $videoPart->cid);
        $hashPath = $this->getDownloadHashPath($videoPart->video_id, $videoPart->cid);
        if (is_file($savePath)) {
            unlink($savePath);
        }
        if (is_file($hashPath)) {
            unlink($hashPath);
        }
    }
}
