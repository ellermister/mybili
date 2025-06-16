<?php
namespace App\Services;

use App\Contracts\DownloadImageServiceInterface;
use App\Contracts\VideoManagerServiceInterface;
use App\Events\FavoriteUpdated;
use App\Events\VideoPartUpdated;
use App\Events\VideoUpdated;
use App\Jobs\DownloadDanmakuJob;
use App\Jobs\DownloadVideoJob;
use App\Models\Danmaku;
use App\Models\FavoriteList;
use App\Models\Video;
use App\Models\VideoPart;
use Arr;
use Carbon\Carbon;
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

    public function updateFavList(): void
    {
        Log::info('Update fav list start');
        $favList = $this->bilibiliService->pullFav();
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

        Log::info('Update fav list success');
    }

    public function updateFavVideos(int $favId): void
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

        $remoteVideoIds = array_column($videos, 'id');
        $localVideoIds  = DB::table('favorite_list_videos')
            ->where('favorite_list_id', $favId)
            ->pluck('video_id')
            ->toArray();

        $addVideoIds    = array_diff($remoteVideoIds, $localVideoIds);
        $deleteVideoIds = array_diff($localVideoIds, $remoteVideoIds);

        // 删除收藏夹与视频的关系，但不删除视频表数据
        if (! empty($deleteVideoIds)) {
            FavoriteList::query()->where('id', $favId)->first()->videos()->detach($deleteVideoIds);
            Log::info('Detached videos from favorite list', ['favId' => $favId, 'videoIds' => $deleteVideoIds]);
        }

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

            $video->fill($item);
            $video->save();

            event(new VideoUpdated($oldVideoData, $item));

            Log::info('Update video success', ['id' => $item['id'], 'title' => $item['title']]);
        }
    }

    public function updateVideoParts(Video $video): void
    {
        if ($this->videoIsInvalid($video->toArray())) {
            Log::info('Video is invalid, skip update video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        if ($video->video_downloaded_at && $video->video_downloaded_at > Carbon::now()->subDays(7)) {
            Log::info('Video parts has been saved in the last 7 days', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        try {
            $videoParts = $this->bilibiliService->getVideoParts($video->bvid);
        } catch (\Exception $e) {
            Log::error('Get video parts failed', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        // 找出本地多余远端的数据
        $localVideoParts     = $this->getAllPartsVideo($video->id);
        $localVideoPartsIds  = array_column($localVideoParts, 'cid');
        $remoteVideoPartsIds = array_column($videoParts, 'cid');

        $deleteVideoPartsIds = array_diff($localVideoPartsIds, $remoteVideoPartsIds);

        if (! empty($deleteVideoPartsIds)) {
            Log::info('Delete video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title, 'deleteVideoPartsIds' => $deleteVideoPartsIds]);
            return;
        }

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

        Log::info('Update video parts success', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
    }

    public function videoIsInvalid(array $video): bool
    {
        return $video['attr'] > 0 || $video['title'] == '已失效视频';
    }

    public function videoDownloaded(string $id): bool
    {
        return Video::where('id', $id)->value('video_downloaded_at') ? true : false;
    }

    public function setVideoDownloaded(string $avId): void
    {
        Video::where('id', $avId)->update([
            'video_downloaded_at' => Carbon::now(),
        ]);
    }

    public function delVideoDownloaded(string $avId): void
    {
        Video::where('id', $avId)->update([
            'video_downloaded_at' => null,
        ]);
    }

    public function danmakuDownloadedTime(string $avId): ?int
    {
        $time = Video::where('id', $avId)->value('danmaku_downloaded_at');
        return $time ? Carbon::parse($time)->timestamp : null;
    }

    public function setDanmakuDownloadedTime(string $avId): void
    {
        Video::where('id', $avId)->update([
            'danmaku_downloaded_at' => Carbon::now(),
        ]);
    }

    public function getAllPartsVideo(string $id): array
    {
        return VideoPart::where('video_id', $id)
            ->orderBy('page', 'asc')
            ->get()
            ->toArray();

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

    public function dispatchDownloadVideoJob(array $video): void
    {
        $id     = $video['id'];
        $exists = redis()->setnx(sprintf('video_downloading:%s', $id), 1);
        if ($exists) {
            redis()->expire(sprintf('video_downloading:%s', $id), 3600 * 8);
            $job = new DownloadVideoJob($video);
            dispatch($job);
        }
    }

    public function finishDownloadVideo(string $id): void
    {
        redis()->del(sprintf('video_downloading:%s', $id));
    }

    public function dispatchDownloadDanmakuJob(int $avId): void
    {
        $exists = redis()->setnx(sprintf('danmaku_downloading:%s', $avId), 1);
        if ($exists) {
            redis()->expire(sprintf('danmaku_downloading:%s', $avId), 3600 * 8);
            $job = new DownloadDanmakuJob($avId);
            dispatch($job);
        }
    }

    public function saveDanmaku(string $cid, array $danmaku): void
    {
        $danmakuIds = Danmaku::where('cid', $cid)->select([
            'id',
        ])->get();
        $danmakuIds = $danmakuIds->pluck('id')->toArray();

        // 过滤出新增的数据
        $insertData = array_filter($danmaku, function ($item) use ($danmakuIds) {
            return ! in_array($item['id'], $danmakuIds);
        });

        // 释放内存
        unset($danmakuIds);

        $videoPart  = VideoPart::where('cid', $cid)->first();
        $insertData = array_map(function ($item) use ($videoPart) {
            $item['video_id'] = $videoPart->video_id;
            return $item;
        }, $insertData);

        // 分批插入数据，每批1000条
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
            $video = Video::where('id', $videoPart->video_id)->first();
            // 如果上次下载时间小于7天则不更新
            if ($videoPart->danmaku_downloaded_at && $videoPart->danmaku_downloaded_at > Carbon::now()->subDays(7)) {
                Log::info('Danmaku has been saved in the last 7 days', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                return;
            }

            $partDanmakus = $this->bilibiliService->getDanmaku($videoPart->cid, intval($videoPart->duration));

            Log::info('Download danmaku success', [
                'id'       => $videoPart->cid,
                'title'    => $videoPart->part,
                'count'    => count($partDanmakus),
                'video_id' => $video->id,
                'bvid'     => $video->bvid,
                'title'    => $video->title,
            ]);
            $this->saveDanmaku($videoPart->cid, $partDanmakus);

            $videoPart->danmaku_downloaded_at = Carbon::now();
            $videoPart->save();
        } catch (\Throwable $e) {
            Log::error('Download danmaku failed', ['id' => $videoPart->cid, 'title' => $videoPart->part, 'error' => $e->getMessage()]);
            return;
        } finally {
            // 移除下载锁
            redis()->del(sprintf('danmaku_downloading:%s', $videoPart->cid));
        }
    }

    public function getVideos(array $conditions = []): array
    {
        $query = Video::query();
        if (isset($conditions['invalid'])) {
            $query->where('invalid', $conditions['invalid']);
        }
        if (isset($conditions['frozen'])) {
            $query->where('frozen', $conditions['frozen']);
        }
        return $query->get()->toArray();
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
}
