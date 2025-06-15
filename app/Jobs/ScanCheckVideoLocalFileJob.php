<?php
namespace App\Jobs;

use App\Contracts\VideoManagerServiceInterface;
use App\Models\VideoPart;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

class ScanCheckVideoLocalFileJob implements ShouldQueue
{
    public $queue = 'slow';
    /**
     * Create a new job instance.
     */
    public function __construct(public int $videoPartId)
    {
        //
    }

    /**
     * Execute the job.
     */
    public function handle(VideoManagerServiceInterface $videoManagerService): void
    {
        $videoPart = VideoPart::where('id', $this->videoPartId)->first();
        if ($videoPart) {
            $videoManagerService->downloadVideoPartFile($videoPart, true);
        }
    }
}
