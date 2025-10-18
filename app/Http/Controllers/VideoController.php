<?php
namespace App\Http\Controllers;

use App\Services\DPlayerDanmakuService;
use App\Services\VideoManager\Contracts\DanmakuServiceInterface;
use App\Services\VideoManager\Contracts\FavoriteServiceInterface;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Carbon\Carbon;
use Illuminate\Http\Request;

class VideoController extends Controller
{

    public function __construct(
        public VideoServiceInterface $videoService,
        public FavoriteServiceInterface $favoriteService,
        public DanmakuServiceInterface $danmakuService,
        public DPlayerDanmakuService $dplayerDanmakuService
    ) {

    }

    public function index(Request $request)
    {
        $data = $request->validate([
            'query'      => 'nullable|string',
            'page'       => 'nullable|integer|min:1',
            'status'     => 'nullable|string',
            'downloaded' => 'nullable|string',
            'multi_part' => 'nullable|string',
            'fav_id'     => 'nullable|integer',
            'page_size'  => 'nullable|integer|min:1',
        ]);
        $page    = $data['page'] ?? 1;
        $perPage = 30;
        $result  = $this->videoService->getVideosByPage([
            'query'      => $data['query'] ?? '',
            'status'     => $data['status'] ?? '',
            'downloaded' => $data['downloaded'] ?? '',
            'multi_part' => $data['multi_part'] ?? '',
            'fav_id'     => $data['fav_id'] ?? '',
        ], $page, intval($data['page_size'] ?? $perPage));
        return response()->json([
            'stat'  => $result['stat'],
            'list'  => $result['list'],
        ]);
    }

    public function destroy(Request $request, string $id)
    {
        // 补充其他ID
        $extend_ids = $request->input('extend_ids');
        if ($extend_ids && is_array($extend_ids)) {
            $ids = array_merge([$id], $extend_ids);
        } else {
            $ids = [$id];
        }
        $ids        = array_map('intval', $ids);
        $deletedIds = $this->videoService->deleteVideos($ids);
        if ($deletedIds) {
            return response()->json([
                'code'        => 0,
                'message'     => 'Video deleted successfully',
                'deleted_ids' => $deletedIds,
            ]);
        } else {
            return response()->json([
                'code'    => 1,
                'message' => 'Video deletion failed',
            ]);
        }
    }

    public function show(Request $request, int $id)
    {
        $video = $this->videoService->getVideoInfo($id, true);
        if ($video) {
            $video->video_parts   = $this->videoService->getAllPartsVideoForUser($video);
            $video->danmaku_count = $this->danmakuService->getVideoDanmakuCount($video);
            $video->load('favorite');
            $video->load('subscriptions');
            $video->load('upper');

            return response()->json($video);
        }
        abort(404);
    }

    public function progress()
    {
        $list = $this->videoService->getVideos()
            ->sortBy(function ($video) {
                // 优先使用 fav_time，如果不存在或为 null 则使用 created_at
                return Carbon::parse($video->fav_time ?? $video->created_at)->timestamp;
            }, SORT_REGULAR, true) // true 表示降序排序
            ->values()
            ->all();

        $data = [
            'data' => $list,
            'stat' => $this->videoService->getVideosStat([]),
        ];

        return response()->json($data, 200, [], JSON_UNESCAPED_UNICODE);
    }

    public function danmaku(Request $request, int $cid)
    {
        $result = $this->danmakuService->getDanmaku($cid);
        return response()->json($result);
    }

    public function danmakuV3(Request $request)
    {
        $cid = $request->input('id');
        if (! $cid) {
            return response()->json([
                'code'    => 0,
                'message' => 'empty cid request',
                'data'    => [],
            ]);
        }
        $result = $this->danmakuService->getDanmaku($cid);
        $result = $this->dplayerDanmakuService->convertDanmaku($result);
        return response()->json([
            'code' => 0,
            'data' => $result,
        ]);
    }

    protected function covertMode($mode)
    {
        // default: right
        // 1：普通弹幕 => right
        // 4：底部弹幕 => bottom
        // 5：顶部弹幕 => top
        // 7：高级弹幕
        switch ($mode) {
            case 1:
                return 'right';
            case 4:
                return 'bottom';
            case 5:
                return 'top';
            default:
                return 'right';
        }
        // number2Type: (number) => {
        //     switch (number) {
        //         case 0:
        //             return 'right';
        //         case 1:
        //             return 'top';
        //         case 2:
        //             return 'bottom';
        //         default:
        //             return 'right';
        //     }
        // },
    }

    protected function covertColor($color)
    {
        //默认为 #ffffff
        return '#' . str_pad(dechex($color), 6, '0', STR_PAD_LEFT);
    }
}
