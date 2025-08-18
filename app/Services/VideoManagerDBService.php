<?php
namespace App\Services;

use App\Contracts\DownloadImageServiceInterface;
use App\Contracts\VideoManagerServiceInterface;
use App\Events\FavoriteUpdated;
use App\Events\VideoPartUpdated;
use App\Events\VideoUpdated;
use App\Models\Danmaku;
use App\Models\FavoriteList;
use App\Models\Video;
use App\Models\VideoPart;
use Arr;
use Carbon\Carbon;
use Illuminate\Database\Eloquent\Collection;
use Illuminate\Support\Facades\DB;
use Log;

class VideoManagerDBService implements VideoManagerServiceInterface
{

    public function __construct(
        public SettingsService $settings,
        public BilibiliService $bilibiliService,
        public DownloadImageServiceInterface $downloadImageService
    ) {

    }

    public function getVideoInfo(string $id, bool $withParts = false): ?Video
    {
        $video = Video::query()->where('id', $id)->first();
        
        if ($video && $withParts) {
            $video->load('parts');
        }
        return $video;
    }

    public function getVideoDanmakuCount(Video $video): int
    {
        $count = 0;
        foreach ($video->parts as $part) {
            $count += Danmaku::where('cid', $part->cid)->count();
        }
        return $count;
    }

    public function getVideoListByFav(int $favId): array
    {
        return FavoriteList::query()->where('id', $favId)->first()->videos()->get()->toArray();
    }

    /**
     * 获取所有收藏夹列表
     * @return array<FavoriteList>
     */
    public function getFavList(): array
    {
        return FavoriteList::query()->get()
            ->toArray();
    }

    public function getFavDetail(int $favId, array $columns = ['*']): ?FavoriteList
    {
        return FavoriteList::query()->where('id', $favId)->first($columns);
    }

    public function updateFavList(): void
    {
        Log::info('Update fav list start');
        $favList = $this->bilibiliService->pullFav();

        DB::transaction(function() use ($favList){
            array_map(function ($item) {
                $favorite = FavoriteList::query()->where('id', $item['id'])->first();
                if (! $favorite) {
                    $favorite = new FavoriteList();
                }
    
                $oldFav = $favorite->toArray();
    
                $favorite->fill($item);
                $favorite->save();
    
                event(new FavoriteUpdated($oldFav, $item));
    
                return $item;
            }, $favList);            
        });

        Log::info('Update fav list success');
    }

    public function getFavoriteVideo(int $favoriteId): array
    {
        $videos = redis()->get(sprintf('favorite_video_saving:%s', $favoriteId));
        if ($videos === null) {
            return [];
        }
        $decoded = json_decode($videos, true);
        return is_array($decoded) ? $decoded : [];
    }

    public function saveFavoriteVideo(int $favoriteId, array $videos): void
    {
        $existVideos = $this->getFavoriteVideo($favoriteId);
        if (is_array($existVideos) && count($existVideos) > 0) {
            // 使用id作为键进行去重合并，新数据会覆盖旧数据
            $mergedVideos = [];
            
            // 先添加已存在的视频
            foreach ($existVideos as $video) {
                if (isset($video['id'])) {
                    $mergedVideos[$video['id']] = $video;
                }
            }
            
            // 再添加新视频，会覆盖相同id的旧数据
            foreach ($videos as $video) {
                if (isset($video['id'])) {
                    $mergedVideos[$video['id']] = $video;
                }
            }
            
            $videos = array_values($mergedVideos);
        }
        redis()->set(sprintf('favorite_video_saving:%s', $favoriteId), json_encode($videos));
    }


