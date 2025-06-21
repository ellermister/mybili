<?php

namespace App\Services;

use App\Contracts\VideoManagerServiceInterface;
use App\Contracts\VideoDownloadServiceInterface;
use App\Models\Video;
use App\Models\VideoPart;
use Illuminate\Support\Facades\Log;

class HumanReadableNameService
{
    protected VideoDownloadServiceInterface $videoDownloadService;
    protected VideoManagerServiceInterface $videoManagerService;
    protected string $humanReadableDir;
    protected array $subDirs = ['tvs', 'movies'];

    public function __construct(
        VideoDownloadServiceInterface $videoDownloadService,
        VideoManagerServiceInterface $videoManagerService
    ) {
        $this->videoDownloadService = $videoDownloadService;
        $this->videoManagerService = $videoManagerService;
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
        if(!is_dir($this->humanReadableDir)){
            // 目录不存在，中止
            throw new \Exception('Human readable directory not found, path:'.$this->humanReadableDir);
        }

        // 创建子目录
        foreach ($this->subDirs as $subDir) {
            $path = $this->humanReadableDir . '/' . $subDir;
            
            // 删除重建目录
            if (is_dir($path)) {
                $this->deleteDirectory($path);
            }
            
            mkdir($path, 0777, true);
        }
    }

    /**
     * 处理所有视频
     */
    protected function processVideos(): void
    {
        $videos = $this->videoManagerService->getVideos();
        
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
        $parts = $this->videoManagerService->getAllPartsVideo($video['id']);
        $videoType = count($parts) == 1 ? 'movies' : 'tvs';
        $sanitizedName = $this->sanitizeFileName($video['title']);

        if ($videoType == 'movies') {
            $this->processMovie($video, $parts->first(), $sanitizedName);
        } else {
            $this->processTvSeries($video, $parts, $sanitizedName);
        }
    }

    /**
     * 处理电影类型视频
     */
    protected function processMovie(Video $video, VideoPart $videoPart, string $sanitizedName): void
    {
        $videoPath = $this->videoDownloadService->getVideoPartValidFilePath($videoPart);
        
        if (!is_file($videoPath)) {
            return;
        }

        $this->createVideoLink($videoPath, 'movies', $sanitizedName, $sanitizedName . '.mp4');
        $this->createCoverLink($video, 'movies', $sanitizedName, $sanitizedName . '.jpg');
    }

    /**
     * 处理电视剧类型视频
     */
    protected function processTvSeries(Video $video, $parts, string $sanitizedName): void
    {
        foreach ($parts as $part) {
            $videoPath = $this->videoDownloadService->getVideoPartValidFilePath($part);
            
            if (!is_file($videoPath)) {
                continue;
            }

            $episodeName = $this->formatEpisodeName($sanitizedName, $part->page);
            $this->createVideoLink($videoPath, 'tvs', $sanitizedName, $episodeName . '.mp4');
            $this->createCoverLink($video, 'tvs', $sanitizedName, $episodeName . '.jpg');
        }
    }

    /**
     * 创建视频硬链接
     */
    protected function createVideoLink(string $sourcePath, string $videoType, string $seriesName, string $fileName): void
    {
        $targetPath = $this->buildTargetPath($videoType, $seriesName, $fileName);
        
        if (is_file($targetPath)) {
            unlink($targetPath);
        }
        
        if (link($sourcePath, $targetPath)) {
            Log::info("Created video link: {$targetPath}");
        }
    }

    /**
     * 创建封面硬链接
     */
    protected function createCoverLink(Video $video, string $videoType, string $seriesName, string $fileName): void
    {
        $coverPath = storage_path('app/public/' . $video->cache_image);
        
        if (!is_file($coverPath)) {
            return;
        }

        $targetPath = $this->buildTargetPath($videoType, $seriesName, $fileName);
        
        if (is_file($targetPath)) {
            unlink($targetPath);
        }
        
        if (link($coverPath, $targetPath)) {
            Log::info("Created cover link: {$targetPath}");
        }
    }

    /**
     * 构建目标路径
     */
    protected function buildTargetPath(string $videoType, string $seriesName, string $fileName): string
    {
        $targetPath = $this->humanReadableDir . '/' . $videoType . '/' . $seriesName . '/' . $fileName;
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

        $files = array_diff(scandir($dir), array('.', '..'));
        
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