<?php
namespace App\Services\VideoManager\Traits;

trait VideoDataTrait
{
    public function videoIsInvalid(array $video): bool
    {
        return ($video['attr'] ?? 0) > 0 || ($video['title'] ?? '') == '已失效视频';
    }
}
