<?php
namespace App\Services\VideoManager;

use App\Events\VideoUpdated;
use App\Models\Danmaku;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\DownloadVideoService;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Illuminate\Database\Eloquent\Collection;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Storage;
use Log;

class VideoService implements VideoServiceInterface
{

    public $ttl = 86400; // 1 day

    public function count(): int
    {
        return Video::count();
    }

    public function getVideoInfo(string $id, bool $withParts = false): ?Video
    {
        $video = Video::query()->where('id', $id)->first();

        if ($video && $withParts) {
            $video->load('parts');
        }
        return $video;
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
        // 默认按时间逆序排列：优先使用 fav_time，如果不存在或为 null 则使用 created_at
        $query->orderByRaw('COALESCE(fav_time, created_at) DESC');
        return $query->get();
    }

    public function getVideosByPage(array $conditions = [], $page = 1, $perPage = 10): array
    {
        $query = Video::query();
        if (isset($conditions['query'])) {
            if (str_starts_with(strtolower($conditions['query']), 'bv')) {
                $query->where('bvid', $conditions['query']);
            } else if (preg_match('/^\d+$/', $conditions['query'])) {
                $query->where('id', $conditions['query']);
            } else {
                $query->where('title', 'like', '%' . $conditions['query'] . '%');
            }
        }

        if (isset($conditions['status'])) {
            if ($conditions['status'] == 'valid') {
                $query->where('invalid', 0);
            } else if ($conditions['status'] == 'invalid') {
                $query->where('invalid', 1);
            } else if ($conditions['status'] == 'frozen') {
                $query->where('frozen', 1);
            }
        }

        if (isset($conditions['downloaded'])) {
            if ($conditions['downloaded'] == 'yes') {
                $query->where('video_downloaded_num', '>', 0);
            } else if ($conditions['downloaded'] == 'no') {
                $query->where('video_downloaded_num', 0);
            }
        }

        if (isset($conditions['multi_part'])) {
            if ($conditions['multi_part'] == 'yes') {
                $query->whereHas('parts', function ($query) {
                    $query->where('page', '>', 1);
                });
            } else if ($conditions['multi_part'] == 'no') {
                $query->whereDoesntHave('parts', function ($query) {
                    $query->where('page', 1);
                });
            }
        }

        if (isset($conditions['fav_id'])) {
            if (intval($conditions['fav_id']) > 0) {
                $query->whereHas('favorite', function ($query) use ($conditions) {
                    $query->where('id', $conditions['fav_id']);
                });
            } else if (intval($conditions['fav_id']) < 0) {
                $query->whereHas('subscriptions', function ($query) use ($conditions) {
                    $query->where('id', abs($conditions['fav_id']));
                });
            }
        }

        $stat = [
            'count'      => (clone $query)->count(),
            'downloaded' => (clone $query)->where('video_downloaded_num', '>', 0)->count(),
            'invalid'    => (clone $query)->where('invalid', 1)->count(),
            'valid'      => (clone $query)->where('invalid', 0)->count(),
            'frozen'     => (clone $query)->where('frozen', 1)->count(),
        ];
        return [
            'list' => $query->offset(($page - 1) * $perPage)->limit($perPage)->get(),
            'stat' => $stat,
        ];
    }

    public function getAllPartsVideoForUser(Video $video): array
    {
        if ($video->isAudio()) {
            $audioPart = $video->audioPart;
            if (! $audioPart) {
                return [];
            }
            return [[
                'id'         => $audioPart->sid,
                'part'       => 1,
                'url'        => $audioPart->audio_download_url,
                'title'      => $video->title,
                'downloaded' => (bool) $audioPart->audio_download_path,
            ]];
        }

        $list = [];
        foreach ($video->parts as $videoPart) {
            $list[] = [
                'id'         => $videoPart['cid'],
                'part'       => $videoPart['page'],
                'url'        => $videoPart['video_download_url'] ?: null,
                'title'      => $videoPart['part'] ?? 'P' . $videoPart['page'],
                'downloaded' => (bool) $videoPart['video_download_path'],
            ];
        }
        return $list;
    }

