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

        // 使用cURL替代file_get_contents来解决SSL问题
        $content = $this->fetchContentWithCurl($url);
        if (! $content) {
            throw new \Exception("Failed to fetch image from: " . $url);
        }
        
        file_put_contents($savePath, $content);
        file_put_contents($hashPath, hash_file('sha256', $savePath));
    }

    /**
     * 使用cURL获取内容，解决SSL证书验证问题
     */
    protected function fetchContentWithCurl(string $url): string|false
    {
        if (!function_exists('curl_init')) {
            // 如果cURL不可用，回退到file_get_contents但禁用SSL验证
            return $this->fetchContentWithFileGetContents($url);
        }

        $ch = curl_init();
        curl_setopt_array($ch, [
            CURLOPT_URL => $url,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_MAXREDIRS => 5,
            CURLOPT_TIMEOUT => 30,
            CURLOPT_CONNECTTIMEOUT => 10,
            CURLOPT_SSL_VERIFYPEER => false,  // 禁用SSL证书验证
            CURLOPT_SSL_VERIFYHOST => false,  // 禁用主机名验证
            CURLOPT_USERAGENT => 'Mozilla/5.0 (compatible; ImageDownloader/1.0)',
            CURLOPT_HTTPHEADER => [
                'Accept: image/*,*/*;q=0.8',
                'Accept-Language: zh-CN,zh;q=0.9,en;q=0.8',
            ],
        ]);

        $content = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);

        if ($content === false) {
            throw new \Exception("cURL error: " . $error);
        }

        if ($httpCode !== 200) {
            throw new \Exception("HTTP error: " . $httpCode . " for URL: " . $url);
        }

        return $content;
    }

    /**
     * 回退方案：使用file_get_contents但禁用SSL验证
     */
    protected function fetchContentWithFileGetContents(string $url): string|false
    {
        $context = stream_context_create([
            'ssl' => [
                'verify_peer' => false,
                'verify_peer_name' => false,
                'allow_self_signed' => true,
            ],
            'http' => [
                'timeout' => 30,
                'user_agent' => 'Mozilla/5.0 (compatible; ImageDownloader/1.0)',
            ],
        ]);

        return file_get_contents($url, false, $context);
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
