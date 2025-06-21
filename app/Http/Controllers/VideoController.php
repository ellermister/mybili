<?php
namespace App\Http\Controllers;

use App\Contracts\VideoManagerServiceInterface;
use App\Services\DPlayerDanmakuService;
use Illuminate\Http\Request;

class VideoController extends Controller
{

    public function __construct(public VideoManagerServiceInterface $videoManagerService, public DPlayerDanmakuService $dplayerDanmakuService)
    {

    }

    public function show(Request $request, int $id)
    {
        $video = $this->videoManagerService->getVideoInfo($id, true);
        if ($video) {
            $video->video_parts   = $this->videoManagerService->getAllPartsVideoForUser($video);
            $video->danmaku_count = $this->videoManagerService->getVideoDanmakuCount($video);
            $video->load('favorite');
            
            return response()->json($video);
        }
        abort(404);
    }

    public function progress()
    {
        $list = $this->videoManagerService->getVideos()
            ->sortByDesc('fav_time')
            ->values()
            ->all();

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
