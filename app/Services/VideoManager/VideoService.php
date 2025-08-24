<?php
namespace App\Services\VideoManager;

use App\Models\Video;
use App\Models\VideoPart;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use App\Services\DownloadVideoService;
use Illuminate\Database\Eloquent\Collection;

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
}