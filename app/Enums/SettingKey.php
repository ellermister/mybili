<?php
namespace App\Enums;

enum SettingKey: string {
    case FAVORITE_SYNC_ENABLED            = 'favorite_sync_enabled';
    case VIDEO_DOWNLOAD_ENABLED           = 'video_download_enabled';
    case DANMAKU_DOWNLOAD_ENABLED         = 'danmaku_download_enabled';
    case MULTI_PARTITION_DOWNLOAD_ENABLED = 'multi_partition_download_enabled';
    case HUMAN_READABLE_NAME_ENABLED = 'human_readable_name_enabled';

    case USAGE_ANALYTICS_ENABLED = 'usage_analytics_enabled';

    case FAVORITE_EXCLUDE = 'fav_exclude';
    case NAME_EXCLUDE     = 'name_exclude';
    case SIZE_EXCLUDE     = 'size_exclude';
    case DURATION_VIDEO_EXCLUDE = 'duration_video_exclude';
    case DURATION_VIDEO_PART_EXCLUDE = 'duration_video_part_exclude';


    case INSTALLED_DATETIME = 'installed_datetime';
    case COOKIES_CONTENT = 'cookies_content';

    // Telegram Bot 设置
    case TELEGRAM_BOT_API_URL = 'telegram_bot_api_url';
    case TELEGRAM_BOT_ENABLED = 'telegram_bot_enabled';
    case TELEGRAM_BOT_TOKEN = 'telegram_bot_token';
    case TELEGRAM_CHAT_ID = 'telegram_chat_id';

    public function label(): string
    {
        return match ($this) {
            self::FAVORITE_SYNC_ENABLED => '收藏夹同步',
            self::VIDEO_DOWNLOAD_ENABLED => '视频下载',
            self::DANMAKU_DOWNLOAD_ENABLED => '弹幕下载',
            self::MULTI_PARTITION_DOWNLOAD_ENABLED => '多分P下载',
            self::HUMAN_READABLE_NAME_ENABLED => '可读文件名',
            self::USAGE_ANALYTICS_ENABLED => '使用情况统计',

            self::FAVORITE_EXCLUDE => '收藏夹排除',
            self::NAME_EXCLUDE => '名称排除',
            self::SIZE_EXCLUDE => '大小排除',
            self::DURATION_VIDEO_EXCLUDE => '视频时长排除',
            self::DURATION_VIDEO_PART_EXCLUDE => '分P时长排除',

            self::INSTALLED_DATETIME => '安装日期时间',
            self::COOKIES_CONTENT => 'Cookies 内容',
            
            // Telegram Bot 设置
            self::TELEGRAM_BOT_API_URL => 'Telegram Bot API URL',
            self::TELEGRAM_BOT_ENABLED => 'Telegram Bot 启用',
            self::TELEGRAM_BOT_TOKEN => 'Telegram Bot Token',
            self::TELEGRAM_CHAT_ID => 'Telegram 聊天 ID',
        };
    }
}