    public function getVideosStat(array $conditions = []): array
    {
        // 使用单次查询合并所有 COUNT，减少数据库往返
        $stats = Video::selectRaw('
            COUNT(*) as count,
            SUM(CASE WHEN video_downloaded_num > 0 THEN 1 ELSE 0 END) as downloaded,
            SUM(CASE WHEN invalid = 1 THEN 1 ELSE 0 END) as invalid,
            SUM(CASE WHEN invalid = 0 THEN 1 ELSE 0 END) as valid,
            SUM(CASE WHEN frozen = 1 THEN 1 ELSE 0 END) as frozen
        ')->first();

        return [
            'count'      => (int) $stats->count,
            'downloaded' => (int) $stats->downloaded,
            'invalid'    => (int) $stats->invalid,
            'valid'      => (int) $stats->valid,
            'frozen'     => (int) $stats->frozen,
        ];
    }

    public function getVideoPartFileSize(VideoPart $videoPart): int
    {
        $filePath = app(DownloadVideoService::class)->getVideoPartValidFilePath($videoPart);
        if ($filePath) {
            return filesize($filePath);
        }
        return 0;
    }

    public function deleteVideos(array $ids): array
    {
        $deletedIds = [];
        $videos     = Video::query()->whereIn('id', $ids)->get();

        foreach ($videos as $video) {
            $video->parts->each(function (VideoPart $videoPart) {
                app(DownloadVideoService::class)->deleteVideoPartFile($videoPart);
            });
            if ($video->delete()) {
                $deletedIds[] = $video->id;
                event(new VideoUpdated($video->getAttributes(), []));
            }
        }
        Log::info(sprintf('Delete %d videos', count($deletedIds)), ['ids' => $ids, 'deleted_ids' => $deletedIds]);

        // 删除视频弹幕
        Danmaku::query()->whereIn('video_id', $deletedIds)->delete();
        return $deletedIds;
    }

    public function getFavVideosLightweight(int $favId): array
    {
        $cacheKey = "fav_videos:{$favId}";
        $cached = redis()->get($cacheKey);
        if ($cached) {
            return json_decode($cached, true);
        }

        $videos = $this->queryFavVideosLightweight($favId);

        redis()->set($cacheKey, json_encode($videos));
        redis()->expire($cacheKey, 600);

        return $videos;
    }

    public function getSubVideosLightweight(int $subId): array
    {
        $cacheKey = "sub_videos:{$subId}";
        $cached = redis()->get($cacheKey);
        if ($cached) {
            return json_decode($cached, true);
        }

        $videos = $this->querySubVideosLightweight($subId);

        redis()->set($cacheKey, json_encode($videos));
        redis()->expire($cacheKey, 600);

        return $videos;
    }

    private function queryFavVideosLightweight(int $favId): array
    {
        $rows = DB::table('videos')
            ->join('favorite_list_videos', 'videos.id', '=', 'favorite_list_videos.video_id')
            ->leftJoin('coverables', function ($join) {
                $join->on('coverables.coverable_id', '=', 'videos.id')
                     ->where('coverables.coverable_type', '=', Video::class);
            })
            ->leftJoin('covers', 'covers.id', '=', 'coverables.cover_id')
            ->where('favorite_list_videos.favorite_list_id', $favId)
            ->select([
                'videos.id', 'videos.title', 'videos.bvid',
                'videos.pubtime', 'videos.fav_time', 'videos.page',
                'videos.video_downloaded_num', 'videos.audio_downloaded_num',
                'videos.frozen', 'videos.invalid', 'videos.cover',
                'videos.created_at',
                'covers.path as cover_path',
            ])
            ->orderByDesc('videos.fav_time')
            ->orderByDesc('videos.created_at')
            ->get();

        return $rows->map(fn ($row) => $this->transformLightweightVideo($row))->toArray();
    }

    private function querySubVideosLightweight(int $subId): array
    {
        $rows = DB::table('videos')
            ->join('subscription_videos', 'videos.id', '=', 'subscription_videos.video_id')
            ->leftJoin('coverables', function ($join) {
                $join->on('coverables.coverable_id', '=', 'videos.id')
                     ->where('coverables.coverable_type', '=', Video::class);
            })
            ->leftJoin('covers', 'covers.id', '=', 'coverables.cover_id')
            ->where('subscription_videos.subscription_id', $subId)
            ->select([
                'videos.id', 'videos.title', 'videos.bvid',
                'videos.pubtime', 'videos.fav_time', 'videos.page',
                'videos.video_downloaded_num', 'videos.audio_downloaded_num',
                'videos.frozen', 'videos.invalid', 'videos.cover',
                'videos.created_at',
                'covers.path as cover_path',
            ])
            ->orderByDesc('videos.pubtime')
            ->orderByDesc('videos.created_at')
            ->get();

        return $rows->map(fn ($row) => $this->transformLightweightVideo($row))->toArray();
    }

    private function transformLightweightVideo(object $row): array
    {
        $item = (array) $row;
        $item['cover_image_url'] = $item['cover_path']
            ? Storage::url($item['cover_path'])
            : null;
        unset($item['cover_path']);
        $item['pubtime'] = $item['pubtime'] ? strtotime($item['pubtime']) : null;
        $item['fav_time'] = $item['fav_time'] ? strtotime($item['fav_time']) : null;
        return $item;
    }

    public function updateVideosCache(?array $videos = null): void
    {
        if ($videos && is_array($videos) && count($videos) > 0) {
            $list = $videos;
        } else {
            $list = $this->getVideos()->toArray();
        }
        redis()->set('video_list', json_encode($list));
        redis()->expire('video_list', $this->ttl);
        Log::info('Update videos cache success', ['count' => count($list)]);
    }

    public function getVideosCache(): array
    {
        $list = redis()->get('video_list');
        if ($list) {
            return json_decode($list, true);
        }
        $videos = $this->getVideos()->values()->toArray();
        $this->updateVideosCache($videos);
        return $videos;
    }
}
