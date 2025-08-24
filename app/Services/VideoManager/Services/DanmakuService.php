<?php
namespace App\Services\VideoManager\Services;

use App\Models\Danmaku;
use App\Models\Video;
use App\Services\VideoManager\Contracts\DanmakuServiceInterface;

class DanmakuService implements DanmakuServiceInterface
{
    public function getVideoDanmakuCount(Video $video): int
    {
        $count = 0;
        foreach ($video->parts as $part) {
            $count += Danmaku::where('cid', $part->cid)->count();
        }
        return $count;
    }

    public function getDanmaku(string $cid): array
    {
        return Danmaku::where('cid', $cid)->get()->toArray();
    }
}