<?php
namespace App\Console\Commands;

use App\Contracts\DownloadImageServiceInterface;
use App\Jobs\DownloadVideoImageJob;
use App\Models\Video;
use Illuminate\Console\Command;

class ScanVideoImage extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:scan-video-image {--id=} {--force}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'scan video image';

    /**
     * Execute the console command.
     */
    public function handle(DownloadImageServiceInterface $downloadImageService)
    {
        $id = $this->option('id');
        if ($id) {
            $video = Video::find($id);
            if ($video) {
                dispatch(new DownloadVideoImageJob($video->toArray(), $downloadImageService));
            } else {
                $this->error('Video not found');
            }
        } else {
            $force = $this->option('force');
            $query = Video::query();
            if (!$force) {
                $query->where('cache_image', '=', '');
            }
            $count = $query->count();
            $this->info('Total videos to scan: ' . $count);
            $query->chunk(100, function ($videos) use ($downloadImageService) {
                foreach ($videos as $video) {
                    dispatch(new DownloadVideoImageJob($video->toArray(), $downloadImageService));
                }
            });
        }
    }
}
