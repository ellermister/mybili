<?php
namespace App\Services;

use App\Models\Cover;
use App\Models\Coverables;
use Illuminate\Database\Eloquent\Model;
use App\Jobs\DownloadCoverImageJob;

class CoverService extends DownloadImageService
{
    

    public function downloadCoverImageJob(string $url, string $type, Model $model): void
    {
        dispatch(new DownloadCoverImageJob($url, $type, $model));
    }

    /**
     * 下载封面并创建关联关系
     * 
     * @param string $url 图片URL
     * @param string $type 封面类型 (video, avatar, favorite)
     * @param Model $model 关联的模型实例（Video、FavoriteList、Upper等）
     * @return Cover
     * @throws \Exception
     */
    public function downloadCover(string $url, string $type, Model $model): Cover
    {
        // 1. 检查是否已存在该封面
        $filename = $this->convertToFilename($url);
        $cover = Cover::where('filename', $filename)->first();
        
        // 2. 如果封面不存在，下载并创建记录
        if (!$cover) {
            $localPath = $this->getImageLocalPath($url);
            
            // 下载图片
            if (!is_file($localPath)) {
                $this->downloadImage($url, $localPath);
            }
            
            // 获取图片信息
            $imageInfo = $this->getImageInfo($localPath);
            
            // 创建封面记录
            $cover = Cover::create([
                'url'       => $url,
                'type'      => $type,
                'filename'  => $filename,
                'path'      => get_relative_path($localPath),
                'mime_type' => $imageInfo['mime_type'],
                'size'      => $imageInfo['size'],
                'width'     => $imageInfo['width'],
                'height'    => $imageInfo['height'],
            ]);
        }
        
        // 3. 检查关联关系是否存在
        $coverableExists = Coverables::where([
            'cover_id'        => $cover->id,
            'coverable_id'    => $model->id,
            'coverable_type'  => get_class($model),
        ])->exists();
        
        // 4. 不存在则创建关联
        if (!$coverableExists) {
            Coverables::create([
                'cover_id'        => $cover->id,
                'coverable_id'    => $model->id,
                'coverable_type'  => get_class($model),
            ]);
        }
        
        return $cover;
    }
    
    /**
     * 获取图片详细信息
     * 
     * @param string $filePath 图片本地路径
     * @return array
     * @throws \Exception
     */
    protected function getImageInfo(string $filePath): array
    {
        if (!is_file($filePath)) {
            throw new \Exception("Image file not found: {$filePath}");
        }
        
        // 获取图片尺寸和类型信息
        $imageSize = @getimagesize($filePath);
        if ($imageSize === false) {
            throw new \Exception("Failed to get image size for: {$filePath}");
        }
        
        // 获取文件大小（字节）
        $fileSize = filesize($filePath);
        if ($fileSize === false) {
            throw new \Exception("Failed to get file size for: {$filePath}");
        }
        
        return [
            'width'     => $imageSize[0] ?? 0,
            'height'    => $imageSize[1] ?? 0,
            'mime_type' => $imageSize['mime'] ?? 'image/jpeg',
            'size'      => $fileSize,
        ];
    }

    public function isDownloaded(string $url): bool
    {
        $filename = $this->convertToFilename($url);
        return Cover::where('filename', $filename)->exists();
    }

    public function isCoverable(string $url, Model $model): bool
    {
        $filename = $this->convertToFilename($url);
        return Coverables::where('cover_id', $filename)->where('coverable_id', $model->id)->where('coverable_type', get_class($model))->exists();
    }
}
