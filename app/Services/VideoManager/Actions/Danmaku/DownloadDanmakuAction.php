<?php
namespace App\Services\VideoManager\Actions\Danmaku;

use App\Models\Danmaku;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\BilibiliService;
use Carbon\Carbon;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class DownloadDanmakuAction
{

    public function __construct(
        public BilibiliService $bilibiliService
    ) {
    }

    protected function saveDanmaku(string $cid, array $danmaku): void
    {
        $danmakuIds = Danmaku::where('cid', $cid)->select([
            'id',
        ])->get();
        $danmakuIds = $danmakuIds->pluck('id')->toArray();

        // 过滤出新增的数据
        $insertData = array_filter($danmaku, function ($item) use ($danmakuIds) {
            // 内容为空的也不要
            if (empty($item['content'])) {
                return false;
            }
            return ! in_array($item['id'], $danmakuIds);
        });

        // 释放内存
        unset($danmakuIds);

        $start_time = microtime(true);
        $videoPart  = VideoPart::where('cid', $cid)->first();
        $insertData = array_map(function ($item) use ($videoPart) {
            $item['video_id'] = $videoPart->video_id;
            return $item;
        }, $insertData);
        Log::info('Save danmaku count, array_map time', ['time' => microtime(true) - $start_time]);

        $start_time = microtime(true);
        // 分批插入数据，每批1000条
        foreach (array_chunk($insertData, 1000) as $chunk) {
            $videoPart->danmakus()->createMany($chunk);
            unset($chunk);
        }

        // 释放原始数据
        unset($insertData);

        $videoPart->danmakus()->update([
            'video_id' => $videoPart->video_id,
        ]);

        unset($videoPart);
        Log::info('Save danmaku count, time', ['time' => microtime(true) - $start_time]);
    }

    public function execute(VideoPart $videoPart): void
    {
        //加锁
        $lock = redis()->setnx(sprintf('danmaku_downloading:%s', $videoPart->cid), 1);
        if (! $lock) {
            Log::info('Danmaku is being downloaded', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
            return;
        }
        redis()->expire(sprintf('danmaku_downloading:%s', $videoPart->cid), 3600 * 8);

        try {
            Log::info('Download danmaku start', ['id' => $videoPart->cid, 'title' => $videoPart->part]);
            $video = Video::where('id', $videoPart->video_id)->first();

            $start_time   = microtime(true);
            $partDanmakus = $this->bilibiliService->getDanmaku($videoPart->cid, intval($videoPart->duration));
            Log::info('Download danmaku time', ['time' => microtime(true) - $start_time]);

            Log::info('Download danmaku success', [
                'id'       => $videoPart->cid,
                'title'    => $videoPart->part,
                'count'    => count($partDanmakus),
                'video_id' => $video->id,
                'bvid'     => $video->bvid,
                'title'    => $video->title,
            ]);
            $this->saveDanmaku($videoPart->cid, $partDanmakus);
            Log::info('Save danmaku time', ['time' => microtime(true) - $start_time]);
            $videoPart->danmaku_downloaded_at = Carbon::now();
            $videoPart->save();

        } catch (\Throwable $e) {
            Log::error('Download danmaku failed', ['id' => $videoPart->cid, 'title' => $videoPart->part, 'error' => $e->getMessage()]);
            return;
        } finally {
            // 移除下载锁
            redis()->del(sprintf('danmaku_downloading:%s', $videoPart->cid));
            Log::info('Download danmaku end, take time', ['time' => microtime(true) - $start_time, 'id' => $videoPart->cid, 'title' => $videoPart->part]);
        }
    }
}
