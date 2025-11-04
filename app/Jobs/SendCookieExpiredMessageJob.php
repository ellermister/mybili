<?php

namespace App\Jobs;

use App\Contracts\TelegramBotServiceInterface;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

class SendCookieExpiredMessageJob implements ShouldQueue
{

    use Queueable;

    /**
     * Create a new job instance.
     */
    public function __construct()
    {
        $this->queue = 'fast';
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        $telegramBotService = app(TelegramBotServiceInterface::class);

        $htmlMessage = $this->buildNotificationMessage();
        $telegramBotService->sendHtmlMessage($htmlMessage);
    }

    /**
     * Build the notification message with improved styling
     */
    private function buildNotificationMessage(): string
    {
        return "
🤖 <b>Mybili Notification</b>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ <b>Cookie Expired</b>

🍪 <b>Status</b>: Cookie has expired
⏰ <b>Time</b>: " . now()->format('Y-m-d H:i:s') . "

🔧 <i>Please update your cookie as soon as possible to continue using the service.</i>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
        ";
    }
}
