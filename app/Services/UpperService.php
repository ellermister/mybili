<?php
namespace App\Services;

use App\Models\Upper;

class UpperService
{
    public function getUpperInfo(int $mid)
    {
        return Upper::query()->where('mid', $mid)->first();
    }

    public function saveUpperInfo(int $mid, string $name, string $face)
    {
        return Upper::query()->updateOrCreate(['mid' => $mid], ['name' => $name, 'face' => $face]);
    }
}