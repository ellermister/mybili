<?php
namespace App\Services\VideoManager\Contracts;

use App\Models\Video;
use App\Models\VideoPart;
use Illuminate\Database\Eloquent\Collection;

interface VideoServiceInterface
{
    public function getVideoInfo(string $id, bool $withParts = false): ?Video;

    public function getVideos(array $conditions = []): Collection;

    public function getAllPartsVideoForUser(Video $video): array;

    public function getVideosStat(array $conditions = []): array;

    public function getVideoPartFileSize(VideoPart $videoPart): int;

    public function count(): int;

}