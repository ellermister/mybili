<?php
namespace App\Http\Controllers;

use App\Models\DownloadQueue;
use App\Services\DownloadQueueService;
use Illuminate\Http\Request;

class DownloadQueueController extends Controller
{
    public function __construct(private DownloadQueueService $downloadQueueService)
    {
    }

    /**
     * 获取队列列表（支持按 status 过滤、分页）
     */
    public function index(Request $request)
    {
        $status  = $request->query('status', 'pending,running');
        $page    = (int) $request->query('page', 1);
        $perPage = (int) $request->query('per_page', 30);

        $statuses = array_filter(explode(',', $status));

        $query = DownloadQueue::with(['video:id,title,cover,type,bvid'])
            ->when(! empty($statuses), fn ($q) => $q->whereIn('status', $statuses))
            ->orderByRaw("CASE status WHEN 'running' THEN 0 WHEN 'pending' THEN 1 ELSE 2 END")
            ->orderByDesc('priority')
            ->orderBy('id');

        $total = (clone $query)->count();
        $items = $query->offset(($page - 1) * $perPage)->limit($perPage)->get();

        $list = $items->map(function (DownloadQueue $item) {
            return [
                'id'           => $item->id,
                'type'         => $item->type,
                'video_id'     => $item->video_id,
                'video_part_id'=> $item->video_part_id,
                'status'       => $item->status,
                'priority'     => $item->priority,
                'error_msg'    => $item->error_msg,
                'scheduled_at' => $item->scheduled_at,
                'completed_at' => $item->completed_at,
                'created_at'   => $item->created_at,
                'video_title'  => $item->video?->title,
                'video_cover'  => $item->video?->cover_info,
                'video_bvid'   => $item->video?->bvid,
                'video_type'   => $item->video?->type,
            ];
        });

        return response()->json([
            'list'  => $list,
            'total' => $total,
            'stat'  => $this->downloadQueueService->getStat(),
        ]);
    }

    /**
     * 取消待下载任务
     */
    public function cancel(int $id)
    {
        $ok = $this->downloadQueueService->cancel($id);
        return response()->json(['success' => $ok]);
    }

    /**
     * 重试失败/取消的任务
     */
    public function retry(int $id)
    {
        $ok = $this->downloadQueueService->retry($id);
        return response()->json(['success' => $ok]);
    }

    /**
     * 调整优先级
     */
    public function priority(Request $request, int $id)
    {
        $data     = $request->validate(['priority' => 'required|integer|min:0|max:9999']);
        $ok       = $this->downloadQueueService->setPriority($id, $data['priority']);
        return response()->json(['success' => $ok]);
    }

    /**
     * 队列统计信息
     */
    public function stat()
    {
        return response()->json($this->downloadQueueService->getStat());
    }
}
