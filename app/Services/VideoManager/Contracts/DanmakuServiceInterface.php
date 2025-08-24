<?php
namespace App\Services\VideoManager\Contracts;

use App\Models\Video;

interface DanmakuServiceInterface
{
    public function getVideoDanmakuCount(Video $video): int;
    public function getDanmaku(string $cid): array;
}