<?php
namespace App\Console\Commands;

use App\Enums\SettingKey;
use App\Models\Danmaku;
use App\Models\FavoriteList;
use App\Models\FavoriteListVideo;
use App\Models\Setting;
use App\Models\Video;
use App\Models\VideoPart;
use Carbon\Carbon;
use Illuminate\Console\Command;

class UpgradeRedisToSqlite extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:upgrade-redis-to-sqlite {--favorite-list} {--video} {--video-part} {--favorite-list-video} {--danmaku} {--video-downloaded-at} {--danmaku-downloaded-at} {--settings}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = '将Redis中的数据迁移到Sqlite中';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        if ($this->option('favorite-list')) {
            $this->upgradeFavoriteList();
        }
        if ($this->option('video')) {
            $this->upgradeVideo();
        }
        if ($this->option('video-part')) {
            $this->upgradeVideoPart();
        }
        if ($this->option('favorite-list-video')) {
            $this->upgradeFavoriteListVideo();
        }
        if ($this->option('danmaku')) {
            $this->upgradeDanmaku();
        }
        if ($this->option('video-downloaded-at')) {
            $this->upgradeVideoDownloadedAt();
        }
        if ($this->option('danmaku-downloaded-at')) {
            $this->upgradeDanmakuDownloadedAt();
        }
        if ($this->option('settings')) {
            $this->upgradeSettings();
        }

    }

    protected function upgradeFavoriteList()
    {
        $redis = redis();
        $data  = $redis->get('fav_list');
        $data  = json_decode($data, true);
        foreach ($data as $item) {
            FavoriteList::query()->updateOrCreate([
                'id' => $item['id'],
            ], [
                'title'       => $item['title'],
                'cover'       => $item['cover'],
                'ctime'       => Carbon::createFromTimestamp($item['ctime']),
                'mtime'       => Carbon::createFromTimestamp($item['mtime']),
                'media_count' => $item['media_count'],
            ]);
        }
    }

    protected function upgradeVideo()
    {
        $redis = redis();
        $keys  = $redis->keys('video:*');
        foreach ($keys as $key) {
            $value = $redis->get($key);
            $data  = json_decode($value, true);

            Video::query()->updateOrCreate([
                'id' => $data['id'],
            ], [
                'id'          => $data['id'],
                'link'        => $data['link'],
                'title'       => $data['title'],
                'intro'       => $data['intro'],
                'cover'       => $data['cover'],
                'bvid'        => $data['bvid'],
                'pubtime'     => Carbon::createFromTimestamp($data['pubtime']),
                'attr'        => $data['attr'],
                'invalid'     => $data['invalid'],
                'frozen'      => $data['frozen'],
                'cache_image' => $data['cache_image'],
                'page'        => $data['page'] ?? 1,
                'fav_time'    => Carbon::createFromTimestamp($data['fav_time']),
            ]);
        }
    }

    protected function upgradeVideoPart()
    {
        $redis = redis();
        $keys  = $redis->hkeys('video_parts');
        foreach ($keys as $key) {
            $videoId = $key;
            $value   = $redis->hget('video_parts', $key);
            $data    = json_decode($value, true);
            foreach ($data['parts'] as $part) {
                VideoPart::query()->updateOrCreate([
                    'video_id' => $videoId,
                    'cid'      => $part['cid'],
                ], [
                    'video_id'    => $videoId,
                    'cid'         => $part['cid'],
                    'page'        => $part['page'],
                    'from'        => $part['from'],
                    'part'        => $part['part'],
                    'duration'    => $part['duration'],
                    'vid'         => $part['vid'],
                    'weblink'     => $part['weblink'],
                    'width'       => $part['dimension']['width'] ?? 0,
                    'height'      => $part['dimension']['height'] ?? 0,
                    'rotate'      => $part['dimension']['rotate'] ?? 0,
                    'first_frame' => $part['first_frame'] ?? '',
                    'created_at'  => Carbon::createFromTimestamp($data['save_time']),
                ]);
            }

        }
    }

    protected function upgradeFavoriteListVideo()
    {
        $redis = redis();
        $keys  = $redis->keys('fav_list:*');
        foreach ($keys as $key) {
            $favoriteListId = explode(':', $key)[1];
            $value          = $redis->get($key);
            $data           = json_decode($value, true);
            foreach ($data as $media) {
                FavoriteListVideo::query()->updateOrCreate([
                    'favorite_list_id' => $favoriteListId,
                    'video_id'         => $media['id'],
                ], [
                    'favorite_list_id' => $favoriteListId,
                    'video_id'         => $media['id'],
                    'created_at'       => Carbon::createFromTimestamp($media['fav_time']),
                ]);
            }
        }
    }

    protected function upgradeDanmaku()
    {
        // 记录任务消耗时间
        $start = microtime(true);
        $redis = redis();
        $keys  = $redis->hkeys('danmaku');

        foreach ($keys as $key => $cid) {
            $value = $redis->hget('danmaku', $cid);

            // 输出进度
            echo 'Danmaku 迁移进度：' . $key . ' / ' . count($keys) . "\r";
            $data = json_decode($value, true);
            if (! $data) {
                continue;
            }
            foreach ($data['danmaku'] as $item) {
                Danmaku::query()->updateOrCreate(
                    [
                        'id' => $item['id'],
                    ],
                    [
                        'video_id'   => 0,
                        'cid'        => $cid,
                        'progress'   => $item['progress'] ?? 0,
                        'mode'       => $item['mode'] ?? 0,
                        'color'      => $item['color'] ?? 0,
                        'content'    => $item['content'] ?? '',
                        'created_at' => Carbon::createFromTimestamp($data['save_time']),
                    ]
                );
                unset($item);
            }
            unset($data);
        }
        $end = microtime(true);
        $this->info('Danmaku 迁移完成，耗时：' . ($end - $start) . '秒');

    }

    protected function upgradeVideoDownloadedAt()
    {
        $redis = redis();
        $keys  = $redis->hkeys('video_downloaded');
        foreach ($keys as $key) {
            Video::query()->where('id', $key)->update([
                'video_downloaded_at' => Carbon::now(),
            ]);
        }
    }
    protected function upgradeDanmakuDownloadedAt()
    {
        $redis = redis();
        $keys  = $redis->hkeys('danmaku_downloaded');
        foreach ($keys as $key) {
            Video::query()->where('id', $key)->update([
                'danmaku_downloaded_at' => Carbon::createFromTimestamp($redis->hget('danmaku_downloaded', $key)),
            ]);
        }
    }

    protected function upgradeSettings()
    {
        $redis = redis();
        $keys  = $redis->hkeys('settings');
        foreach ($keys as $key) {
            $value = $redis->hget('settings', $key);

            // 将key驼峰转为下划线命名
            $key = $this->convertSettingKey($key);

            Setting::query()->updateOrCreate([
                'name' => $key,
            ], [
                'value' => $value,
            ]);
        }
    }

    /**
     * 将驼峰命名转换为下划线命名
     * 例如：camelCase -> camel_case, MultiPartitionCache -> multi_partition_cache
     */
    private function camelToSnake(string $input): string
    {
        // 先处理首字母大写的情况
        $input = lcfirst($input);
        
        // 在单词边界处插入下划线
        $pattern = '/([a-z])([A-Z])/';
        $replacement = '$1_$2';
        
        // 转换为小写
        return strtolower(preg_replace($pattern, $replacement, $input));
    }
    
    protected function convertSettingKey(string $key): string
    {
        switch ($key) {
            case 'MultiPartitionCache':
                return SettingKey::MULTI_PARTITION_DOWNLOAD_ENABLED->value;
            case 'danmakuCache':
                return SettingKey::DANMAKU_DOWNLOAD_ENABLED->value;
            case 'nameExclude':
                return SettingKey::NAME_EXCLUDE->value;
            case 'sizeExclude':
                return SettingKey::SIZE_EXCLUDE->value;
            case 'favExclude':
                return SettingKey::FAVORITE_EXCLUDE->value;
            default:
                return strtolower($key);
        }
    }
}
