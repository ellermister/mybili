<?php
namespace App\Services;

use App\Models\Video;
use App\Models\VideoPart;
use App\Services\DownloadVideoService;
use Illuminate\Support\Facades\Log;

class HumanReadableNameService
{
    protected string $humanReadableDir;

    public function __construct(
        public DownloadVideoService $downloadVideoService
    ) {
        $this->humanReadableDir = config('app.human_readable_dir');
    }

    /**
     * 生成可读文件名和目录结构
     */
    public function generateHumanReadableNames(): void
    {
        $this->prepareDirectories();
        $this->processVideos();
    }

    /**
     * 准备目录结构
     */
    protected function prepareDirectories(): void
    {
        // 创建主目录
        if (!is_dir($this->humanReadableDir)) {
            // 目录不存在，中止
            throw new \Exception('Human readable directory not found, path:' . $this->humanReadableDir);
        }

        $subDirs = scandir($this->humanReadableDir);

        // 删除子目录
        foreach ($subDirs as $subDir) {
            if ($subDir != "." && $subDir != "..") {
                $path = $this->humanReadableDir . '/' . $subDir;

                if (is_dir($path)) {
                    $this->deleteDirectory($path);
                }
            }
        }
    }

    /**
     * 处理所有视频
     */
    protected function processVideos(): void
    {
        $videos = Video::query()->get();

        foreach ($videos as $video) {
            if ($video['video_downloaded_num'] == 0) {
                continue;
            }

            $this->processVideo($video);
        }
    }

    /**
     * 处理单个视频
     */
    protected function processVideo(Video $video): void
    {
        $parts = VideoPart::query()->where('video_id', $video['id'])->get();
        $videoType = count($parts) == 1 ? 'movies' : 'tvs';
        $sanitizedName = $this->sanitizeFileName($video['title']);

        # get name from favorites and subscriptions
        $favoriteNames = $video->favorite() ? $video->favorite()->pluck('title')->toArray() : [];
        $substriptionNames = $video->subscriptions() ? $video->subscriptions()->pluck('name')->toArray() : [];

        # merge favoriteNames and subscriptionNames
        $allNames = array_unique(array_merge($favoriteNames, $substriptionNames));

        # for each name, create links
        foreach ($allNames as $favoriteName) {
            $sanitizedFavoriteName = $this->sanitizeFileName($favoriteName);

            if ($videoType == 'movies') {
                $this->processMovie($video, $parts->first(), $sanitizedName, $sanitizedFavoriteName);
            } else {
                $this->processTvSeries($video, $parts, $sanitizedName, $sanitizedFavoriteName);
            }
        }
    }

    /**
     * 处理电影类型视频
     */
    protected function processMovie(Video $video, VideoPart $videoPart, string $sanitizedName, string $sanitizedFavoriteName): void
    {
        $videoPath = $this->downloadVideoService->getVideoPartValidFilePath($videoPart);

        if (!is_file($videoPath)) {
            return;
        }

        $this->createVideoLink($videoPath, $sanitizedFavoriteName, $sanitizedName, $sanitizedName . '.mp4');
        $this->createCoverLink($video, $sanitizedFavoriteName, $sanitizedName, $sanitizedName . '.jpg');
    }

    /**
     * 处理电视剧类型视频
     */
    protected function processTvSeries(Video $video, $parts, string $sanitizedName, string $sanitizedFavoriteName): void
    {
        $successCount = 0;
        foreach ($parts as $part) {
            $videoPath = $this->downloadVideoService->getVideoPartValidFilePath($part);

            if (!is_file($videoPath)) {
                continue;
            }

            $episodeName = trim($this->formatEpisodeName($part->part ?? '', $part->page));
            $this->createVideoLink($videoPath, $sanitizedFavoriteName, $sanitizedName, $episodeName . '.mp4');
            $this->createCoverLink($video, $sanitizedFavoriteName, $sanitizedName, $episodeName . '.jpg');
            $successCount++;
        }

        if ($successCount >= 1) {
            $this->createCoverLink($video, $sanitizedFavoriteName, $sanitizedName, $sanitizedName . '.jpg');
        }
    }

