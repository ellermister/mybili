<?php
namespace App\Services\VideoManager\Actions\Video;

use App\Contracts\DownloadImageServiceInterface as ContractsDownloadImageServiceInterface;
use App\Events\VideoUpdated;
use App\Jobs\DownloadCoverImageJob;
use App\Models\Video;
use App\Services\BilibiliService;
use Illuminate\Support\Facades\Log;

class FixFavoriteInvalidVideoAction
{

    public $ttl = 3600;

    public function __construct(
        public BilibiliService $bilibiliService,
        public ContractsDownloadImageServiceInterface $downloadImageServiceInterface
    ) {
    }

    /**
     * 修复收藏夹的无效视频
     * @param int $favId
     * @return void
     */
    public function execute(int $favId, int $page = 1): void
    {
        $videos = cache()->remember('fav_folder_videos_' . $favId . '_page_' . $page, $this->ttl, function () use ($favId, $page) {
            return $this->bilibiliService->getFavFolderResources($favId, $page);
        });

        $count = 0;
        foreach ($videos as $video) {
            if (! $video['is_invalid']) {
                continue;
            }

            $videoId    = $video['oid'];
            $videoTitle = $video['title'];
            $videoCover = $video['cover'];

            // 检测是否污染
            if ($videoTitle == '已失效视频' || str_contains($videoCover, "be27fd62c99036dce67efface486fb0a88ffed06")) {
                continue;
            }

            $existVideo = Video::query()->where('id', $videoId)->first();
            if ($existVideo) {
                $oldVideoData        = $existVideo->getAttributes();
                $existVideo->invalid = true;
                $existVideo->fill([
                    'title' => $videoTitle,
                    'cover' => $videoCover,
                ]);
                $existVideo->save();
                Log::info('Fix fav invalid video', ['id' => $videoId, 'title' => $videoTitle]);
                $count++;

                if ($oldVideoData['cover'] != $videoCover) {
                    dispatch(new DownloadCoverImageJob($videoCover, 'video', $existVideo));
                }

                event(new VideoUpdated($oldVideoData, $existVideo->getAttributes()));
            }
        }
        Log::info('Fix fav invalid video count', ['count' => $count]);
    }
}
