<?php
namespace App\Services;

use App\Contracts\DownloadImageServiceInterface;

class DownloadImageService implements DownloadImageServiceInterface
{
    public function downloadImage(string $url, string $savePath): void
    {
        if(empty($url) || empty($savePath)){
            throw new \Exception("url or savePath is empty");
        }

        $hashPath = $savePath . '.hash';
        if (is_file($hashPath)) {
            $hash     = file_get_contents($hashPath);
            $hashText = hash_file('sha256', $savePath);
            if ($hash === $hashText) {
                return;
            }
        }
        $content = file_get_contents($url, false);
        if (! $content) {
            throw new \Exception("Failed to fetch image.");
        }
        file_put_contents($savePath, $content);
        file_put_contents($hashPath, hash_file('sha256', $savePath));
    }

    protected function convertToFilename(string $url): string
    {
        if (empty($url)) {
            return "";
        }

        $urlParts = explode('/', $url);
        $filename = end($urlParts);

        if (! $filename) {
            $filename = base64_encode($url) . '.jpg';
        }
        $filename = preg_replace('/[^a-zA-Z90-9](?!(jpg|png|gif|svg|webp))/', '', $filename);
        return $filename;
    }

    public function getImagesDirIfNotExistCreate(): string
    {
        $dir = storage_path('app/public/images');
        if (! is_dir($dir)) {
            mkdir($dir, 0644);
        }
        return $dir;
    }

    public function getImageLocalPath(string $url): string
    {
        return $this->getImagesDirIfNotExistCreate() . '/' . $this->convertToFilename($url);
    }
}
