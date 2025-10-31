<?php
namespace App\Services;

use App\Enums\SettingKey;
use App\Services\SettingsService;
use Arr;
use GuzzleHttp\Client;
use Log;

class BilibiliService
{

    const API_HOST     = 'https://api.bilibili.com';
    const APP_API_HOST = 'https://app.biliapi.com';
    const APP_KEY      = '1d8b6e7d45233436';
    const APP_SECRET   = '560c52ccd288fed045859ed18bffd973';

    protected $favVideosPageSize;

    public function __construct(
        public SettingsService $settingsService,
        public BilibiliSuspendService $bilibiliSuspendService
    ) {
        $this->favVideosPageSize = intval(config('services.bilibili.fav_videos_page_size'));
    }

    private function getClient()
    {
        $cookies = parse_netscape_cookie_content($this->settingsService->get(SettingKey::COOKIES_CONTENT));
        return new Client([
            'cookies' => $cookies,
            'headers' => [
                'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3',
                'Referer'    => 'https://www.bilibili.com/',
            ],
        ]);
    }

    private function getAppClient()
    {
        return new Client([
            'headers' => [
                'User-Agent' => 'bili-universal/77100100 CFNetwork/1404.0.5 Darwin/22.3.0',
                'Buvid'      => 'XY' . bin2hex(random_bytes(16)),
                'Display-ID' => 'MAC-' . strtoupper(bin2hex(random_bytes(6))),
                'Device-ID'  => bin2hex(random_bytes(20)),
                'Channel'    => 'AppStore',
                'App-Key'    => 'iphone',
                'Env'        => 'prod',
            ],
        ]);
    }

    public function getDanmaku(int $cid, int $duration)
    {
        $cookies = parse_netscape_cookie_content($this->settingsService->get(SettingKey::COOKIES_CONTENT));
        $client  = $this->getClient();

        // 获取 WBI keys
        $navResponse = $client->request('GET', self::API_HOST . '/x/web-interface/nav', [
            'cookies' => $cookies,
        ]);
        $navData = json_decode($navResponse->getBody()->getContents(), true);

        if (! isset($navData['data']['wbi_img'])) {
            throw new \Exception('无法获取 WBI keys');
        }

        $imgUrl = $navData['data']['wbi_img']['img_url'];
        $subUrl = $navData['data']['wbi_img']['sub_url'];
        $imgKey = substr($imgUrl, strrpos($imgUrl, '/') + 1, -4);
        $subKey = substr($subUrl, strrpos($subUrl, '/') + 1, -4);

        $segmentCount = ceil($duration / 360);
        $danmakus     = [];
        for ($i = 1; $i <= $segmentCount; $i++) {
            // 准备参数
            $params = [
                'type'          => 1,
                'oid'           => $cid,
                'segment_index' => $i,
                'web_location'  => 1315873,
                'wts'           => time(),
            ];

            // 生成 WBI 签名
            $query = $this->encWbi($params, $imgKey, $subKey);

            // 请求弹幕数据
            $url      = self::API_HOST . "/x/v2/dm/wbi/web/seg.so?" . $query;
            $response = $client->request('GET', $url, [
                'cookies' => $cookies,
                'headers' => [
                    'User-Agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3',
                    'Referer'    => 'https://www.bilibili.com/',
                ],
            ]);
            $content         = $response->getBody()->getContents();
            $currentDanmakus = $this->parseDanmakuProtobuf($content);
            $danmakus        = array_merge($danmakus, $currentDanmakus);
        }
        return array_map(function ($danmaku) {
            // [
            //     "id" => 1620000000000000000000000000
            //     "progress" => 390379
            //     "mode" => 1
            //     "fontsize" => 25
            //     "color" => 16777215
            //     "midHash" => b"\x010È\x01\x01Ð\x01Æ¡´“\x03Ø\x01\x01"
            //     "content" => "是我"
            //     "ctime" => 1720936142
            //     "weight" => 10
            //     "action" => "1620000000000000000000000000"
            // ]
            return Arr::only($danmaku, ['content', 'mode', 'color', 'progress', 'id']);
        }, $danmakus);
        // 手动解析二进制数据
    }

