<?php
namespace App\Enums;

enum SettingKey: string {
    case FAVORITE_SYNC_ENABLED            = 'favorite_sync_enabled';
    case VIDEO_DOWNLOAD_ENABLED           = 'video_download_enabled';
    case DANMAKU_DOWNLOAD_ENABLED         = 'danmaku_download_enabled';
    case MULTI_PARTITION_DOWNLOAD_ENABLED = 'multi_partition_download_enabled';

    case FAVORITE_EXCLUDE = 'fav_exclude';
    case NAME_EXCLUDE     = 'name_exclude';
    case SIZE_EXCLUDE     = 'size_exclude';

    public function label(): string
    {
        return match ($this) {
            self::FAVORITE_SYNC_ENABLED => '收藏夹同步',
            self::VIDEO_DOWNLOAD_ENABLED => '视频下载',
            self::DANMAKU_DOWNLOAD_ENABLED => '弹幕下载',
            self::MULTI_PARTITION_DOWNLOAD_ENABLED => '多分P下载',

            self::FAVORITE_EXCLUDE => '收藏夹排除',
            self::NAME_EXCLUDE => '名称排除',
            self::SIZE_EXCLUDE => '大小排除',
        };
    }
}
