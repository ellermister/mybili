<?php

namespace App\Services;

use App\Jobs\DownloadDanmakuJob;
use App\Jobs\DownloadVideoJob;
use Arr;
use Log;

class VideoManagerService
{
    public function __construct(public SettingsService $settings, public BilibiliService $bilibiliService)
    {

    }

    public function getImagesDirIfNotExistCreate()
    {
        $dir = storage_path('app/public/images');
        if (!is_dir($dir)) {
            mkdir($dir, 0644);
        }
        return $dir;
    }

    public function getVideoList()
    {
        $iterator = null;
        $keys     = [
        ];
        do {
            $result = redis()->scan($iterator, 'video:*', 50);
            $keys   = array_merge($keys, $result);
        } while ($iterator != 0);

        $videos = [];
        foreach ($keys as $key) {
            $videos[] = json_decode(redis()->get($key), true);
        }
        return $videos;
    }

    public function getVideoInfo(string $id)
    {
        $videoInfo = redis()->get(sprintf('video:%s', $id));
        if ($videoInfo) {
            return json_decode($videoInfo, true);
        }
        return null;
    }

    public function getVideoListByFav(int $favId)
    {
        $result = redis()->get(sprintf('fav_list:%d', $favId));
        $data   = json_decode($result, true);
        return $data ?? [];
    }

    public function getFavList()
    {
        $result = redis()->get('fav_list');
        $data   = json_decode($result, true);
        $data   = $data ?? [];
        $data   = array_map(function ($item) {
            // $item['downloaded'] = $this->videoDownloaded($item['id']);
            $item['cache_image'] = $this->convertToFilename($item['cover']);
            return $item;
        }, $data);
        return $data;
    }

    public function updateFavList()
    {
        $favList = $this->bilibiliService->pullFav();
        array_map(function ($item) {
            if ($item['cover']) {
                $savePath = $this->getImagesDirIfNotExistCreate() . '/' . $this->convertToFilename($item['cover']);
                $this->downloadImage($item['cover'], $savePath);
                Log::info('Download image success', ['path' => $savePath]);
            }

            $this->updateFavVideos($item['id']);
            return $item;
        }, $favList);
        redis()->set('fav_list', json_encode($favList, JSON_UNESCAPED_UNICODE));

        Log::info('Update fav list success');
    }

    private function updateFavVideos(int $favId)
    {
        $videos = $this->bilibiliService->pullFavVideoList($favId);
        $videos = array_map(function ($item) {
            $videoInvalid = $this->videoIsInvalid($item);

            $exist = $this->getVideoInfo($item['id']);

            // 是否冻结该视频: 是否已经保护备份了该视频
            // 如果已经冻结了该视频, 就不更新该视频的三元素信息
            $frozen          = $exist && $exist['title'] !== '已失效视频' && $videoInvalid;
            $item['frozen']  = $frozen;
            $item['invalid'] = $videoInvalid;

            if ($frozen) {
                Log::info('Frozen video', ['id' => $item['id'], 'title' => $exist['title']]);
                $item     = array_merge($exist, Arr::except($item, ['attr', 'title', 'cover', 'cache_image']));
                $newValue = $item;
            } else {
                $newValue = $item;
            }

            //在此做键值对映射，避免字段未来变更
            return [
                'id'          => $item['id'],
                'link'        => $newValue['link'],
                'title'       => $newValue['title'],
                'intro'       => $newValue['intro'],
                'cover'       => $newValue['cover'],
                'bvid'        => $newValue['bvid'],
                'pubtime'     => $newValue['pubtime'],
                'attr'        => $newValue['attr'],
                'invalid'     => $newValue['invalid'],
                'frozen'      => $newValue['frozen'],
                'cache_image' => $newValue['cache_image'] ?? '',
                'page'        => $newValue['page'],
                'fav_time'    => $newValue['fav_time'],
            ];
        }, $videos);

        foreach ($videos as $key => $item) {

            $cacheImage = $item['cache_image'] ?? '';
            if ($item['cover']) {
                $savePath = $this->getImagesDirIfNotExistCreate() . '/' . $this->convertToFilename($item['cover']);
                $this->downloadImage($item['cover'], $savePath);
                $cacheImage = $this->convertToFilename($item['cover']);
                Log::info('Download image success', ['path' => $savePath]);
            }

            $item['cache_image'] = $cacheImage;

            redis()->set(sprintf('video:%d', $item['id']), json_encode($item, JSON_UNESCAPED_UNICODE));

            $videos[$key] = $item;

            $this->updateVideoParts($item);

            Log::info('Update video success', ['id' => $item['id'], 'title' => $item['title']]);
        }

        redis()->set(sprintf('fav_list:%d', $favId), json_encode($videos, JSON_UNESCAPED_UNICODE));
    }

