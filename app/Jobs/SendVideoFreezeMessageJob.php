<?php

namespace App\Jobs;

use App\Contracts\TelegramBotServiceInterface;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\VideoDownloadService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

class SendVideoFreezeMessageJob implements ShouldQueue
{
    
    use Queueable;
    
    public $queue = 'fast';
    
    public array $video;

    /**
     * Create a new job instance.
     */
    public function __construct(array $video)
    {
        $this->video = $video;
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        $telegramBotService = app(TelegramBotServiceInterface::class);

        $newVideo = Video::where('id', $this->video['id'])->first();

        $videoParts = VideoPart::where('video_id', $this->video['id'])->get();

        $videoFileSize = 0;
        $videoDuration = 0;
        foreach($videoParts as $videoPart){
            $filePath = app(VideoDownloadService::class)->getVideoPartValidFilePath($videoPart);
            if ($filePath) {
                $videoFileSize += filesize($filePath);
            }

            $videoFileSize += $videoPart->danmakus()->sum('size');

            $videoDuration += $videoPart->duration;
        }

        $htmlMessage = $this->buildNotificationMessage($newVideo, format_file_size($videoFileSize), format_duration($videoDuration));
        $telegramBotService->sendHtmlMessage($htmlMessage);
    }



        /**
     * Build the notification message with improved styling
     */
    private function buildNotificationMessage($video, string $fileSize, string $duration): string
    {
        return "
🤖 <b>Mybili Notification</b>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎬 <b>Video Frozen</b>

📺 <b>Title</b>: {$video->title}
⏱️ <b>Duration</b>: {$duration}
💾 <b>File Size</b>: {$fileSize}
🔗 <b>Source</b>: <a href='https://www.bilibili.com/video/{$video->bvid}'>View on Bilibili</a>

🧊 <i>Video frozen successfully!</i>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
        ";
    }
}
