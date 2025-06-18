<?php
namespace App\Http\Controllers;

use App\Contracts\VideoManagerServiceInterface;
use Illuminate\Http\Request;

class VideoController extends Controller
{

    public function __construct(public VideoManagerServiceInterface $videoManagerService)
    {

    }

    public function show(Request $request, int $id)
    {
        $video = $this->videoManagerService->getVideoInfo($id, true);
        if ($video) {
            $video->video_parts   = $this->videoManagerService->getAllPartsVideoForUser($video);
            $video->danmaku_count = $this->videoManagerService->getVideoDanmakuCount($video);
            return response()->json($video);
        }
        abort(404);
    }

    public function progress()
    {
        $list = $this->videoManagerService->getVideos();
        usort($list, function ($a, $b) {
            if ($a['fav_time'] == $b['fav_time']) {
                return 0;
            }
            return $a['fav_time'] > $b['fav_time'] ? -1 : 1;
        });

        $data = [
            'data' => $list,
            'stat' => $this->videoManagerService->getVideosStat([]),
        ];

        return response()->json($data, 200, [], JSON_UNESCAPED_UNICODE);
    }

    public function danmaku(Request $request, int $cid)
    {
        $result = $this->videoManagerService->getDanmaku($cid);
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
        $result = $this->videoManagerService->getDanmaku($cid);
        $result = array_map(function ($item) {

            return [
                ($item['progress'] ?? 0) / 1000,
                $this->covertMode($item['mode'] ?? 0),
                $this->covertColor($item['color'] ?? 0),
                '',
                $item['content'] ?? '',
            ];
        }, $result);
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
    }

    protected function covertColor($color)
    {
        //默认为 #ffffff    
        return '#' . str_pad(dechex($color), 6, '0', STR_PAD_LEFT);
    }
}
