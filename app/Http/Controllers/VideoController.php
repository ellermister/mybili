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
        $stat = [
            'count'      => count($keys),
            'downloaded' => $downloaded,
            'invalid'    => 0,
            'valid'      => 0,
            'frozen'     => 0,
        ];
        foreach ($keys as $vKey) {
            $result = redis()->get($vKey);
            $vInfo  = json_decode($result, true);
            if ($vInfo) {
                $vInfo['downloaded'] = !!redis()->hExists('video_downloaded', $vInfo['id']);

                $vInfo['invalid'] = video_has_invalid($vInfo);
                $vInfo['valid']   = !video_has_invalid($vInfo);

                $list[] = $vInfo;

                $stat['invalid'] += $vInfo['invalid'] ? 1 : 0;
                $stat['valid'] += $vInfo['valid'] ? 1 : 0;
                $stat['frozen'] += $vInfo['frozen'] ? 1 : 0;
            }
        }

        usort($list, function ($a, $b) {
            if ($a['fav_time'] == $b['fav_time']) {
                return 0;
            }
            return $a['fav_time'] > $b['fav_time'] ? -1 : 1;
        });

        $data = [
            'data' => $list,
            'stat' => $stat,
        ];

        return response()->json($data, 200, [], JSON_UNESCAPED_UNICODE);
    }
}
