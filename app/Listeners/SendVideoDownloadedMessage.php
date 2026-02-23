<?php
namespace App\Listeners;

use App\Contracts\TelegramBotServiceInterface;
use App\Events\VideoPartDownloaded;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Illuminate\Contracts\Queue\ShouldQueue;
use Laravel\Horizon\Contracts\Silenced;

class SendVideoDownloadedMessage implements ShouldQueue, Silenced
{
    public $queue = 'fast';

    /**
     * Create the event listener.
     */
    public function __construct()
    {

    }

    /**
     * Handle the event.
     */
    public function handle(VideoPartDownloaded $event): void
    {
        $videoPart = $event->videoPart;
        $video     = $videoPart->video;

        $videoFileSize = app(VideoServiceInterface::class)->getVideoPartFileSize($videoPart);

        $readableFileSize = format_file_size($videoFileSize);
        $formattedDuration = format_duration($videoPart->duration);

        $telegramBotService = app(TelegramBotServiceInterface::class);
        $htmlMessage = $this->buildNotificationMessage($video, $videoPart, $readableFileSize, $formattedDuration);

        $telegramBotService->sendHtmlMessage($htmlMessage);
    }

    /**
     * Build the notification message with improved styling
     */
    private function buildNotificationMessage($video, $videoPart, string $fileSize, string $duration): string
    {
        $partInfo = $videoPart->page > 1 ? "Part {$videoPart->page}" : "Single Part";
        
        return "
🤖 <b>Mybili Notification</b>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎬 <b>Video Download Completed</b>

📺 <b>Title</b>: {$video->title}
🎥 <b>Part</b>: {$videoPart->part} ({$partInfo})
⏱️ <b>Duration</b>: {$duration}
💾 <b>File Size</b>: {$fileSize}
🔗 <b>Source</b>: <a href='https://www.bilibili.com/video/{$video->bvid}'>View on Bilibili</a>

✅ <i>Download completed successfully!</i>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
        ";
    }
}