    private function encWbi($params, $imgKey, $subKey)
    {
        // WBI 签名算法
        $mixinKeyEncTab = [
            46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
            33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
            61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
            36, 20, 34, 44, 52,
        ];

        // 获取混合密钥
        $orig     = $imgKey . $subKey;
        $mixinKey = '';
        foreach ($mixinKeyEncTab as $n) {
            if (isset($orig[$n])) {
                $mixinKey .= $orig[$n];
            }
        }
        $mixinKey = substr($mixinKey, 0, 32);

        // 过滤参数
        $filteredParams = [];
        foreach ($params as $key => $value) {
            $filteredParams[$key] = preg_replace("/[!'()*]/", '', (string) $value);
        }

        // 按键排序
        ksort($filteredParams);

        // 构建查询字符串
        $query = http_build_query($filteredParams);

        // 计算 MD5
        $wbiSign = md5($query . $mixinKey);

        return $query . '&w_rid=' . $wbiSign;
    }

    /**
     * APP签名方法
     */
    private function appSign(array $params, string $appkey, string $appsec): array
    {
        // 添加appkey参数
        $params['appkey'] = $appkey;

        // 按照key重排参数
        ksort($params);

        // 序列化参数
        $query = http_build_query($params);

        // 计算api签名
        $sign = md5($query . $appsec);

        // 添加签名参数
        $params['sign'] = $sign;

        return $params;
    }

    /**
     * 解析弹幕 protobuf 数据
     */
    private function parseDanmakuProtobuf($binary)
    {
        $danmakus = [];
        $pos      = 0;
        $length   = strlen($binary);

        while ($pos < $length) {
            // 添加边界检查
            if ($pos >= $length) {
                break;
            }

            // 读取字段标识和类型
            $byte        = ord($binary[$pos]);
            $fieldNumber = $byte >> 3;
            $wireType    = $byte & 0x07;
            $pos++;

            // 如果是弹幕数组字段 (field number = 1)
            if ($fieldNumber === 1) {
                // 读取长度
                $msgLen = 0;
                $shift  = 0;
                do {
                    // 添加边界检查
                    if ($pos >= $length) {
                        break 2; // 跳出外层循环
                    }
                    $byte = ord($binary[$pos]);
                    $msgLen |= ($byte & 0x7F) << $shift;
                    $shift += 7;
                    $pos++;
                } while ($byte & 0x80);

                // 添加长度检查
                if ($pos + $msgLen > $length) {
                    break;
                }

                // 解析单条弹幕
                $danmaku = $this->parseSingleDanmaku(substr($binary, $pos, $msgLen));
                if ($danmaku) {
                    $danmakus[] = $danmaku;
                }
                $pos += $msgLen;
            } else {
                // 跳过其他字段
                switch ($wireType) {
                    case 0: // Varint
                        while ($pos < $length && (ord($binary[$pos++]) & 0x80));
                        break;
                    case 1: // 64-bit
                        $pos += 8;
                        if ($pos > $length) {
                            break 2;
                        }

                        break;
                    case 2: // Length-delimited
                        $len   = 0;
                        $shift = 0;
                        do {
                            if ($pos >= $length) {
                                break 3;
                            }
                            // 跳出所有循环
                            $byte = ord($binary[$pos]);
                            $len |= ($byte & 0x7F) << $shift;
                            $shift += 7;
                            $pos++;
                        } while ($byte & 0x80);
                        $pos += $len;
                        if ($pos > $length) {
                            break 2;
                        }

                        break;
                    case 5: // 32-bit
                        $pos += 4;
                        if ($pos > $length) {
                            break 2;
                        }

                        break;
                }
            }
        }

        return $danmakus;
    }

