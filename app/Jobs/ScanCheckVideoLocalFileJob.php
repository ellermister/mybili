<?php
namespace App\Jobs;

use App\Contracts\VideoDownloadServiceInterface;
use App\Models\VideoPart;
use Illuminate\Contracts\Queue\ShouldQueue;

class ScanCheckVideoLocalFileJob implements ShouldQueue
{
    public $queue = 'slow';
    /**
     * Create a new job instance.
     */
    public function __construct(public int $videoPartId, public bool $download = false)
    {
        //
    }

    /**
     * Execute the job.
     */
    public function handle(VideoDownloadServiceInterface $videoDownloadService): void
    {
        $videoPart = VideoPart::where('id', $this->videoPartId)->first();
        if ($videoPart) {
            $videoDownloadService->downloadVideoPartFile($videoPart, $this->download);
        }
    }
}
