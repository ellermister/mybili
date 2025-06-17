<?php
namespace App\Jobs;

use App\Contracts\DownloadImageServiceInterface;
use App\Models\Video;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Log;

class DownloadVideoImageJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, SerializesModels;

    public $queue = 'slow';

    public function __construct(
        private array $video,
        private DownloadImageServiceInterface $downloadImageService
    ) {
    }

    public function handle(): void
    {
        $cover = $this->video['cover'] ?? '';

        if (empty($cover)) {
            Log::info('Cover is empty, skip download');
            return;
        }

        $localPath = $this->downloadImageService->getImageLocalPath($cover);
        if (is_file($localPath)) {
            $relativePath = get_relative_path($localPath);
            Video::where('id', $this->video['id'])->update([
                'cache_image' => $relativePath,
            ]);
            return;
        }

        try {
            $this->downloadImageService->downloadImage(
                $cover,
                $localPath
            );

            $relativePath = get_relative_path($localPath);
            Video::where('id', $this->video['id'])->update([
                'cache_image' => $relativePath,
            ]);
        } catch (\Exception $e) {
            Log::error('Download video image failed', ['error' => $e->getMessage()]);
        }
    }
}
