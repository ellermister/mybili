<?php
namespace App\Services\VideoManager;

use App\Models\Danmaku;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\DownloadVideoService;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Illuminate\Database\Eloquent\Collection;
use Log;

class VideoService implements VideoServiceInterface
{

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
        $videos = Video::query()->whereIn('id', $ids)->get();

        foreach ($videos as $video) {
            $video->parts->each(function (VideoPart $videoPart) {
                app(DownloadVideoService::class)->deleteVideoPartFile($videoPart);
            });
            if($video->delete()){
                $deletedIds[] = $video->id;
            }
        }
        Log::info(sprintf('Delete %d videos', count($deletedIds)), ['ids' => $ids, 'deleted_ids' => $deletedIds]);

        // 删除视频弹幕
        Danmaku::query()->whereIn('video_id', $deletedIds)->delete();
        return $deletedIds;
    }
}