    /**
     * 解析单条弹幕数据
     */
    private function parseSingleDanmaku($binary)
    {
        $result = [];
        $pos    = 0;
        $length = strlen($binary);

        while ($pos < $length) {
            // 读取字段标识和类型
            $byte        = ord($binary[$pos]);
            $fieldNumber = $byte >> 3;
            $wireType    = $byte & 0x07;
            $pos++;

            // 根据字段编号解析对应的值
            switch ($fieldNumber) {
                case 1:  // id
                case 2:  // progress
                case 3:  // mode
                case 4:  // fontsize
                case 5:  // color
                case 8:  // ctime
                case 9:  // weight
                case 11: // attr
                    $value = 0;
                    $shift = 0;
                    do {
                        $byte = ord($binary[$pos]);
                        $value |= ($byte & 0x7F) << $shift;
                        $shift += 7;
                        $pos++;
                    } while ($byte & 0x80);

                    switch ($fieldNumber) {
                        case 1:$result['id'] = $value;
                            break;
                        case 2:$result['progress'] = $value;
                            break;
                        case 3:$result['mode'] = $value;
                            break;
                        case 4:$result['fontsize'] = $value;
                            break;
                        case 5:$result['color'] = $value;
                            break;
                        case 8:$result['ctime'] = $value;
                            break;
                        case 9:$result['weight'] = $value;
                            break;
                        case 11:$result['attr'] = $value;
                            break;
                    }
                    break;

                case 6:  // midHash
                case 7:  // content
                case 10: // idStr
                case 12: // action
                    $len   = 0;
                    $shift = 0;
                    do {
                        $byte = ord($binary[$pos]);
                        $len |= ($byte & 0x7F) << $shift;
                        $shift += 7;
                        $pos++;
                    } while ($byte & 0x80);

                    $value = substr($binary, $pos, $len);
                    $pos += $len;

                    switch ($fieldNumber) {
                        case 6:$result['midHash'] = $value;
                            break;
                        case 7:$result['content'] = $value;
                            break;
                        case 10:$result['idStr'] = $value;
                            break;
                        case 12:$result['action'] = $value;
                            break;
                    }
                    break;

                default:
                    // 跳过未知字段
                    if ($wireType === 0) {
                        // 添加边界检查
                        while ($pos < $length && (ord($binary[$pos++]) & 0x80));
                    } elseif ($wireType === 2) {
                        $len   = 0;
                        $shift = 0;
                        do {
                            // 添加边界检查
                            if ($pos >= $length) {
                                break 2; // 跳出外层循环
                            }
                            $byte = ord($binary[$pos]);
                            $len |= ($byte & 0x7F) << $shift;
                            $shift += 7;
                            $pos++;
                        } while ($byte & 0x80);

                        // 添加长度检查
                        if ($pos + $len > $length) {
                            break 2; // 跳出外层循环
                        }
                        $pos += $len;
                    }
                    break;
            }
        }

        return $result;
    }

    private function getVideoPartFromWebpage(string $bvid)
    {
        $cookies  = parse_netscape_cookie_content($this->settingsService->get(SettingKey::COOKIES_CONTENT));
        $client   = $this->getClient();
        $url      = "https://www.bilibili.com/video/{$bvid}";
        $response = $client->request('GET', $url, [
            'cookies' => $cookies,
        ]);
        $content = $response->getBody()->getContents();

        if (preg_match('/"pages":\s*(\[.*?\])/', $content, $matches)) {
            $pages = json_decode($matches[1], true);
            if ($pages === null) {
                throw new \Exception("JSON 解析失败：" . json_last_error_msg());
            }
            return $pages;
        }
        throw new \Exception("未找到视频分P信息");
    }

    private function getVideoPartFromApi(string $id)
    {
        $info = $this->getVideoInfo($id);
        if ($info && isset($info['pages']) && is_array($info['pages'])) {
            return $info['pages'];
        }
        return null;
    }

