<?php

namespace App\Jobs;

use App\Services\DownloadFilterService;
use App\Services\VideoManagerService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Log;

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
     * 任务失败前等待的时间（以秒为单位）
     *
     * @var array
     */
    public $backoff =  [1800, 3600, 7200]; 

    
    /**
     * Execute the job.
     */
    public function handle(): void
    {
        if (video_has_invalid($this->vInfo)) {
            Log::info('exclude video by invalid', ['video_id' => $this->vInfo['id'], 'title' => $this->vInfo['title']]);
            return;
        }

        $videoFilterService = app(DownloadFilterService::class);
        if ($videoFilterService->shouldExcludeByName($this->vInfo['title'])) {
            Log::info('exclude video by name', ['video_id' => $this->vInfo['id'], 'title' => $this->vInfo['title']]);
            return;
        }

        // 获取视频信息
        $url = sprintf('https://www.bilibili.com/video/%s/', $this->vInfo['bvid']);
        $command = sprintf('yt-dlp_linux -j %s', escapeshellarg($url));
        exec($command, $output, $result);

        if ($result !== 0) {
            Log::error('获取视频信息失败', ['video_id' => $this->vInfo['id'], 'output' => $output]);
            throw new \Exception("获取视频信息失败");
        }

        $videoManagerService = app(VideoManagerService::class);

        $totalParts = count($output);
        // $output 多个分集的视频就会有多个数组，默认单集只有一个

        $skipParts = [];
        foreach ($output as $key => $item) {
            $partNum = $key + 1;
            // 如果视频已经存在，则再次标记，避免已经下载视频但没有标记
            if ($videoManagerService->hasVideoFile($this->vInfo['id'], $partNum)
                && $videoManagerService->getVideoFileHash($this->vInfo['id'], $partNum) !== null) {
                $videoManagerService->markVideoDownloaded($this->vInfo['id'], $partNum);
                $skipParts[] = $partNum;
                Log::info('video file already exists', ['video_id' => $this->vInfo['id'], 'part' => $partNum]);
                continue;
            }
            $videoInfo    = json_decode($item, true);
            $theVideoSize = $videoInfo['filesize'] ?? $videoInfo['filesize_approx'] ?? 0;

            if ($videoFilterService->shouldExcludeBySize($theVideoSize)) {
                Log::info('exclude video by size', [
                    'video_id' => $this->vInfo['id'],
                    'title'    => $this->vInfo['title'],
                    'size'     => $theVideoSize,
                ]);
                return;
            }
        }

        $videoManagerService->createVideoDirectory();

        $binPath  = base_path('download-video.sh');
        for ($i = 1; $i <= $totalParts; $i++) {

            if(in_array($i, $skipParts)){
                Log::info('skip video', ['video_id' => $this->vInfo['id'], 'part' => $i]);
                continue;
            }

            // 判断是否开启多P缓存
            if(!$videoFilterService->isMultiPEnabled() && $i > 1){
                Log::info('multiP is not enabled', ['video_id' => $this->vInfo['id'], 'part' => $i]);
                continue;
            }

            $savePath = $videoManagerService->getVideoDownloadPath($this->vInfo['id'], $i);
            Log::info('download video', ['video_id' => $this->vInfo['id'], 'part' => $i, 'url' => escapeshellarg($url), 'savePath' => escapeshellarg($savePath), 'binPath' => escapeshellarg($binPath)]);

            $command = sprintf('%s %s %s %s', $binPath, escapeshellarg($url), escapeshellarg($savePath), escapeshellarg($i));
            exec($command, $output, $result);
            if ($result != 0) {
                $msg = implode('', $output);
                throw new \Exception("下载异常:\n" . $msg);
            }
            Log::info('download video output', ['video_id' => $this->vInfo['id'], 'part' => $i, 'output' => $output]);
            Log::info('download video success', ['video_id' => $this->vInfo['id'], 'part' => $i]);
            $videoManagerService->markVideoDownloaded($this->vInfo['id'], $i);
        }

    }

    /**
     * 任务成功完成时的处理
     */
    public function then(): void
    {
        $videoManagerService = app(VideoManagerService::class);
        $videoManagerService->finishDownloadVideo($this->vInfo['id']);
        Log::info('Video download task completed', ['video_id' => $this->vInfo['id']]);
    }

    /**
     * 任务最终失败时的处理
     */
    public function failed(\Throwable $exception): void
    {
        $videoManagerService = app(VideoManagerService::class);
        $videoManagerService->finishDownloadVideo($this->vInfo['id']);
        Log::error('Video download task failed', [
            'video_id' => $this->vInfo['id'],
            'error' => $exception->getMessage()
        ]);
    }
}
