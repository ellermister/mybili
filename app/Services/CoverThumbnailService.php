<?php

namespace App\Services;

use App\Models\Cover;
use Illuminate\Support\Facades\File;
use Illuminate\Support\Facades\Log;

class CoverThumbnailService
{
    private int $maxWidth;

    private int $maxHeight;

    public function __construct()
    {
        $this->maxWidth = max(1, (int) config('cover_thumbnail.max_width', 320));
        $this->maxHeight = max(1, (int) config('cover_thumbnail.max_height', 320));
    }

    /**
     * 封面 path（相对 storage/app/public）对应的缩略图相对路径，例如 images/a.png -> thumbs/images/a.png
     */
    public function thumbRelativePathForCoverPath(string $coverPath): string
    {
        $normalized = ltrim(str_replace('\\', '/', $coverPath), '/');

        return 'thumbs/'.$normalized;
    }

    public function absolutePublicDiskPath(string $relativePath): string
    {
        return storage_path('app/public/'.ltrim(str_replace('\\', '/', $relativePath), '/'));
    }

    /**
     * 为单条封面记录生成缩略图；成功则写入 thumbnail_generated_at。
     *
     * @return bool 是否已具备缩略图（含此前已生成）
     */
    public function generateForCover(Cover $cover): bool
    {
        if ($cover->thumbnail_generated_at !== null) {
            return true;
        }

        $path = $cover->path;
        if ($path === null || $path === '') {
            return false;
        }

        $source = $this->absolutePublicDiskPath($path);
        if (! is_file($source)) {
            Log::warning('Cover thumbnail: source file missing', ['cover_id' => $cover->id, 'path' => $source]);

            return false;
        }

        $thumbRelative = $this->thumbRelativePathForCoverPath($path);
        $dest = $this->absolutePublicDiskPath($thumbRelative);

        if (! $this->resizeImageToFile($source, $dest, $this->maxWidth, $this->maxHeight)) {
            Log::warning('Cover thumbnail: resize failed', ['cover_id' => $cover->id, 'source' => $source]);

            return false;
        }

        $cover->forceFill(['thumbnail_generated_at' => now()])->save();

        return true;
    }

    /**
     * 批量处理尚未生成缩略图的封面（用于定时任务）。
     *
     * @return int 本次成功生成并写入字段的条数（失败或源文件缺失不计入）
     */
    public function syncMissingThumbnails(int $limit = 500): int
    {
        $newCount = 0;
        Cover::query()
            ->whereNull('thumbnail_generated_at')
            ->whereNotNull('path')
            ->orderBy('id')
            ->limit($limit)
            ->get()
            ->each(function (Cover $cover) use (&$newCount) {
                if ($this->generateForCover($cover)) {
                    $newCount++;
                }
            });

        return $newCount;
    }

    /**
     * 将源图缩放写入目标路径（等比，不超过 max，不放大）。
     */
    public function resizeImageToFile(string $sourcePath, string $destPath, int $maxWidth, int $maxHeight): bool
    {
        if (! extension_loaded('gd')) {
            return false;
        }

        $maxWidth = max(1, $maxWidth);
        $maxHeight = max(1, $maxHeight);

        $info = @getimagesize($sourcePath);
        if ($info === false) {
            return false;
        }

        [$width, $height, $type] = $info;

        if ($width <= 0 || $height <= 0) {
            return false;
        }

        $ratio = min($maxWidth / $width, $maxHeight / $height, 1.0);
        $newW = (int) max(1, round($width * $ratio));
        $newH = (int) max(1, round($height * $ratio));

        $src = $this->createImageFromFile($sourcePath, $type);
        if ($src === false) {
            return false;
        }

        $dst = imagecreatetruecolor($newW, $newH);
        if ($dst === false) {
            imagedestroy($src);

            return false;
        }

        if ($type === IMAGETYPE_PNG || $type === IMAGETYPE_GIF) {
            imagealphablending($dst, false);
            imagesavealpha($dst, true);
            $transparent = imagecolorallocatealpha($dst, 0, 0, 0, 127);
            imagefill($dst, 0, 0, $transparent);
        }

        imagecopyresampled($dst, $src, 0, 0, 0, 0, $newW, $newH, $width, $height);
        imagedestroy($src);

        $ok = $this->saveImage($dst, $destPath, $type);
        imagedestroy($dst);

        return $ok;
    }

    /**
     * @return \GdImage|false
     */
    private function createImageFromFile(string $path, int $type)
    {
        return match ($type) {
            IMAGETYPE_JPEG => imagecreatefromjpeg($path),
            IMAGETYPE_PNG => imagecreatefrompng($path),
            IMAGETYPE_GIF => imagecreatefromgif($path),
            IMAGETYPE_WEBP => function_exists('imagecreatefromwebp') ? imagecreatefromwebp($path) : false,
            default => false,
        };
    }

    private function saveImage(\GdImage $image, string $destPath, int $type): bool
    {
        $dir = dirname($destPath);
        if (! is_dir($dir)) {
            File::makeDirectory($dir, 0755, true);
        }

        return match ($type) {
            IMAGETYPE_JPEG => imagejpeg($image, $destPath, 85),
            IMAGETYPE_PNG => imagepng($image, $destPath, 6),
            IMAGETYPE_GIF => imagegif($image, $destPath),
            IMAGETYPE_WEBP => function_exists('imagewebp') ? imagewebp($image, $destPath, 85) : false,
            default => false,
        };
    }
}
