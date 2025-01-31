<?php

namespace App\Jobs;

use App\Services\DownloadFilterService;
use App\Services\VideoManagerService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Log;

class DownloadAllVideoJob implements ShouldQueue
{
    use Queueable;


    public VideoManagerService $videoManagerService;
    public DownloadFilterService $downloadFilterService;
    /**
     * Create a new job instance.
     */
    public function __construct()
    {
        $this->videoManagerService = app(VideoManagerService::class);
        $this->downloadFilterService = app(DownloadFilterService::class);
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
       $favList = $this->videoManagerService->getFavList();
       foreach($favList as $item){
            if($this->downloadFilterService->shouldExcludeByFav($item['id'])){
                Log::info('Exclude fav', ['id' => $item['id'], 'title' => $item['title']]);
                continue;
            }

            $videoList = $this->videoManagerService->getVideoListByFav($item['id']);
            foreach($videoList as $video){

                if($this->videoManagerService->videoIsInvalid($video) || $video['invalid']){
                    Log::info('Video is invalid, skip', ['id' => $video['id'], 'title' => $video['title']]);
                    continue;
                }

                if($this->videoManagerService->videoDownloaded($video['id'])){
                    Log::info('Video already downloaded, try to check parts', ['id' => $video['id'], 'title' => $video['title']]);

                    // 已经缓存的情况下进一步检查分P是否缓存
                    $videoTotalParts = intval($video['page'] ?? 1);
                    $existsParts = $this->videoManagerService->getAllPartsVideo($video['id'], $videoTotalParts);

                    if(count($existsParts) >= $videoTotalParts){
                        Log::info('Video all parts already downloaded,skip', ['id' => $video['id'], 'title' => $video['title'], 'parts' => $videoTotalParts]);
                        continue;
                    }
                }

                //检查名称是否符合
                if($this->downloadFilterService->shouldExcludeByName($video['title'])){
                    Log::info('Video name not match, skip', ['id' => $video['id'], 'title' => $video['title']]);
                    continue;
                }

                $this->videoManagerService->dispatchDownloadVideoJob($video);
            }
        }
    }
}