    /**
     * 创建视频硬链接
     */
    protected function createVideoLink(string $sourcePath, string $sanitizedFavoriteName, string $seriesName, string $fileName): void
    {
        $targetPath = $this->buildTargetPath($sanitizedFavoriteName, $seriesName, $fileName);

        if (is_file($targetPath)) {
            unlink($targetPath);
        }

        $this->createLink($sourcePath, $targetPath);
    }

    /**
     * 创建封面硬链接
     */
    protected function createCoverLink(Video $video, string $sanitizedFavoriteName, string $seriesName, string $fileName): void
    {
        $coverPath = storage_path('app/public/' . $video->cover_info->path);

        if (!is_file($coverPath)) {
            return;
        }

        $targetPath = $this->buildTargetPath($sanitizedFavoriteName, $seriesName, $fileName);

        if (is_file($targetPath)) {
            unlink($targetPath);
        }

        $this->createLink($coverPath, $targetPath);
    }

    protected function createLink(string $sourcePath, string $targetPath): void
    {
        try {
            if (link($sourcePath, $targetPath)) {
                Log::info("Created link: {$targetPath}");
            }
        } catch (\Throwable $e) {
            // php link(): No error information
            // 很奇怪的一种错误，尝试用过 shell 创建硬链接
            $sourcePath = escapeshellarg($sourcePath);
            $targetPath = escapeshellarg($targetPath);
            exec("ln {$sourcePath} {$targetPath}", $output, $returnCode);

            if ($returnCode === 0 && is_file($targetPath)) {
                Log::info("Created link: {$targetPath}");
            } else {
                Log::error("createLink error: {$e->getMessage()}", [
                    'output' => $output,
                    'returnCode' => $returnCode,
                ]);
            }
        }
    }

    /**
     * 构建目标路径
     */
    protected function buildTargetPath(string $sanitizedFavoriteName, string $seriesName, string $fileName): string
    {
        $targetPath = $this->humanReadableDir . '/' . $sanitizedFavoriteName . '/' . $seriesName . '/' . $fileName;
        $dir = dirname($targetPath);

        if (!is_dir($dir)) {
            mkdir($dir, 0777, true);
        }

        return $targetPath;
    }

    /**
     * 格式化剧集名称
     */
    protected function formatEpisodeName(string $seriesName, int $episode): string
    {
        return $seriesName . ' S00E' . str_pad($episode, 2, '0', STR_PAD_LEFT);
    }

    /**
     * 清理文件名中的非法字符
     */
    protected function sanitizeFileName(string $name): string
    {
        // 移除Windows和Linux文件系统中的非法字符
        $illegalChars = [
            '/', '\\', ':', '*', '?', '"', '<', '>', '|', // Windows非法字符
            '\0',                                         // Null字节
            '#', '%', '&', '{', '}',                      // 可能在某些系统中有问题的字符
            '`', '$', '!', '@', '+', '=',                 // 特殊字符
        ];

        // 替换非法字符为空格
        $name = str_replace($illegalChars, ' ', $name);

        // 移除控制字符
        $name = preg_replace('/[\x00-\x1F\x7F]/', '', $name);

        // 合并多个空格
        $name = preg_replace('/\s+/', ' ', $name);

        // 移除首尾空格
        $name = trim($name);

        // 限制文件名长度
        $maxLength = 80;
        if (strlen($name) > $maxLength) {
            $name = mb_substr($name, 0, $maxLength);
        }

        return empty($name) ? 'untitled' : $name;
    }

    /**
     * 递归删除目录及其所有内容
     */
    protected function deleteDirectory(string $dir): bool
    {
        if (!is_dir($dir)) {
            return false;
        }

        $files = array_diff(scandir($dir), ['.', '..']);

        foreach ($files as $file) {
            $filePath = $dir . '/' . $file;

            if (is_dir($filePath)) {
                $this->deleteDirectory($filePath);
            } else {
                unlink($filePath);
            }
        }

        return rmdir($dir);
    }
}
