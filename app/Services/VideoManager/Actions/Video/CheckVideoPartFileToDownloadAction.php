<?php
namespace App\Services\VideoManager\Actions\Video;

use App\Enums\SettingKey;
use App\Jobs\DownloadVideoJob;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\DownloadFilterService;
use App\Services\DownloadVideoService;
use App\Services\SettingsService;
use Carbon\Carbon;
use Log;
use Storage;
use Str;

class CheckVideoPartFileToDownloadAction
{
    public function __construct(
        public SettingsService $settingsService,
        public DownloadVideoService $downloadVideoService,
        public DownloadFilterService $downloadFilterService
    ) {
    }

    public function execute(VideoPart $videoPart, bool $tryDownload = false): void
    {
        $video = Video::where('id', $videoPart->video_id)->first();

        // 获取当前分片在视频第几个索引
        $videoParts = VideoPart::where('video_id', $video->id)->select(['video_id', 'page', 'cid'])->get();

        // 如果是全数字命名的，则数字对应的就是文件名索引
        $isNumberNamed = false;
        foreach ($videoParts as $item) {
            if (preg_match('/^\d+$/', $item['page']) && intval($item['page']) == $item['page']) {
                $isNumberNamed = true;
            } else {
                $isNumberNamed = false;
                break;
            }
        }

        $currentIndex = 1;
        if ($isNumberNamed) {
            $currentIndex = $videoPart->page;
        } else {
            // 如果不是数字索引的，则按照排序，取当前分片在视频第几个索引
            $videoParts   = $videoParts->sortBy('video_id');
            $currentIndex = $videoParts->search(function ($item) use ($videoPart) {
                return $item['cid'] == $videoPart->cid;
            }) + 1;
        }

        if ($video) {
            $isDownloaded = false;
            if ($videoPart->video_download_path) {
                $savePath = Storage::disk('public')->path(Str::after($videoPart->video_download_path, '/storage/'));
                if (is_file($savePath)) {
                    $isDownloaded = true;
                }
            } else {
                // 检查是否存在遗留旧格式路径， 新路径都存储了video_download_path走上面的逻辑
                $savePath = $this->downloadVideoService->getOldDownloadPath($video->id, $currentIndex);
                if (is_file($savePath)) {
                    $isDownloaded = true;
                }
            }

            if (! $isDownloaded) {
                if (! $tryDownload) {
                    return;
                }

                // 下载视频分P文件
                if ($this->settingsService->get(SettingKey::VIDEO_DOWNLOAD_ENABLED) == 'on') {

                    if ($this->downloadFilterService->shouldExcludeByDuration(intval($video->duration))) {
                        Log::info('Video file not exists, download video file excluded', ['id' => $video->id, 'title' => $video->title]);
                        return;
                    }


                    if ($this->downloadFilterService->shouldExcludeByDurationPart($videoPart->duration)) {
                        Log::info('Video part file not exists, download video part file excluded', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                        return;
                    }

                    // 加锁， 控流
                    $lock = redis()->setnx(sprintf('video_downloading:%s', $videoPart->cid), 1);
                    if (! $lock) {
                        Log::info('Video is being downloaded', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                        return;
                    }
                    redis()->expire(sprintf('video_downloading:%s', $videoPart->cid), 3600 * 8);

                    DownloadVideoJob::dispatchWithRateLimit($videoPart);
                } else {
                    Log::info('Video part file not exists, download video part file disabled', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
                }
            } else {
                // 如果文件存在，则更新记录到数据库并写入日志
                Log::info('Video part file already exists', ['id' => $videoPart->cid, 'title' => $videoPart->part]);

                // 时间取文件创建时间
                $videoPart->video_downloaded_at = Carbon::createFromTimestamp(filectime($savePath));

                // 获取相对路径
                $videoPart->video_download_path = get_relative_path($savePath);
                $videoPart->save();
            }

            // 更新視頻已經下載的文件數量
            $video->video_downloaded_num = VideoPart::where('video_id', $video->id)->whereNotNull('video_download_path')->count();
            $video->save();
        }
    }

}