    private function updateVideoParts(array $video)
    {
        if($this->videoIsInvalid($video)){
            Log::info('Video is invalid, skip update video parts', ['id' => $video['id'], 'bvid' => $video['bvid'], 'title' => $video['title']]);
            return;
        }
        
        $exist = json_decode(redis()->hGet('video_parts', $video['id']), true);

        if ($exist && $exist['save_time'] > time() - 3600 * 24 * 7) {
            Log::info('Video parts has been saved in the last 7 days', ['id' => $video['id'], 'bvid' => $video['bvid'], 'title' => $video['title']]);
            return;
        }

        try{
            $videoParts = $this->bilibiliService->getVideoParts($video['bvid']);
        }catch(\Exception $e){
            Log::error('Get video parts failed', ['id' => $video['id'], 'bvid' => $video['bvid'], 'title' => $video['title']]);
            return;
        }
        if ($exist && count($exist['parts']) > count($videoParts)) {
            Log::info('Video parts has been deleted by creator', ['id' => $video['id'], 'bvid' => $video['bvid'], 'title' => $video['title']]);
            return;
        }
        $saveData = [
            'save_time' => time(),
            'parts'     => $videoParts,
        ];
        redis()->hset('video_parts', $video['id'], json_encode($saveData, JSON_UNESCAPED_UNICODE));
        Log::info('Update video parts success', ['id' => $video['id'], 'bvid' => $video['bvid'], 'title' => $video['title']]);
    }

    public function videoIsInvalid(array $video)
    {
        return $video['attr'] > 0 || $video['title'] == '已失效视频';
    }

    private function downloadImage(string $url, string $savePath)
    {
        $hashPath = $savePath . '.hash';
        if (is_file($hashPath)) {
            $hash     = file_get_contents($hashPath);
            $hashText = hash_file('sha256', $savePath);
            if ($hash === $hashText) {
                return;
            }
        }
        $content = file_get_contents($url, false);
        if (!$content) {
            throw new \Exception("Failed to fetch image.");
        }
        file_put_contents($savePath, $content);
        file_put_contents($hashPath, hash_file('sha256', $savePath));
    }

    private function convertToFilename(string $url)
    {
        if (empty($url)) {
            return "";
        }

        $urlParts = explode('/', $url);
        $filename = end($urlParts);

        if (!$filename) {
            $filename = base64_encode($url) . '.jpg';
        }
        $filename = preg_replace('/[^a-zA-Z90-9](?!(jpg|png|gif|svg|webp))/', '', $filename);
        return $filename;
    }

    public function videoDownloaded(string $id)
    {
        $value = redis()->hGet('video_downloaded', $id);
        return $value ? true : false;
    }

    private function setVideoDownloaded(string $avId)
    {
        redis()->hSet('video_downloaded', $avId, 1);
    }

    private function delVideoDownloaded(string $avId)
    {
        redis()->hDel('video_downloaded', $avId);
    }

    public function danmakuDownloadedTime(string $avId)
    {
        $value = redis()->hGet('danmaku_downloaded', $avId);
        return $value ? intval($value) : null;
    }

    private function setDanmakuDownloadedTime(string $avId)
    {
        redis()->hSet('danmaku_downloaded', $avId, time());
    }

    /**
     * 视频文件和hash同时存在才认为有效
     */
    public function hasVideoFile(string $id, int $part = 1)
    {
        $savePath = $this->getVideoDownloadPath($id, $part);
        $hashPath = $this->getVideoDownloadHashPath($id, $part);
        return is_file($savePath) && is_file($hashPath);
    }

    public function getVideoFileHash(string $id)
    {
        $hashPath = $this->getVideoDownloadHashPath($id);

        if (is_file($hashPath)) {
            $content = file_get_contents($hashPath);
            if ($content) {
                return $content;
            }
        }
        return null;
    }

    public function createVideoDirectory()
    {
        $videoPath = storage_path('app/public/videos');
        if (!is_dir($videoPath)) {
            mkdir($videoPath, 0644);
            if (!is_dir($videoPath)) {
                throw new \Exception("下载路径不存在, 创建失败");
            }
        }
    }

    public function getVideoDownloadPath(string $id, int $part = 1)
    {
        $videoPath = storage_path('app/public/videos');
        if ($part === 1) {
            return sprintf('%s/%s.mp4', $videoPath, $id);
        }
        return sprintf('%s/%s.part%d.mp4', $videoPath, $id, $part);
    }

