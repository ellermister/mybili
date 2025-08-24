<?php
namespace App\Services\VideoManager\Actions\Video;

use App\Events\VideoUpdated;
use App\Models\Video;
use App\Services\BilibiliService;
use App\Services\VideoManager\Traits\VideoDataTrait;
use Carbon\Carbon;
use Illuminate\Support\Facades\Log;

class PullVideoInfoAction
{
    use VideoDataTrait;

    public function execute(string $bvid): void
    {
        try {
            $videoInfo = app(BilibiliService::class)->getVideoInfo($bvid);
            $aid       = $videoInfo['aid'];

            $video        = Video::query()->where('id', $aid)->first();
            $oldVideoData = $video?->toArray() ?? [];

            // 如果是已存在的视频且无效，跳过更新
            if ($video && $this->videoIsInvalid($videoInfo)) {
                Log::info('Video is invalid, skip update video info', [
                    'id'    => $aid,
                    'title' => $videoInfo['title'],
                ]);
                $this->markVideoAsInvalidByBvid($bvid);
                return;
            }

            // 创建或更新视频
            if (! $video) {
                $video = new Video(['id' => $aid]);
            }

            $videoData = $this->mapVideoInfoToVideoData($aid, $videoInfo);
            $video->fill($videoData);
            $video->save();

            event(new VideoUpdated($oldVideoData, $video->toArray()));

        } catch (\Exception $e) {
            Log::error("PullVideoInfoJob failed: " . $e->getMessage());

            // 处理404错误，标记视频为无效
            if ($e->getCode() == -404) {
                $this->markVideoAsInvalidByBvid($bvid);
            }
            throw $e;
        }
    }

    /**
     * 将视频信息映射为数据库字段
     */
    private function mapVideoInfoToVideoData(string $aid, array $videoInfo): array
    {
        $isInvalid = $videoInfo['state'] != 0;

        return [
            'link'    => sprintf('bilibili://video/%s', $aid),
            'title'   => $videoInfo['title'],
            'intro'   => $videoInfo['desc'],
            'cover'   => $videoInfo['pic'],
            'bvid'    => $videoInfo['bvid'],
            'pubtime' => Carbon::createFromTimestamp($videoInfo['pubdate']),
            'invalid' => $isInvalid,
            'frozen'  => $isInvalid,
            'page'    => count($videoInfo['pages']),
        ];
    }

    /**
     * 根据BVID标记视频为无效
     */
    private function markVideoAsInvalidByBvid(string $bvid): void
    {
        $video = Video::where('bvid', $bvid)->first();

        if (! $video || $video->invalid) {
            return;
        }

        $oldVideo       = $video->toArray();
        $video->invalid = true;
        $video->frozen  = true;
        $video->save();

        event(new VideoUpdated($oldVideo, $video->toArray()));
    }
}