    public function updateFavVideos(array $fav, ?int $page = null): void
    {
        $favId = $fav['id'];
        $videos = $this->bilibiliService->pullFavVideoList($favId, $page);

        if(count($videos) === 0){
            Log::info('No videos found in fav', ['favId' => $favId]);
            return;
        }

        $videos = array_map(function ($item) {
            $videoInvalid = $this->videoIsInvalid($item);

            $exist = $this->getVideoInfo($item['id']);

            // 是否冻结该视频: 是否已经保护备份了该视频
            // 如果已经冻结了该视频, 就不更新该视频的主要信息
            $frozen          = $exist && $exist['title'] !== '已失效视频' && $videoInvalid;
            $item['frozen']  = $frozen;
            $item['invalid'] = $videoInvalid;

            if ($frozen) {
                Log::info('Frozen video', ['id' => $item['id'], 'title' => $exist['title']]);
                $item     = array_merge($exist->toArray(), Arr::except($item, ['attr', 'title', 'cover', 'intro']));
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
                'page'        => $newValue['page'],
                'fav_time'    => $newValue['fav_time'],
            ];
        }, $videos);

        // 暂存视频数据
        $this->saveFavoriteVideo($favId, $videos);

        DB::transaction(function() use ($videos, $favId, $fav){
            $remoteVideoIds = array_column($videos, 'id');
            $localVideoIds  = DB::table('favorite_list_videos')
                ->where('favorite_list_id', $favId)
                ->pluck('video_id')
                ->toArray();
    
            $addVideoIds    = array_diff($remoteVideoIds, $localVideoIds);

            // 添加新的视频关联关系
            if (! empty($addVideoIds)) {
                $attachData = [];
                foreach ($videos as $video) {
                    if (in_array($video['id'], $addVideoIds)) {
                        $attachData[$video['id']] = [
                            'created_at' => date('Y-m-d H:i:s', $video['fav_time']),
                            'updated_at' => date('Y-m-d H:i:s'),
                        ];
                    }
                }
                if (! empty($attachData)) {
                    $favoriteList = FavoriteList::query()->where('id', $favId)->first();
                    $favoriteList->videos()->attach($attachData);
                    Log::info('Attached videos to favorite list', ['favId' => $favId, 'videoIds' => array_keys($attachData)]);
                }
            }


            foreach ($videos as $key => $item) {
                $video = Video::query()->where('id', $item['id'])->first();
                if (! $video) {
                    $video = new Video();
                }
    
                $oldVideoData = $video->toArray();
    
                // 检查视频数据是否真正发生了变化
                $hasChanges = false;
                $changedFields = [];
                
                foreach ($item as $field => $value) {
                    if (!isset($oldVideoData[$field]) || $oldVideoData[$field] != $value) {
                        $hasChanges = true;
                        $changedFields[$field] = [
                            'old' => $oldVideoData[$field] ?? null,
                            'new' => $value
                        ];
                    }
                }
    
                $video->fill($item);
                $video->save();
    
                // 只有在数据真正发生变化时才触发事件
                if ($hasChanges) {
                    Log::info('Video data changed, triggering VideoUpdated event', [
                        'id' => $item['id'], 
                        'title' => $item['title'],
                        'changed_fields' => array_keys($changedFields)
                    ]);
                    event(new VideoUpdated($oldVideoData, $video->toArray()));
                } else {
                    Log::info('Video data unchanged, skipping VideoUpdated event', [
                        'id' => $item['id'], 
                        'title' => $item['title']
                    ]);
                }
    
                Log::info('Update video success', ['id' => $item['id'], 'title' => $item['title']]);
            }

            $savedVideos = $this->getFavoriteVideo($favId);
            // media_count 是不准确的，可能是用户手动取消收藏视频系统没更新、也可能是哔哩哔哩自己故意把用户视频搞丢失
            // 经过测试，用户收藏夹的视频获取没有错误，但收藏夹的media_count 是错误的，目前没有合适的方法更新
            if(intval($fav['media_count']) == count($savedVideos)){
                $remoteCacheVideoIds = array_column($savedVideos, 'id');
                $deleteVideoIds = array_diff($localVideoIds, $remoteCacheVideoIds);
                if(!empty($deleteVideoIds)){
                    FavoriteList::query()->where('id', $favId)->first()->videos()->detach($deleteVideoIds);
                    Log::info('Detached videos from favorite list', ['favId' => $favId, 'videoIds' => $deleteVideoIds]);
                }
            }
        });
    }

