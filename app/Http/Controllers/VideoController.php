<?php
namespace App\Http\Controllers;

use App\Services\DanmakuConverterService;
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
        public DanmakuConverterService $danmakuConverterService
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
            'load_all'   => 'nullable|boolean',
        ]);
        $loadAll = filter_var($data['load_all'] ?? false, FILTER_VALIDATE_BOOLEAN);
        $filters = [
            'query'      => $data['query'] ?? '',
            'status'     => $data['status'] ?? '',
            'downloaded' => $data['downloaded'] ?? '',
            'multi_part' => $data['multi_part'] ?? '',
            'fav_id'     => $data['fav_id'] ?? '',
        ];

        if ($loadAll) {
            // 如果需要加载全部，直接获取所有数据
            $videos = $this->videoService->getVideos($filters);
            return response()->json([
                'stat' => $this->videoService->getVideosStat($filters),
                'list' => $videos,
            ]);
        } else {
            // 否则使用分页
            $page = $data['page'] ?? 1;
            $perPage = intval($data['page_size'] ?? 30);
            $result = $this->videoService->getVideosByPage($filters, $page, $perPage);

            return response()->json([
                'stat' => $result['stat'],
                'list' => $result['list'],
                'pagination' => [
                    'total' => $result['total'] ?? count($result['list']),
                    'per_page' => $perPage,
                    'current_page' => $page,
                    'last_page' => ceil(($result['total'] ?? count($result['list'])) / $perPage),
                ],
            ]);
        }
    }

    public function destroy(Request $request, string $id)
    {
        if (config('services.bilibili.setting_read_only')) {
            abort(403);
        }
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

    /**
     * 获取指定 CID 的弹幕数据（新格式）
     * 
     * @param Request $request
     * @return \Illuminate\Http\JsonResponse
     */
    public function danmaku(Request $request)
    {
        $cid = $request->input('id');
        
        if (!$cid) {
            return response()->json([
                'code'    => 1,
                'message' => 'CID 参数不能为空',
                'data'    => [],
            ]);
        }

        // 获取原始弹幕数据
        $danmakuList = $this->danmakuService->getDanmaku($cid);
        
        // 转换为新格式
        $convertedData = $this->danmakuConverterService->convert($danmakuList);
        
        return response()->json([
            'code' => 0,
            'data' => $convertedData,
        ]);
    }
}
