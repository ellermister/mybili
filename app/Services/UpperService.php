<?php
namespace App\Services;

use App\Models\Upper;

class UpperService
{
    public function __construct(public CoverService $coverService)
    {
    }

    public function getUpperInfo(int $mid)
    {
        return Upper::query()->where('mid', $mid)->first();
    }

    public function saveUpperInfo(int $mid, string $name, string $face)
    {
        $upper = Upper::query()->firstOrNew(['mid' => $mid]);
        if (str_contains($name, '注销')) {
            if (str_contains(strval($upper->name ?? ''), '注销')) {
                return;
            }

            $upper->name = "$name (注销)";
            $upper->save();
            return;
        } else {
            $upper->name = $name;
            $upper->face = $face;
            $upper->save();

            $this->coverService->downloadCoverImageJob($face, 'avatar', $upper);
        }
        return $upper;
    }
}
