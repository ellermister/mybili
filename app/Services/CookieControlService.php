<?php

namespace App\Services;

use App\Jobs\SendCookieExpiredMessageJob;
use Illuminate\Support\Facades\Redis;
use Log;

class CookieControlService
{
    private const COOKIE_EXPIRED_NOTIFICATION_KEY = 'cookie:expired:notification:sent';

    /**
     * 检查并发送 Cookie 过期通知
     * 如果 Redis 中不存在标记 key，则发送通知并设置标记
     * 
     * @return bool 是否发送了通知
     */
    public function checkAndNotifyCookieExpired(): bool
    {
        // 检查 Redis 中是否已存在标记
        if (Redis::exists(self::COOKIE_EXPIRED_NOTIFICATION_KEY)) {
            Log::info('Cookie expired notification already sent, skipping');
            return false;
        }

        // 发送通知
        dispatch(new SendCookieExpiredMessageJob());
        
        // 设置标记（24小时过期，避免永久占用）
        Redis::setex(self::COOKIE_EXPIRED_NOTIFICATION_KEY, 86400, 1);
        
        Log::info('Cookie expired notification sent');
        return true;
    }

    /**
     * 清除 Cookie 过期通知标记
     * 在 Cookie 更新成功后调用，避免重复通知
     * 
     * @return bool
     */
    public function clearCookieExpiredNotification(): bool
    {
        $result = Redis::del(self::COOKIE_EXPIRED_NOTIFICATION_KEY);
        Log::info('Cookie expired notification flag cleared', ['deleted' => $result]);
        return $result > 0;
    }

    /**
     * 检查是否已发送过 Cookie 过期通知
     * 
     * @return bool
     */
    public function hasNotifiedCookieExpired(): bool
    {
        return Redis::exists(self::COOKIE_EXPIRED_NOTIFICATION_KEY) > 0;
    }
}

