<?php
namespace App\Listeners;

use App\Contracts\TelegramBotServiceInterface;
use App\Events\VideoPartDownloaded;
use App\Services\VideoDownloadService;
use Illuminate\Contracts\Queue\ShouldQueue;

class SendVideoDownloadedMessage implements ShouldQueue
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

        $videoFileSize = 0;
        $filePath = app(VideoDownloadService::class)->getVideoPartValidFilePath($videoPart);
        if ($filePath) {
            $videoFileSize = filesize($filePath);
        }

        $readableFileSize = $this->formatFileSize($videoFileSize);
        $formattedDuration = $this->formatDuration($videoPart->duration);

        $telegramBotService = app(TelegramBotServiceInterface::class);
        $htmlMessage = $this->buildNotificationMessage($video, $videoPart, $readableFileSize, $formattedDuration);

        $telegramBotService->sendHtmlMessage($htmlMessage);
    }

    /**
     * Convert bytes to human readable file size
     * Maximum unit is GB, minimum unit is MB
     */
    private function formatFileSize(int $bytes): string
    {
        if ($bytes === 0) {
            return '0 MB';
        }

        $gb = $bytes / (1024 * 1024 * 1024);
        if ($gb >= 1) {
            return round($gb, 2) . ' GB';
        }

        $mb = $bytes / (1024 * 1024);
        return round($mb, 2) . ' MB';
    }

    /**
     * Format duration from seconds to readable format (HH:MM:SS or MM:SS)
     */
    private function formatDuration(string $duration): string
    {
        // Convert duration to integer (assuming it's in seconds)
        return gmdate('H:i:s', intval($duration));
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
