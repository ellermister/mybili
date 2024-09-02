<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

class VideoController extends Controller
{
    public function show(Request $request, int $id)
    {
        $result = redis()->get(sprintf('video:%d', $id));
        if ($result) {
            $data = json_decode($result, true);
            if ($data) {
                return response()->json($data);
            }
        }
        abort(404);
    }

    public function progress()
    {
        // laravel_horizon:job:App\Jobs\DownloadVideoJob
        $iterator = null;
        $keys     = [
        ];
        do {
            $result = redis()->scan($iterator, 'video:*', 50);
            $keys   = array_merge($keys, $result);
        } while ($iterator != 0);

        $downloaded = redis()->hlen('video_downloaded');

        $list = [];
        foreach ($keys as $vKey) {
            $result = redis()->get($vKey);
            $vInfo  = json_decode($result, true);
            if ($vInfo) {
                $vInfo['downloaded'] = !!redis()->hExists('video_downloaded', $vInfo['id']);
                $list[]              = $vInfo;
            }
        }

        usort($list, function ($a, $b) {
            if ($a['fav_time'] == $b['fav_time']) {
                return 0;
            }
            return $a['fav_time'] > $b['fav_time'] ? -1 : 1;
        });

        $data = [
            'count'      => count($keys),
            'downloaded' => $downloaded,
            'data'       => $list,
        ];

        return response()->json($data, 200, [], JSON_UNESCAPED_UNICODE);
    }
}