    /**
     *  获取视频分P信息
     * 兼容av,bv
     */
    public function getVideoParts(string $strId)
    {
        $parsedParts = null;
        try {
            // api 接口太敏感，优先从网页获取
            if (str_starts_with(strtolower($strId), 'bv')) {
                $bvid        = $strId;
                $parsedParts = $this->getVideoPartFromWebpage($bvid);
            } else {
                $parsedParts = $this->getVideoPartFromWebpage('av' . $strId);
            }
        } catch (\Exception $e) {
            Log::error("通过网页获取视频分P信息失败: " . $e->getMessage());
            try {
                $parsedParts = $this->getVideoPartFromApi($strId);
                Log::info("通过API获取视频分P信息成功: " . $strId, ['parsedParts' => $parsedParts]);
            } catch (\Exception $e) {
                Log::error("通过API获取视频分P信息失败: " . $e->getMessage());
                throw new \Exception("获取视频分P信息失败: " . $e->getMessage());
            }
        }

        return array_map(function ($item) {
            return [
                'cid'         => $item['cid'] ?? 0,
                'page'        => $item['page'] ?? 0,
                'from'        => $item['from'] ?? '',
                'part'        => $item['part'] ?? '',
                'duration'    => $item['duration'] ?? 0,
                'vid'         => $item['vid'] ?? '',
                'weblink'     => $item['weblink'] ?? '',
                'dimension'   => $item['dimension'] ?? null,
                'first_frame' => $item['first_frame'] ?? null,
            ];
        }, $parsedParts ?? []);
    }

    public function pullFav()
    {
        $cookies    = parse_netscape_cookie_content($this->settingsService->get(SettingKey::COOKIES_CONTENT));
        $dedeUserID = $cookies->getCookieByName('DedeUserID');
        if (! $dedeUserID) {
            throw new \Exception("DedeUserID 不存在");
        }
        $mid = $dedeUserID->getValue();

        $pn        = 1;
        $ps        = 20;
        $client    = $this->getClient();
        $favorites = [];
        while (true) {
            $response = $client->request('GET', self::API_HOST . "/x/v3/fav/folder/created/list?pn={$pn}&ps={$ps}&up_mid={$mid}");

            $result = json_decode($response->getBody()->getContents(), true);

            if ($result && $result['code'] == 0) {

                if ($result['data'] == null) {
                    Log::error(sprintf("Account cookie is invalid when accessing the get fav folder api."));
                    return [];
                }
                foreach ($result['data']['list'] as $value) {
                    $favorites[] = [
                        'title'       => $value['title'],
                        'cover'       => $value['cover'],
                        'ctime'       => $value['ctime'],
                        'mtime'       => $value['mtime'],
                        'media_count' => $value['media_count'],
                        'id'          => $value['id'],
                    ];
                }
            }
            if (isset($result['data']['has_more']) && $result['data']['has_more']) {
                $pn++;
            } else {
                break;
            }
        }

        return $favorites;
    }

    public function pullFavVideoList(int $favId, ?int $page = null)
    {
        $client = $this->getClient();
        $pn     = $page ?? 1;
        $videos = [];

        while (true) {
            $url = self::API_HOST . "/x/v3/fav/resource/list?media_id=$favId&pn=$pn&ps={$this->favVideosPageSize}&keyword=&order=mtime&type=0&tid=0&platform=web";
            Log::info("pullFavVideoList fetch $url");

            try {
                $response = $client->request('GET', $url);
                $result   = json_decode($response->getBody()->getContents(), true);

                if (isset($result['data']) && is_array($result['data']['medias'])) {
                    foreach ($result['data']['medias'] as $value) {
                        $videos[] = $value;
                    }
                }

                // 如果指定了页码,只获取该页数据
                if ($page !== null) {
                    break;
                }

                if (isset($result['data']['has_more']) && $result['data']['has_more']) {
                    $pn++;
                } else {
                    break;
                }
            } catch (\Exception $e) {
                Log::error("API request failed: " . $e->getMessage());
                // 如果是频率限制错误，等待更长时间
                // 检查接口响应是否包含429 或者412，如果包含则通过redis记录2个小时。
                if (strpos($e->getMessage(), '429') !== false || strpos($e->getMessage(), '412') !== false) {
                    Log::warning("Rate limit detected, waiting 60 seconds before retry");
                    $this->bilibiliSuspendService->setSuspend();
                    continue;
                }
                throw $e;
            }
        }
        return $videos;
    }

