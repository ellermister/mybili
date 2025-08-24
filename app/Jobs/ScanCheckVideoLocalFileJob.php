<?php
namespace App\Jobs;

use App\Models\VideoPart;
use App\Services\VideoManager\Actions\Video\CheckVideoPartFileToDownloadAction;
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
    public function handle(CheckVideoPartFileToDownloadAction $checkVideoPartFileToDownloadAction): void
    {
        $videoPart = VideoPart::where('id', $this->videoPartId)->first();
        if ($videoPart) {
            $checkVideoPartFileToDownloadAction->execute($videoPart, $this->download);
        }
    }
}
