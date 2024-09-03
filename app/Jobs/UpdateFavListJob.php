<?php

namespace App\Jobs;

use Arr;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

class UpdateFavListJob implements ShouldQueue
{
    use Queueable;

    protected $SESSDATA = '';
    protected $mid      = 0;

    /**
     * Create a new job instance.
     */
    public function __construct()
    {

    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        list($this->SESSDATA, $this->mid) = match_cookie_main();

        $saveDir = storage_path('app/public/images');

        if (!is_dir($saveDir)) {
            mkdir($saveDir, 0644);
        }

        $favList = $this->pullFav();

        $favList = array_map(function ($value) use ($saveDir) {
            $filename = $this->convertToFilename($value['cover']);

            if ($value['cover']) {
                $savePath = "$saveDir/$filename";
                $this->downloadImg($value['cover'], $savePath);
            }

            $value = array_merge($value, ['cache_image' => $filename]);

            $this->pullFavList($value['id']);

            return $value;
        }, $favList);

        redis()->set('fav_list', json_encode($favList, JSON_UNESCAPED_UNICODE));

        echo "Queues processed: " . count($favList) . "\n";
    }

    public function pullFavList($id)
    {

        $options = [
            'http' => [
                'header' => "Cookie: SESSDATA={$this->SESSDATA}",
            ],
        ];

        $context = stream_context_create($options);
        $pn      = 1;
        $favList = $this->loadFavList($id);

        $saveDir = storage_path('app/public/images');

        while (true) {
            $url = "https://api.bilibili.com/x/v3/fav/resource/list?media_id=$id&pn=$pn&ps=20&keyword=&order=mtime&type=0&tid=0&platform=web";
            echo "fetch $url\n";
            $response = file_get_contents($url, false, $context);
            $result   = json_decode($response, true);

            if (isset($result['data'])) {
                foreach ($result['data']['medias'] as $value) {

                    $filename = $this->convertToFilename($value['cover']);

                    $videoInvalid = $value['attr'] > 0 || $value['title'] == '已失效视频';
                    $exist        = collect($favList)->where('id', $value['id'])->first();

                    if ($value['cover'] && !$videoInvalid) {
                        $this->downloadImg($value['cover'], "$saveDir/$filename");
                    }

                    // 是否冻结该视频: 是否已经保护备份了该视频
                    // 如果已经冻结了该视频, 就不更新该视频的三元素信息
                    $frozen = $exist && $exist['title'] !== '已失效视频' && $videoInvalid;

                    $newValue = [
                        'link'        => $value['link'],
                        'title'       => $value['title'],
                        'intro'       => $value['intro'],
                        'id'          => $value['id'],
                        'fav_time'    => $value['fav_time'],
                        'cover'       => $value['cover'],
                        'bvid'        => $value['bvid'],
                        'pubtime'     => $value['pubtime'],
                        'attr'        => $value['attr'],
                        'invalid'     => $videoInvalid,
                        'frozen'      => $frozen,
                        // metas 是处理过后的文件路径
                        'cache_image' => $filename,
                    ];

                    $favList = $this->updateFavItem($favList, $value['id'], $newValue);
                }
            }

            if (isset($result['data']['has_more']) && $result['data']['has_more']) {
                $pn++;
            } else {
                break;
            }
        }

        $favList = array_map(function ($value) {
            return Arr::only($value, [
                'link',
                'title',
                'intro',
                'id',
                'fav_time',
                'cover',
                'bvid',
                'pubtime',
                'attr',
                'invalid',
                'frozen',
                'cache_image',
            ]);
        }, $favList);

        redis()->set(sprintf('fav_list:%d', $id), json_encode($favList, JSON_UNESCAPED_UNICODE));

        foreach ($favList as $video) {
            redis()->set(sprintf('video:%d', $video['id']), json_encode($video, JSON_UNESCAPED_UNICODE));
        }
    }

    public function pullFav()
    {

        $options = [
            'http' => [
                'header' => "Cookie: SESSDATA=" . $this->SESSDATA,
            ],
        ];

        $context  = stream_context_create($options);
        $response = file_get_contents("https://api.bilibili.com/x/v3/fav/folder/created/list?pn=1&ps=20&up_mid={$this->mid}", false, $context);

        $result    = json_decode($response, true);
        $favorites = [];
        if ($result && $result['code'] == 0) {
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
        return $favorites;
    }

/**
 * 下载图片并保存到指定路径
 * @param string $url 图片的 URL
 * @param string $outputFilename 保存图片的本地路径
 */
    public function downloadImg($url, $outputFilename)
    {
        try {
            $hashPath = $outputFilename . '.hash';
            if (is_file($outputFilename) && is_file($hashPath)) {
                $hashRecord = file_get_contents($hashPath);
                $hashText   = hash_file('sha256', $outputFilename);
                if ($hashRecord === $hashText) {
                    return;
                }
            }
            $imgData = file_get_contents($url);
            if ($imgData === false) {
                throw new \Exception("Failed to fetch image.");
            }

            file_put_contents($outputFilename, $imgData);
            echo "Image downloaded to $outputFilename\n";

            //write hash
            $hashText = hash_file('sha256', $outputFilename);
            file_put_contents($hashPath, $hashText);
        } catch (\Exception $e) {
            echo "Error downloading image: " . $e->getMessage() . " url:\"$url\"\n";
        }
    }

    public function convertToFilename($url)
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

    public function loadFavList($id)
    {

        if ($result = redis()->get(sprintf('fav_list:%d', $id))) {
            $data = json_decode($result, true);
            if ($data) {
                return $data;
            }
        }
        return [];
    }

    public function updateFavItem($favList, $id, $newValue)
    {
        $exist = collect($favList)->contains('id', $id);

        if ($exist) {
            if (!$newValue['frozen']) {
                $favList = array_map(function ($value) use ($newValue) {
                    if ($value['id'] == $newValue['id']) {
                        $newValue = Arr::except($newValue, [
                            'attr',
                            'title',
                            'cover',
                            'cache_image',
                        ]);
                        $value = array_merge($value, $newValue);
                    }
                    return $value;
                }, $favList);
            }
        } else {
            $favList[] = $newValue;
        }

        return $favList;
    }
}
