<?php

use App\Enums\SettingKey;
use App\Services\SettingsService;
use GuzzleHttp\Cookie\CookieJar;
use GuzzleHttp\Cookie\SetCookie;

/**
 * Get redis cache instance
 *
 * @return \Illuminate\Contracts\Cache\Repository
 */
function cache_redis(): \Illuminate\Contracts\Cache\Repository
{
    return Cache::store('redis');
}

/**
 * Get redis instance
 *
 * @return mixed|Redis
 */
function redis()
{
    return \Illuminate\Support\Facades\Redis::connection()->client();
}

function video_has_invalid(array $videoInfo)
{
    if ($videoInfo['title'] == '已失效视频') {
        return true;
    }

    if ($videoInfo['attr'] > 0) {
        return true;
    }

    return false;
}

function parse_netscape_cookie_file($filename)
{

    if (! is_file($filename)) {
        throw new ErrorException("cookie.txt 不存在");
    }

    $cookies = [];
    $lines   = file($filename, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);

    foreach ($lines as $line) {
        if ($line[0] === '#' || substr_count($line, "\t") < 6) {
            continue; // Skip comments and invalid lines
        }

        list($domain, $flag, $path, $secure, $expiry, $name, $value) = explode("\t", $line);

        $cookies[] = new SetCookie([
            'Domain'  => $domain,
            'Path'    => $path,
            'Name'    => $name,
            'Value'   => $value,
            'Expires' => $expiry,
            'Secure'  => ($secure === 'TRUE'),
        ]);
    }

    return new CookieJar(false, $cookies);
}


function parse_netscape_cookie_content($content)
{
    if (empty($content)) {
        if (config('services.bilibili.ignore_cookies')) {
            return new CookieJar(false, []);
        }
        throw new \InvalidArgumentException("Cookie 内容不能为空");
    }

    $cookies = [];
    $lines = explode("\n", $content);

    foreach ($lines as $line) {
        $line = trim($line);
        if (empty($line) || ($line[0] ?? '') === '#' || substr_count($line, "\t") < 6) {
            continue; // 跳过注释和无效行
        }

        $parts = explode("\t", $line);
        if (count($parts) < 7) {
            continue; // 确保有足够的字段
        }

        list($domain, $flag, $path, $secure, $expiry, $name, $value) = $parts;

        $cookies[] = new SetCookie([
            'Domain'  => $domain,
            'Path'    => $path,
            'Name'    => $name,
            'Value'   => $value,
            'Expires' => $expiry,
            'Secure'  => ($secure === 'TRUE'),
        ]);
    }

    return new CookieJar(false, $cookies);
}

function get_relative_path(string $absolutePath): string
{
    // 如果是 storage 路径，转换为 URL
    if (Str::startsWith($absolutePath, storage_path('app/public'))) {
        return Str::after($absolutePath, storage_path('app/public/'));
    }

    // 如果是 public 路径，转换为相对路径
    if (Str::startsWith($absolutePath, public_path())) {
        return Str::after($absolutePath, public_path());
    }

    return $absolutePath;
}


function usage_analytics_enabled(): bool
{
    return app(SettingsService::class)->get(SettingKey::USAGE_ANALYTICS_ENABLED->value) != 'off';
}


/**
 * Convert bytes to human readable file size
 * Maximum unit is GB, minimum unit is MB
 */
function format_file_size(int $bytes): string
{
    if ($bytes === 0) {
        return '0 MB';
    }

    $gb = $bytes / (1024 * 1024 * 1024);
    if ($gb >= 1) {
        return round($gb, 2) . ' GB';
    }

    $mb = $bytes / (1024 * 1024);
    return round($mb, 2) . ' MB';
}

/**
 * Format duration from seconds to readable format (HH:MM:SS or MM:SS)
 */
function format_duration(string $duration): string
{
    // Convert duration to integer (assuming it's in seconds)
    return gmdate('H:i:s', intval($duration));
}