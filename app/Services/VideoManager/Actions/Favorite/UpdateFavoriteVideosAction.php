<?php
namespace App\Services\VideoManager\Actions\Favorite;

use App\Events\UpperTryUpdated;
use App\Events\VideoUpdated;
use App\Models\FavoriteList;
use App\Models\Video;
use App\Services\BilibiliService;
use App\Services\VideoManager\Traits\CacheableTrait;
use App\Services\VideoManager\Traits\VideoDataTrait;
use Arr;
use DB;
use Illuminate\Support\Facades\Log;

class UpdateFavoriteVideosAction
{
    use CacheableTrait, VideoDataTrait;

    public function __construct(
        public BilibiliService $bilibiliService,
    ) {
    }

    /**
     * 更新收藏夹视频列表
     */
    public function execute(array $fav, ?int $page = null): void
    {
        $favId  = $fav['id'];
        $videos = $this->bilibiliService->pullFavVideoList($favId, $page);

        if (count($videos) === 0) {
            Log::info('No videos found in fav', ['favId' => $favId]);
            return;
        }

        $videos = array_map(function ($item) {
            $videoInvalid = $this->videoIsInvalid($item);

            $exist = Video::withTrashed()->where('id', $item['id'])->first();

            // 如果视频已经删除即忽略
            if($exist && $exist->trashed()){
                return null;
            }

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

            $upperId = $newValue['upper']['mid'] ?? ($exist['upper_id'] ?? null);

            if($newValue['upper']){
                event(new UpperTryUpdated($newValue['upper']));
            }

            //在此做键值对映射，避免字段未来变更
            return [
                'id'       => $item['id'],
                'link'     => $newValue['link'],
                'title'    => $newValue['title'],
                'intro'    => $newValue['intro'],
                'cover'    => $newValue['cover'],
                'bvid'     => $newValue['bvid'],
                'pubtime'  => date('Y-m-d H:i:s', $newValue['pubtime']),
                'duration' => $newValue['duration'],
                'attr'     => $newValue['attr'],
                'invalid'  => $newValue['invalid'],
                'frozen'   => $newValue['frozen'],
                'page'     => $newValue['page'],
                'fav_time' => date('Y-m-d H:i:s', $newValue['fav_time']),
                'upper_id' => $upperId,
            ];
        }, $videos);

        $videos = array_filter($videos);
        if(empty($videos)){
            return;
        }

        // 暂存视频数据
        $this->saveFavoriteVideo($favId, $videos);

        DB::transaction(function () use ($videos, $favId, $fav) {
            $remoteVideoIds = array_column($videos, 'id');
            $localVideoIds  = DB::table('favorite_list_videos')
                ->where('favorite_list_id', $favId)
                ->pluck('video_id')
                ->toArray();

            $addVideoIds = array_diff($remoteVideoIds, $localVideoIds);

            // 添加新的视频关联关系
            if (! empty($addVideoIds)) {
                $attachData = [];
                foreach ($videos as $video) {
                    if (in_array($video['id'], $addVideoIds)) {
                        $attachData[$video['id']] = [
                            'created_at' => $video['fav_time'],
                            'updated_at' => $video['fav_time'],
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

                $oldVideoData = $video->getAttributes();

                // 检查视频数据是否真正发生了变化
                $hasChanges    = false;
                $changedFields = [];

                foreach ($item as $field => $value) {
                    if (! isset($oldVideoData[$field]) || $oldVideoData[$field] != $value) {
                        $hasChanges            = true;
                        $changedFields[$field] = [
                            'old' => $oldVideoData[$field] ?? null,
                            'new' => $value,
                        ];
                    }
                }

                $video->fill($item);
                $video->save();
                
                // 只有在数据真正发生变化时才触发事件
                if ($hasChanges) {
                    Log::info('Video data changed, triggering VideoUpdated event', [
                        'id'             => $item['id'],
                        'title'          => $item['title'],
                        'changed_fields' => array_keys($changedFields),
                    ]);
                    event(new VideoUpdated($oldVideoData, $video->getAttributes()));
                } else {
                    Log::info('Video data unchanged, skipping VideoUpdated event', [
                        'id'    => $item['id'],
                        'title' => $item['title'],
                    ]);
                }

                Log::info('Update video success', ['id' => $item['id'], 'title' => $item['title']]);
            }

            $savedVideos = $this->getFavoriteVideo($favId);
            // media_count 是不准确的，可能是用户手动取消收藏视频系统没更新、也可能是哔哩哔哩自己故意把用户视频搞丢失
            // 经过测试，用户收藏夹的视频获取没有错误，但收藏夹的media_count 是错误的，目前没有合适的方法更新
            if (intval($fav['media_count']) == count($savedVideos)) {
                $remoteCacheVideoIds = array_column($savedVideos, 'id');
                $deleteVideoIds      = array_diff($localVideoIds, $remoteCacheVideoIds);
                if (! empty($deleteVideoIds)) {
                    FavoriteList::query()->where('id', $favId)->first()->videos()->detach($deleteVideoIds);
                    Log::info('Detached videos from favorite list', ['favId' => $favId, 'videoIds' => $deleteVideoIds]);
                }
            }
        });
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

}
