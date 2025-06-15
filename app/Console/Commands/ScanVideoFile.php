<?php
namespace App\Console\Commands;

use App\Jobs\ScanCheckVideoLocalFileJob;
use App\Models\VideoPart;
use Illuminate\Console\Command;

class ScanVideoFile extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:scan-video-file {--video-id=}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'scan video file';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $videoId = $this->option('video-id');
        if($videoId){
            VideoPart::where('video_id', '=', $videoId)->where('video_download_path', '=', null)->chunk(100, function ($videoParts) {
                foreach ($videoParts as $videoPart) {
                    dispatch(new ScanCheckVideoLocalFileJob($videoPart->id));
                }
            });
        }else{
            VideoPart::where('video_download_path', '=', null)->chunk(100, function ($videoParts) {
                foreach ($videoParts as $videoPart) {
                    dispatch(new ScanCheckVideoLocalFileJob($videoPart->id));
                }
            });
        }
    }
}
