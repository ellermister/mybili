<?php

namespace App\Contracts;

use App\Models\Video;
use App\Models\VideoPart;

interface VideoManagerServiceInterface
{
    // ==================== 文件系统相关 ====================
    /**
     * 获取图片存储目录，如果不存在则创建
     */
    // public function getImagesDirIfNotExistCreate(): string;

    /**
     * 创建视频存储目录
     */
    public function createVideoDirectory(): void;

    /**
     * 获取视频下载路径
     */
    public function getVideoDownloadPath(string $id, int $part = 1): string;

    /**
     * 获取视频下载哈希文件路径
     */
    public function getVideoDownloadHashPath(string $id, int $part = 1): string;

    // ==================== 视频信息相关 ====================
    /**
     * 获取视频信息
     */
    public function getVideoInfo(string $id): ?Video;

    /**
     * 获取视频文件哈希值
     */
    public function getVideoFileHash(string $id): ?string;

    /**
     * 检查视频是否已下载
     */
    public function videoDownloaded(string $id): bool;

    /**
     * 检查视频文件是否存在
     */
    public function hasVideoFile(string $id, int $part = 1): bool;


    /**
     * 获取所有视频
     */
    public function getVideos(array $conditions = []):array;

    /**
     * 获取视频统计信息
     */
    public function getVideosStat(array $conditions = []): array;

    // ==================== 视频分P相关 ====================
    /**
     * 获取视频所有分P信息
     */
    public function getAllPartsVideo(string $id): array;

    /**
     * 获取用户可访问的视频分P信息
     */
    public function getAllPartsVideoForUser(string $id, int $parts = 1): array;

    // ==================== 收藏夹相关 ====================
    /**
     * 获取收藏夹视频列表
     */
    public function getVideoListByFav(int $favId): array;

    /**
     * 获取收藏夹列表
     */
    public function getFavList(): array;

    /**
     * 更新收藏夹列表
     */
    public function updateFavList(): void;

    /**
     * 更新收藏夹视频列表
     */
    public function updateFavVideos(int $favId): void;

    /**
     * 更新视频分P信息
     */
    public function updateVideoParts(Video $video): void;

    // ==================== 弹幕相关 ====================
    /**
     * 获取弹幕信息
     */
    public function getDanmaku(string $cid): array;

    /**
     * 保存弹幕数据
     */
    public function saveDanmaku(string $cid, array $danmaku): void;

    /**
     * 获取弹幕下载时间
     */
    public function danmakuDownloadedTime(string $avId): ?int;

    // ==================== 下载任务相关 ====================
    /**
     * 分发视频下载任务
     */
    public function dispatchDownloadVideoJob(array $video): void;

    /**
     * 完成视频下载
     */
    public function finishDownloadVideo(string $id): void;

    /**
     * 下载视频分P文件
     */
    public function downloadVideoPartFile(VideoPart $videoPart, bool $onlyLocalCheck = false): void;

    /**
     * 分发弹幕下载任务
     */
    public function dispatchDownloadDanmakuJob(int $avId): void;


    /**
     * 下载弹幕数据
     */
    public function downloadDanmaku(VideoPart $videoPart):void;

    // ==================== 视频状态管理 ====================
    /**
     * 标记视频已下载
     */
    public function setVideoDownloaded(string $avId): void;

    /**
     * 取消视频下载标记
     */
    public function delVideoDownloaded(string $avId): void;

    /**
     * 标记弹幕已下载
     */
    public function setDanmakuDownloadedTime(string $avId): void;

    /**
     * 检查视频是否无效
     */
    public function videoIsInvalid(array $video): bool;
}