    public function getVideoDownloadHashPath(string $id, int $part = 1)
    {
        $videoPath = storage_path('app/public/videos');
        if ($part === 1) {
            return sprintf('%s/%s.mp4.hash', $videoPath, $id);
        }
        return sprintf('%s/%s.part%d.mp4.hash', $videoPath, $id, $part);
    }

    public function markVideoDownloaded(string $id, int $part = 1)
    {
        $savePath = $this->getVideoDownloadPath($id, $part);
        $hashPath = $this->getVideoDownloadHashPath($id, $part);
        if (is_file($savePath)) {
            $calcHash = hash_file('sha256', $savePath);
            if ($calcHash !== $this->getVideoFileHash($id)) {
                file_put_contents($hashPath, $calcHash);
                $this->setVideoDownloaded($id);
            }
        }
    }

    public function getAllPartsVideo(string $id)
    {
        $videoParts = json_decode(redis()->hGet('video_parts', $id), true);
        return $videoParts['parts'] ?? [];
    }

    public function getAllPartsVideoForUser(string $id, int $parts = 1)
    {

        // 先从redis获取已经缓存的分P数据
        $saveParts = json_decode(redis()->hGet('video_parts', $id), true);
        $videoParts = null;
        if($saveParts){
            $videoParts = collect($saveParts['parts']);
        }

        $list = [];
        for ($i = 1; $i <= $parts; $i++) {
            $savePath = $this->getVideoDownloadPath($id, $i);
            $urlPath  = str_replace(storage_path('app/public/'), '', $savePath);
            if (is_file($savePath)) {

                $title = '';
                $cid = 0;
                if($videoParts){
                    $title = $videoParts->where('page', $i)->first()['part'] ?? '';
                    $cid = $videoParts->where('page', $i)->first()['cid'] ?? 0;
                }

                $list[] = [
                    'id'  => $cid,
                    'part' => $i,
                    'url' => '/storage/' . $urlPath,
                    'title' => $title,
                ];
            }
        }
        return $list;
    }

    public function dispatchDownloadVideoJob(array $video)
    {
        $id     = $video['id'];
        $exists = redis()->setnx(sprintf('video_downloading:%s', $id), 1);
        if ($exists) {
            redis()->expire(sprintf('video_downloading:%s', $id), 3600 * 8);
            $job = new DownloadVideoJob($video);
            dispatch($job);
        }
    }

    public function finishDownloadVideo(string $id)
    {
        redis()->del(sprintf('video_downloading:%s', $id));
    }


    public function dispatchDownloadDanmakuJob(int $avId)
    {
        $exists = redis()->setnx(sprintf('danmaku_downloading:%s', $avId), 1);
        if ($exists) {
            redis()->expire(sprintf('danmaku_downloading:%s', $avId), 3600 * 8);
            $job = new DownloadDanmakuJob($avId);
            dispatch($job);
        }
    }

    public function saveDanmaku(string $cid, array $danmaku)
    {
        $saveData = [
            'save_time' => time(),
            'danmaku' => $danmaku,
        ];

        //创建去重映射map
        $map = [];
        foreach($danmaku as $item){
            $map[$item['id']] = 1;
        }

        $exist = $this->getDanmaku($cid);
        if($exist && count($exist['danmaku']) > 0){
            // 根据id去重，如果存在则不添加
            foreach($exist['danmaku'] as $item){
                if(!isset($map[$item['id']])){
                    $saveData['danmaku'][] = $item;
                }
            }
        }

        // 释放内存
        unset($map);
        unset($exist);

        usort($saveData['danmaku'], function($a, $b){
            return intval($a['progress'] ?? 0) - intval($b['progress'] ?? 0);
        });
        redis()->hSet('danmaku', $cid, json_encode($saveData, JSON_UNESCAPED_UNICODE));
    }

    public function getDanmaku(string $cid)
    {
        $danmaku = redis()->hGet('danmaku', $cid);
        return json_decode($danmaku, true);
    }

    public function downloadDanmaku(int $avId)
    {
        $parts = $this->getAllPartsVideo($avId);
        foreach ($parts as $part) {
            $danmaku = $this->getDanmaku($part['cid']);
            if($danmaku && $danmaku['save_time'] > time() - 3600 * 24 * 7){
                continue;
            }
            $partDanmakus = $this->bilibiliService->getDanmaku($part['cid'], intval($part['duration']));


            Log::info('Download danmaku success', ['id' => $part['cid'], 'title' => $part['part'], 'count' => count($partDanmakus)]);
            $this->saveDanmaku($part['cid'], $partDanmakus);

            // 移除下载锁
            redis()->del(sprintf('danmaku_downloading:%s', $avId));

            // 标记下载完成
            $this->setDanmakuDownloadedTime($avId); 
        } 
    }
}

