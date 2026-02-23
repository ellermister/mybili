<?php

return [

    /*
    |--------------------------------------------------------------------------
    | Third Party Services
    |--------------------------------------------------------------------------
    |
    | This file is for storing the credentials for third party services such
    | as Mailgun, Postmark, AWS and more. This file provides the de facto
    | location for this type of information, allowing packages to have
    | a conventional file to locate the various service credentials.
    |
    */

    'postmark' => [
        'token' => env('POSTMARK_TOKEN'),
    ],

    'ses'      => [
        'key'    => env('AWS_ACCESS_KEY_ID'),
        'secret' => env('AWS_SECRET_ACCESS_KEY'),
        'region' => env('AWS_DEFAULT_REGION', 'us-east-1'),
    ],

    'resend'   => [
        'key' => env('RESEND_KEY'),
    ],

    'slack'    => [
        'notifications' => [
            'bot_user_oauth_token' => env('SLACK_BOT_USER_OAUTH_TOKEN'),
            'channel'              => env('SLACK_BOT_USER_DEFAULT_CHANNEL'),
        ],
    ],

    'bilibili' => [
        // 每页视频数量，最大40
        'fav_videos_page_size'     => env('BILIBILI_FAV_VIDEOS_PAGE_SIZE', 40),
        'id_type'                  => env('BILIBILI_ID_TYPE', 'bv'),

        // 控制下载任务派发时计算的槽位，也就是并发数，一般用这个，后面考虑是否合并配置
        'download_concurrency'     => env('BILIBILI_DOWNLOAD_CONCURRENCY', 10),

        // 控制下载任务执行时检查频率的参数，用于检查n秒内是否执行过多，保证执行时不会超过最大(风控)限制
        'limit_download_video_job' => env('BILIBILI_LIMIT_DOWNLOAD_VIDEO_JOB', 2),

        'ignore_cookies'           => env('BILIBILI_IGNORE_COOKIES', false),

        'setting_read_only'        => env('BILIBILI_SETTING_READ_ONLY', false),
    ],

];