    public function updateVideoParts(Video $video): void
    {
        if ($this->videoIsInvalid($video->toArray())) {
            Log::info('Video is invalid, skip update video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        // 后面考虑要不要控制频率，减少风控
        // if ($video->video_downloaded_at && $video->video_downloaded_at > Carbon::now()->subDays(7)) {
        //     Log::info('Video parts has been saved in the last 7 days', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
        //     return;
        // }

        try {
            if(config('services.bilibili.id_type') == 'bv'){
                $videoParts = $this->bilibiliService->getVideoParts($video->bvid);
            }else{
                $videoParts = $this->bilibiliService->getVideoParts($video->id);
            }
        } catch (\Exception $e) {
            Log::error('Get video parts failed', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        // 找出本地多余远端的数据
        $localVideoParts     = $this->getAllPartsVideo($video->id)->toArray();
        $localVideoPartsIds  = array_column($localVideoParts, 'cid');
        $remoteVideoPartsIds = array_column($videoParts, 'cid');

        $deleteVideoPartsIds = array_diff($localVideoPartsIds, $remoteVideoPartsIds);

        if (! empty($deleteVideoPartsIds)) {
            Log::info('Delete video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title, 'deleteVideoPartsIds' => $deleteVideoPartsIds]);
            return;
        }

        DB::transaction(function() use ($videoParts, $video){
            foreach ($videoParts as $part) {
                $videoPart = VideoPart::where('cid', $part['cid'])->first();
                if (! $videoPart) {
                    $videoPart = new VideoPart();
                }
                $oldVideoPart = $videoPart->toArray();
                $videoPart->fill(array_merge(
                    [
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
                        'cid'         => $part['cid'],
                    ],
                    [
                        'video_id' => $video->id,
                    ]
                ));
                $videoPart->save();
    
                event(new VideoPartUpdated($oldVideoPart, $videoPart->toArray()));
            }
    
            $video->video_downloaded_at = now();
            $video->save();
        });

        Log::info('Update video parts success', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
    }

    public function videoIsInvalid(array $video): bool
    {
        return $video['attr'] > 0 || $video['title'] == '已失效视频';
    }

    public function getAllPartsVideo(string $id): Collection
    {
        return VideoPart::where('video_id', $id)
            ->orderBy('page', 'asc')
            ->get();

    }

    public function getAllPartsVideoForUser(Video $video): array
    {
        $list = [];
        foreach ($video->parts as $videoPart) {
            if (! empty($videoPart['video_download_url'])) {
                $urlPath = $videoPart['video_download_url'];
            } else {
                $urlPath = null;
            }
            $list[] = [
                'id'    => $videoPart['cid'],
                'part'  => $videoPart['page'],
                'url'   => $urlPath,
                'title' => $videoPart['part'] ?? 'P' . $videoPart['page'],
            ];
        }
        return $list;
    }

    protected function saveDanmaku(string $cid, array $danmaku): void
    {
        $danmakuIds = Danmaku::where('cid', $cid)->select([
            'id',
        ])->get();
        $danmakuIds = $danmakuIds->pluck('id')->toArray();

        // 过滤出新增的数据
        $insertData = array_filter($danmaku, function ($item) use ($danmakuIds) {
            // 内容为空的也不要
            if (empty($item['content'])) {
                return false;
            }
            return ! in_array($item['id'], $danmakuIds);
        });

        // 释放内存
        unset($danmakuIds);

        $start_time = microtime(true);
        $videoPart  = VideoPart::where('cid', $cid)->first();
        $insertData = array_map(function ($item) use ($videoPart) {
            $item['video_id'] = $videoPart->video_id;
            return $item;
        }, $insertData);
        Log::info('Save danmaku count, array_map time', ['time' => microtime(true) - $start_time]);

        $start_time = microtime(true);
        // 分批插入数据，每批1000条
        DB::transaction(function() use (&$insertData, &$videoPart){
            foreach (array_chunk($insertData, 1000) as $chunk) {
                $videoPart->danmakus()->createMany($chunk);
                unset($chunk);
            }

            // 释放原始数据
            unset($insertData);

            $videoPart->danmakus()->update([
                'video_id' => $videoPart->video_id,
            ]);

            unset($videoPart);
        });
        Log::info('Save danmaku count, DB transaction time', ['time' => microtime(true) - $start_time]);
    }

    public function getDanmaku(string $cid): array
    {
        return Danmaku::where('cid', $cid)->get()->toArray();
    }

    public function downloadDanmaku(VideoPart $videoPart): void
    {
        //加锁
        $lock = redis()->setnx(sprintf('danmaku_downloading:%s', $videoPart->cid), 1);
        if (! $lock) {
            Log::info('Danmaku is being downloaded', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
            return;
        }
        redis()->expire(sprintf('danmaku_downloading:%s', $videoPart->cid), 3600 * 8);

        try {
            Log::info('Download danmaku start', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
            $video = Video::where('id', $videoPart->video_id)->first();

            $start_time = microtime(true);
            $partDanmakus = $this->bilibiliService->getDanmaku($videoPart->cid, intval($videoPart->duration));
            Log::info('Download danmaku time', ['time' => microtime(true) - $start_time]);

            Log::info('Download danmaku success', [
                'id'       => $videoPart->cid,
                'title'    => $videoPart->part,
                'count'    => count($partDanmakus),
                'video_id' => $video->id,
                'bvid'     => $video->bvid,
                'title'    => $video->title,
            ]);
            $this->saveDanmaku($videoPart->cid, $partDanmakus);
            Log::info('Save danmaku time', ['time' => microtime(true) - $start_time]);
            $videoPart->danmaku_downloaded_at = Carbon::now();
            $videoPart->save();
            
        } catch (\Throwable $e) {
            Log::error('Download danmaku failed', ['id' => $videoPart->cid, 'title' => $videoPart->part, 'error' => $e->getMessage()]);
            return;
        } finally {
            // 移除下载锁
            redis()->del(sprintf('danmaku_downloading:%s', $videoPart->cid));
            Log::info('Download danmaku end, take time', ['time' => microtime(true) - $start_time, 'id' => $videoPart->cid, 'title' => $videoPart->part]);
        }
    }

    public function getVideos(array $conditions = []): Collection
    {
        $query = Video::query();
        if (isset($conditions['invalid'])) {
            $query->where('invalid', $conditions['invalid']);
        }
        if (isset($conditions['frozen'])) {
            $query->where('frozen', $conditions['frozen']);
        }
        return $query->get();
    }

    public function getVideosStat(array $conditions = []): array
    {
        $stat = [
            'count'      => Video::count(),
            'downloaded' => Video::where('video_downloaded_num', '>', 0)->count(),
            'invalid'    => Video::where('invalid', 1)->count(),
            'valid'      => Video::where('invalid', 0)->count(),
            'frozen'     => Video::where('frozen', 1)->count(),
        ];
        return $stat;
    }

    public function countVideos(): int
    {
        return Video::count();
    }

    public function pullVideoInfo(string $bvid): void
    {
        try{
            $videoInfo = app(BilibiliService::class)->getVideoInfo($bvid);
            $aid       = $videoInfo['aid'];
            $video     = Video::query()->firstOrNew(['id' => $aid]);
            //  ['id', 'link', 'title', 'intro', 'cover', 'bvid', 'pubtime', 'attr', 'invalid', 'frozen', 'cache_image', 'page', 'fav_time', 'danmaku_downloaded_at', 'video_downloaded_at'];
            $video->fill([
                'link'                  => sprintf('bilibili://video/%s', $aid),
                'title'                 => $videoInfo['title'],
                'intro'                 => $videoInfo['desc'],
                'cover'                 => $videoInfo['pic'],
                'bvid'                  => $videoInfo['bvid'],
                'pubtime'               => Carbon::createFromTimestamp($videoInfo['pubdate']),
                'attr'                  => 0,
                'invalid'               => $videoInfo['state'] == 0 ? false : true,
                'frozen'                => $videoInfo['state'] == 0 ? false : true,
                'cache_image'           => '',
                'page'                  => count($videoInfo['pages']),
                'fav_time'              => null,
            ]);
            $video->save();
            event(new VideoUpdated([], $video->toArray()));
        }catch(\Exception $e){
            Log::error("PullVideoInfoJob failed: " . $e->getMessage());
            if($e->getCode() == -404){
                $video = Video::where('bvid', $bvid)->first();
                if($video && !$video->invalid){
                    $oldVideo = $video->toArray();
                    $video->invalid = true;
                    $video->frozen = true;
                    $video->save();
                    event(new VideoUpdated($oldVideo, $video->toArray()));
                }
            }
            throw $e;
        }
    }
}