    public function getSeasonsList(int $mid, int $seasonId, int $page = 1)
    {
        try {
            $pageSize = 30;
            $client   = $this->getClient();
            $url      = self::API_HOST . "/x/polymer/web-space/seasons_archives_list?mid={$mid}&season_id={$seasonId}&sort_reverse=false&page_size={$pageSize}&page_num={$page}";
            $response = $client->request('GET', $url);
            $result   = json_decode($response->getBody()->getContents(), true);
            if (is_array($result) && isset($result['data']) && is_array($result['data'])) {
                return $result['data'];
            }
            Log::error("get seasons list failed", ['mid' => $mid, 'season_id' => $seasonId, 'page' => $page, 'result' => $result]);
            throw new \Exception("get seasons list failed");
        } catch (\Exception $e) {
            Log::error("API request failed: " . $e->getMessage());
            if (strpos($e->getMessage(), '429') !== false || strpos($e->getMessage(), '412') !== false) {
                Log::warning("Rate limit detected, waiting 60 seconds before retry");
                $this->bilibiliSuspendService->setSuspend();
            }
            throw $e;
        }
    }

    public function getSeriesMeta(int $seriesId)
    {
        $client = $this->getClient();
        $url    = self::API_HOST . "/x/series/series?series_id={$seriesId}";
        try {
            $response = $client->request('GET', $url);
            $result   = json_decode($response->getBody()->getContents(), true);
            if (is_array($result) && isset($result['data']) && is_array($result['data'])) {
                return $result['data'];
            }
            Log::error("get series meta failed", ['series_id' => $seriesId, 'result' => $result]);
            throw new \Exception("get series meta failed");
        } catch (\Exception $e) {
            Log::error("API request failed: " . $e->getMessage());
            if (strpos($e->getMessage(), '429') !== false || strpos($e->getMessage(), '412') !== false) {
                Log::warning("Rate limit detected, waiting 60 seconds before retry");
                $this->bilibiliSuspendService->setSuspend();
            }
            throw $e;
        }
    }

