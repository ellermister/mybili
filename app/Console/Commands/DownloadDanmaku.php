<?php

namespace App\Console\Commands;

use App\Contracts\VideoManagerServiceInterface;
use App\Services\BilibiliService;
use App\Services\DownloadFilterService;
use Illuminate\Console\Command;
use Log;

class DownloadDanmaku extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:download-danmaku';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = '下载弹幕';

    /**
     * Execute the console command.
     */
    public function handle(VideoManagerServiceInterface $videoManagerService, DownloadFilterService $downloadFilterService)
    {
        $favList = $videoManagerService->getFavList();
        foreach ($favList as $fav) {
            if($downloadFilterService->shouldExcludeByFav($fav['id'])){
                Log::info('Exclude fav, download danmaku skip', ['id' => $fav['id'], 'title' => $fav['title']]);
                continue;
            }


            $videoList = $videoManagerService->getVideoListByFav($fav['id']);
            foreach ($videoList as $video) {
                if($videoManagerService->videoIsInvalid($video) || $video['invalid']){
                    Log::info('Video is invalid, download danmaku skip', ['id' => $video['id'], 'title' => $video['title']]);
                    continue;
                }

                //检查名称是否符合
                if($downloadFilterService->shouldExcludeByName($video['title'])){
                    Log::info('Video name not match, download danmaku skip', ['id' => $video['id'], 'title' => $video['title']]);
                    continue;
                }


                // 检查上次更新时间是否太短
                $lastDownloadTime = $videoManagerService->danmakuDownloadedTime($video['id']);
                if($lastDownloadTime && time() - $lastDownloadTime < 3600 * 24 * 7){
                    Log::info('Danmaku already downloaded, download danmaku skip', ['id' => $video['id'], 'title' => $video['title']]);
                    continue;
                }

                $videoManagerService->dispatchDownloadDanmakuJob($video['id']);
            }
        }
    }
}
