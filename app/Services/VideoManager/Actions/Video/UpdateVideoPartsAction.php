<?php
namespace App\Services\VideoManager\Actions\Video;

use App\Events\VideoPartUpdated;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\BilibiliService;
use App\Services\VideoManager\Traits\VideoDataTrait;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class UpdateVideoPartsAction
{

    public function __construct(
        public BilibiliService $bilibiliService
    ) {
    }

    use VideoDataTrait;

    /**
     * 更新视频分P信息
     */
    public function execute(Video $video): void
    {
        if ($this->videoIsInvalid($video->toArray())) {
            Log::info('Video is invalid, skip update video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        try {
            if (config('services.bilibili.id_type') == 'bv') {
                $videoParts = $this->bilibiliService->getVideoParts($video->bvid);
            } else {
                $videoParts = $this->bilibiliService->getVideoParts($video->id);
            }
        } catch (\Exception $e) {
            Log::error('Get video parts failed', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        // 找出本地多余远端的数据
        $localVideoParts     = VideoPart::where('video_id', $video->id)->orderBy('page', 'asc')->get()->toArray();
        $localVideoPartsIds  = array_column($localVideoParts, 'cid');
        $remoteVideoPartsIds = array_column($videoParts, 'cid');

        $deleteVideoPartsIds = array_diff($localVideoPartsIds, $remoteVideoPartsIds);

        if (! empty($deleteVideoPartsIds)) {
            Log::info('Delete video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title, 'deleteVideoPartsIds' => $deleteVideoPartsIds]);
            return;
        }

        DB::transaction(function () use ($videoParts, $video) {
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

}
