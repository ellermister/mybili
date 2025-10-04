<?php
namespace App\Console\Commands;

use App\Services\BilibiliService;
use Illuminate\Console\Command;

class DanmakuDownload extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:danmaku-download {id} {filename}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'download danmaku to file';

    /**
     * Execute the console command.
     */
    public function handle(BilibiliService $bilibiliService)
    {

        $videoParts = $bilibiliService->getVideoParts($this->argument('id'));
        foreach ($videoParts as $videoPart) {
            $danmaku = $bilibiliService->getDanmaku($videoPart['cid'], $videoPart['duration']);
            usort($danmaku, function ($a, $b) {
                if (isset($a['progress']) && isset($b['progress'])) {
                    return $a['progress'] - $b['progress'];
                }
                return 0;
            });
            $jsonStr  = json_encode($danmaku, JSON_UNESCAPED_UNICODE);
            $savePath = sprintf('%s-%s-%s.json', base_path($this->argument('filename')), $videoPart['part'], $videoPart['cid']);
            file_put_contents($savePath, $jsonStr);
            $this->info('Danmaku downloaded to file: ' . $savePath);
        }
    }
}
