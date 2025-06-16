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
            $video->video_parts = $this->videoManagerService->getAllPartsVideoForUser($video);
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
            abort(400);
        }
        $result = $this->videoManagerService->getDanmaku($cid);
        $result = array_map(function ($item) {
            return [
                ($item['progress'] ?? 0) / 1000,
                $item['mode'] ?? 0,
                $item['color'] ?? 0,
                '',
                $item['content'] ?? '',
            ];
        }, $result);
        return response()->json([
            'code' => 0,
            'data' => $result,
        ]);
    }
}
