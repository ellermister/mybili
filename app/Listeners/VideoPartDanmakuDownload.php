<?php
namespace App\Listeners;

use App\Enums\SettingKey;
use App\Events\VideoPartUpdated;
use App\Jobs\DownloadDanmakuJob;
use App\Models\VideoPart;
use App\Services\SettingsService;
use Carbon\Carbon;
use Illuminate\Contracts\Queue\ShouldQueue;
use Log;

class VideoPartDanmakuDownload implements ShouldQueue
{
    /**
     * 指定队列名称
     */
    public $queue = 'default';

    /**
     * Create the event listener.
     */
    public function __construct(
        public SettingsService $settings
    ) {
    }

    /**
     * Handle the event.
     */
    public function handle(VideoPartUpdated $event): void
    {
        if (isset($event->newVideoPart)) {
            $videoPart = VideoPart::where('cid', $event->newVideoPart['cid'])->first();
            if ($videoPart) {
                if ($this->settings->get(SettingKey::DANMAKU_DOWNLOAD_ENABLED) != 'on') {
                    Log::info('Download danmaku disabled', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                    return;
                }

                // 如果上次下载时间小于7天则不更新
                if ($videoPart->danmaku_downloaded_at && $videoPart->danmaku_downloaded_at > Carbon::now()->subDays(7)) {
                    Log::info('Danmaku has been saved in the last 7 days', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                    return;
                }

                dispatch(new DownloadDanmakuJob($videoPart));
                Log::info('Download danmaku job dispatched', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
            }
        }
    }
}
