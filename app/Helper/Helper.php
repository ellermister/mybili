<?php

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

function match_cookie_main()
{
    $cookiPath = storage_path('app/cookie.txt');
    if (!is_file($cookiPath)) {
        throw new \LogicException("cookie 文件不存在");
    }
    $result   = file_get_contents($cookiPath);
    $SESSDATA = $DedeUserID = '';
    if (preg_match('/SESSDATA\t([\S]+)/', $result, $matches)) {
        $SESSDATA = $matches[1];
    }
    if (preg_match('/DedeUserID\t([\S]+)/', $result, $matches)) {
        $DedeUserID = $matches[1];
    }

    if ($SESSDATA && $DedeUserID) {
        return [$SESSDATA, $DedeUserID];
    } else {
        throw new \LogicException("未能够从cookie中提取密钥信息");
    }
}

function parse_netscape_cookie_file($filename) {

    if(!is_file($filename)){
        throw new ErrorException("cookie.txt 不存在");
    }

    $cookies = [];
    $lines = file($filename, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);

    foreach ($lines as $line) {
        if ($line[0] === '#' || substr_count($line, "\t") < 6) {
            continue; // Skip comments and invalid lines
        }

        list($domain, $flag, $path, $secure, $expiry, $name, $value) = explode("\t", $line);

        $cookies[] = new SetCookie([
            'Domain'   => $domain,
            'Path'     => $path,
            'Name'     => $name,
            'Value'    => $value,
            'Expires'  => $expiry,
            'Secure'   => ($secure === 'TRUE'),
        ]);
    }

    return new CookieJar(false, $cookies);
}
