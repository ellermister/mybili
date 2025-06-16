<?php

namespace App\Contracts;

use App\Models\VideoPart;
use App\Models\Video;

interface VideoDownloadServiceInterface
{
    /**
     * 判断视频文件是否存在
     */
    public function hasVideoFile(VideoPart $videoPart): bool;

    /**
     * 获得有效的视频文件路径
     */
    public function getVideoPartValidFilePath(VideoPart $videoPart): ?string;

    /**
     * 构建视频文件路径
     */
    public function buildVideoPartFilePath(VideoPart $videoPart): string;

    /**
     * 更新视频分P下载状态
     */
    public function updateVideoPartDownloaded(VideoPart $videoPart, string $savePath): void;

    /**
     * 下载视频分P文件
     */
    public function downloadVideoPartFile(VideoPart $videoPart, bool $tryDownload = false): void;

    /**
     * 下载视频分P文件队列
     */
    public function downloadVideoPartFileQueue(VideoPart $videoPart): void;
}