<?php

namespace App\Contracts;

use App\Models\FavoriteList;
use App\Models\Video;
use App\Models\VideoPart;

interface VideoManagerServiceInterface
{

    // ==================== 视频信息相关 ====================
    /**
     * 获取视频信息
     */
    public function getVideoInfo(string $id, bool $withParts = false): ?Video;

    public function getVideoDanmakuCount(Video $video): int;


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
    public function getAllPartsVideoForUser(Video $video): array;

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
     * 获取收藏夹详情
     */
    public function getFavDetail(int $favId, array $columns = ['*']): ?FavoriteList;

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
     * 下载弹幕数据
     */
    public function downloadDanmaku(VideoPart $videoPart):void;

    // ==================== 视频状态管理 ====================


    /**
     * 检查视频是否无效
     */
    public function videoIsInvalid(array $video): bool;
}