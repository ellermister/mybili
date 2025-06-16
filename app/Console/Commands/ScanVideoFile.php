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
    protected $signature = 'app:scan-video-file {--video-id=} {--force} {--download}';

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
        $download = $this->option('download') ? true : false;
        $videoId  = $this->option('video-id');
        if ($videoId) {
            $videoParts = VideoPart::where('video_id', '=', $videoId)->get();
            if ($videoParts->count() > 0) {
                foreach ($videoParts as $part) {
                    dispatch(new ScanCheckVideoLocalFileJob($part->id, $download));
                }
            } else {
                $this->error('Video parts not found');
            }
        } else {
            $query = VideoPart::query();
            $force = $this->option('force');    
            if (! $force) {
                $query->whereNull('video_download_path');
            }
            $count = $query->count();
            $this->info('Total video parts to scan: ' . $count);
            $query->chunk(100, function ($videoParts) use ($download) {
                foreach ($videoParts as $videoPart) {
                    dispatch(new ScanCheckVideoLocalFileJob($videoPart->id, $download));
                }
            });
        }
    }
}