    public function getSeriesList(int $mid, int $seriesId, int $page = 1)
    {
        try {
            $pageSize = 30;
            $client   = $this->getClient();
            $url      = self::API_HOST . "/x/series/archives?mid={$mid}&series_id={$seriesId}&only_normal=true&sort=desc&ps={$pageSize}&pn={$page}";
            $response = $client->request('GET', $url);
            $result   = json_decode($response->getBody()->getContents(), true);
            if (is_array($result) && isset($result['data']) && is_array($result['data'])) {
                return $result['data'];
            }
            Log::error("get series list failed", ['mid' => $mid, 'series_id' => $seriesId, 'page' => $page, 'result' => $result]);
            throw new \Exception("get series list failed");
        } catch (\Exception $e) {
            Log::error("API request failed: " . $e->getMessage());
            if (strpos($e->getMessage(), '429') !== false || strpos($e->getMessage(), '412') !== false) {
                Log::warning("Rate limit detected, waiting 60 seconds before retry");
                $this->bilibiliSuspendService->setSuspend();
            }
            throw $e;
        }

    }
    public function getUpVideos(int $mid, ?int $offsetAid)
    {
        $client = $this->getAppClient();

        // 准备基础参数
        $params = [
            'vmid'  => $mid,
            'order' => 'pubdate',
            'ps'    => 20,
            'ts'    => time(),
        ];

        // 如果有偏移aid，添加到参数中
        if ($offsetAid !== null) {
            $params['aid'] = $offsetAid;
        }

        // 进行APP签名
        $signedParams = $this->appSign($params, self::APP_KEY, self::APP_SECRET);

        // 构建查询字符串
        $query = http_build_query($signedParams);
        $url   = self::APP_API_HOST . "/x/v2/space/archive/cursor?" . $query;

        try {
            $response = $client->request('GET', $url);
            $result   = json_decode($response->getBody()->getContents(), true);
            if ($result['code'] !== 0) {
                throw new \Exception("get up videos failed: " . $result['message']);
            }
            if (is_array($result['data']['item']) && count($result['data']['item']) > 0) {
                $lastAid = intval(end($result['data']['item'])['param']);
            } else {
                $lastAid = null;
            }

            return [
                'has_next' => $result['data']['has_next'],
                'list'     => $result['data']['item'] ?? [],
                'last_aid' => $lastAid,
            ];
        } catch (\Exception $e) {
            Log::error("API request failed: " . $e->getMessage());
            if (strpos($e->getMessage(), '429') !== false || strpos($e->getMessage(), '412') !== false) {
                Log::warning("Rate limit detected, waiting 60 seconds before retry");
                $this->bilibiliSuspendService->setSuspend();
            }
            throw $e;
        }
    }

    public function getVideoInfo(string $strId): array
    {
        try {
            $client = $this->getClient();
            if (str_starts_with(strtolower($strId), 'bv')) {
                $url = self::API_HOST . "/x/web-interface/view?bvid={$strId}";
            } else {
                $url = self::API_HOST . "/x/web-interface/view?aid={$strId}";
            }
            $response = $client->request('GET', $url);
            $result   = json_decode($response->getBody()->getContents(), true);
            if ($result['code'] !== 0) {
                throw new \Exception("get video info failed: " . $result['message'], $result['code']);
            }
            return $result['data'];
        } catch (\Exception $e) {
            Log::error("API request failed: " . $e->getMessage());
            if (strpos($e->getMessage(), '429') !== false || strpos($e->getMessage(), '412') !== false) {
                Log::warning("Rate limit detected, waiting 60 seconds before retry");
                $this->bilibiliSuspendService->setSuspend();
            }
            throw $e;
        }
    }

    public function getUperCard(int $mid): array
    {
        $client   = $this->getClient();
        $url      = self::API_HOST . "/x/web-interface/card?mid={$mid}";
        $response = $client->request('GET', $url);
        $result   = json_decode($response->getBody()->getContents(), true);
        if ($result['code'] !== 0) {
            throw new \Exception("get uper card failed: " . $result['message'], $result['code']);
        }
        if (isset($result['data']) && isset($result['data']['card'])) {
            return $result['data']['card'];
        }
        return [];
    }

    public function getVideoFestivalJumpUrl(string $strId): ?string
    {
        $data = $this->getVideoInfo($strId);
        if (isset($data['festival_jump_url'])) {
            return $data['festival_jump_url'];
        }
        return null;
    }

    public function getFavFolderResources(int $favId, int $page = 1)
    {
        $client = $this->getClient();
        $url    = self::API_HOST . "/x/v3/fav/folder/resources?media_id={$favId}&pn={$page}&ps={$this->favVideosPageSize}&build=85900200&c_locale=en&device=phone&disable_rcmd=0&mobi_app=iphone&platform=ios&s_locale=en";

        $response = $client->request('GET', $url);
        $result   = json_decode($response->getBody()->getContents(), true);
        $test     = Arr::only($result['data'], ['has_more', 'total', 'has_invalid']);

        if (is_array($result) && $result['code'] == 0) {
            return $result['data']['list'] ? (array) $result['data']['list'] : [];
        }
        return [];
    }
}
