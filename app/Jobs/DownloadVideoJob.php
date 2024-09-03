<?php

namespace App\Jobs;

use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

class DownloadVideoJob implements ShouldQueue
{
    use Queueable;

    public $vInfo;
    /**
     * Create a new job instance.
     */
    public function __construct(array $vInfo)
    {
        $this->vInfo = $vInfo;
    }

    public $tries = 3;

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        if (video_has_invalid($this->vInfo)) {
            $this->updateVideoStatus();
            return;
        }

        $url = sprintf('https://www.bilibili.com/video/%s/', $this->vInfo['bvid']);

        $videoPath = storage_path('app/public/videos');
        if (!is_dir($videoPath)) {
            mkdir($videoPath, 0644);
            if (!is_dir($videoPath)) {
                throw new \Exception("下载路径不存在, 创建失败");
            }
        }
        $savePath = sprintf('%s/%s.mp4', $videoPath, $this->vInfo['id']);
        $hashPath = $savePath . '.hash';

        if (is_file($hashPath) && is_file($savePath)) {
            if (file_get_contents($hashPath) == hash_file('sha256', $savePath)) {
                // 文件存在无需下载
                return;
            }
        }

        $binPath = base_path('download-video.sh');
        $command = sprintf('%s %s %s', $binPath, $url, $savePath);
        exec($command, $output, $result);

        if ($result != 0) {
            $msg = implode('', $output);
            throw new \Exception("下载异常:\n" . $msg);
        }

        $calcHash = hash_file('sha256', $savePath);
        file_put_contents($hashPath, $calcHash);
        $this->updateVideoStatus();

        $key = sprintf('download_lock:%s', $this->vInfo['id']);
        redis()->del($key);
    }

    public function updateVideoStatus()
    {
        $videoPath = storage_path('app/public/videos');
        $savePath  = sprintf('%s/%s.mp4', $videoPath, $this->vInfo['id']);
        if (is_file($savePath)) {
            redis()->hSet('video_downloaded', $this->vInfo['id'], 1);
        }
    }
}
