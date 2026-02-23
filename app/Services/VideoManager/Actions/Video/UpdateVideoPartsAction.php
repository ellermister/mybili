<?php
namespace App\Services\VideoManager\Actions\Video;

use App\Events\VideoPartUpdated;
use App\Models\Video;
use App\Models\VideoPart;
use App\Services\BilibiliService;
use App\Services\VideoManager\Actions\Audio\UpdateAudioPartAction;
use App\Services\VideoManager\Traits\VideoDataTrait;
use Illuminate\Support\Facades\Log;

class UpdateVideoPartsAction
{

    public function __construct(
        public BilibiliService $bilibiliService
    ) {
    }

    use VideoDataTrait;

    /**
 * 更新视频分P信息
     */
    public function execute(Video $video): void
    {
        Log::info('Update video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
        if ($video->isAudio()) {
            app(UpdateAudioPartAction::class)->execute($video);
            return;
        }

        if ($this->videoIsInvalid($video->toArray())) {
            Log::info('Video is invalid, skip update video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        try {
            if (config('services.bilibili.id_type') == 'bv') {
                $videoParts = $this->bilibiliService->getVideoParts($video->bvid);
            } else {
                $videoParts = $this->bilibiliService->getVideoParts($video->id);
            }
        } catch (\Exception $e) {
            Log::error('Get video parts failed', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
            return;
        }

        // 找出本地多余远端的数据
        $localVideoParts     = VideoPart::where('video_id', $video->id)->orderBy('page', 'asc')->get()->toArray();
        $localVideoPartsIds  = array_column($localVideoParts, 'cid');
        $remoteVideoPartsIds = array_column($videoParts, 'cid');

        $deleteVideoPartsIds = array_diff($localVideoPartsIds, $remoteVideoPartsIds);

        if (! empty($deleteVideoPartsIds)) {
            Log::info('Delete video parts', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title, 'deleteVideoPartsIds' => $deleteVideoPartsIds]);
            return;
        }

        // 第一步：收集和分类数据（新增 vs 修改）
        $toInsert = []; // 要新增的数据
        $toUpdate = []; // 要修改的数据
        $oldVideoPartsData = [];
        
        // 查询所有已存在的 video_part，避免在循环中多次查询
        $existingParts = VideoPart::where('video_id', $video->id)
            ->pluck('cid')
            ->flip()
            ->toArray();
        
        $now = now();
        foreach ($videoParts as $part) {
            $data = [
                'cid'         => $part['cid'],
                'page'        => $part['page'],
                'from'        => $part['from'],
                'part'        => $part['part'],
                'duration'    => $part['duration'],
                'vid'         => $part['vid'],
                'weblink'     => $part['weblink'],
                'width'       => $part['dimension']['width'] ?? 0,
                'height'      => $part['dimension']['height'] ?? 0,
                'rotate'      => $part['dimension']['rotate'] ?? 0,
                'first_frame' => $part['first_frame'] ?? '',
                'video_id'    => $video->id,
                'updated_at'  => $now,
            ];
            
            if (isset($existingParts[$part['cid']])) {
                // 已存在，需要更新
                $videoPart = VideoPart::where('cid', $part['cid'])->first();
                $oldVideoPartsData[$part['cid']] = $videoPart->toArray();
                $toUpdate[] = $data;
            } else {
                // 不存在，需要新增（添加 created_at）
                $oldVideoPartsData[$part['cid']] = null;
                $data['created_at'] = $now;
                $toInsert[] = $data;
            }
        }

        // 第二步：批量插入新增的数据
        if (!empty($toInsert)) {
            try {
                // 尝试批量插入
                $insertedCount = 0;
                foreach (array_chunk($toInsert, 100) as $chunk) {
                    $this->retryOperation(function () use ($chunk, &$insertedCount) {
                        VideoPart::insert($chunk);
                        $insertedCount += count($chunk);
                    });
                }
                Log::info("Batch inserted {$insertedCount} video parts", ['video_id' => $video->id]);
            } catch (\Exception $e) {
                // 批量插入失败，可能是重复，逐个插入
                Log::warning('Batch insert failed, trying one by one', [
                    'video_id' => $video->id,
                    'error' => $e->getMessage()
                ]);
                
                foreach ($toInsert as $data) {
                    try {
                        $this->retryOperation(function () use ($data) {
                            VideoPart::create($data);
                        });
                    } catch (\Exception $e) {
                        Log::error('Failed to insert video part', [
                            'cid' => $data['cid'],
                            'error' => $e->getMessage()
                        ]);
                    }
                }
            }
            
            // 触发新增事件
            foreach ($toInsert as $data) {
                try {
                    $newPart = VideoPart::where('cid', $data['cid'])->first();
                    if ($newPart) {
                        event(new VideoPartUpdated([], $newPart->toArray()));
                    }
                } catch (\Exception $e) {
                    Log::debug('Event trigger failed', ['cid' => $data['cid']]);
                }
            }
        }

        // 第三步：逐个更新修改的数据
        if (!empty($toUpdate)) {
            $updatedCount = 0;
            foreach ($toUpdate as $data) {
                try {
                    $cid = $data['cid'];
                    $this->retryOperation(function () use ($cid, $data) {
                        VideoPart::where('cid', $cid)->update($data);
                    });
                    
                    // 触发更新事件
                    try {
                        $newPart = VideoPart::where('cid', $cid)->first();
                        if ($newPart) {
                            event(new VideoPartUpdated($oldVideoPartsData[$cid] ?? [], $newPart->toArray()));
                        }
                    } catch (\Exception $e) {
                        Log::debug('Event trigger failed', ['cid' => $cid]);
                    }
                    
                    $updatedCount++;
                } catch (\Exception $e) {
                    Log::error('Failed to update video part', [
                        'cid' => $data['cid'],
                        'error' => $e->getMessage()
                    ]);
                }
            }
            Log::info("Updated {$updatedCount} video parts", ['video_id' => $video->id]);
        }

        Log::info('Update video parts success', ['id' => $video->id, 'bvid' => $video->bvid, 'title' => $video->title]);
    }

    /**
     * 重试操作辅助方法
     * 
     * @param callable $operation 要执行的操作
     * @param int $maxRetries 最大重试次数
     * @param int $retryDelay 重试延迟（毫秒）
     * @throws \Exception
     */
    private function retryOperation(callable $operation, int $maxRetries = 3, int $retryDelay = 100): void
    {
        for ($attempt = 1; $attempt <= $maxRetries; $attempt++) {
            try {
                $operation();
                return; // 成功则返回
            } catch (\Illuminate\Database\QueryException $e) {
                // 检查是否是数据库锁定错误
                if ($e->getCode() === 'HY000' && str_contains($e->getMessage(), 'database is locked')) {
                    if ($attempt < $maxRetries) {
                        Log::debug("Database locked, retrying attempt {$attempt}/{$maxRetries}");
                        usleep($retryDelay * 1000 * $attempt); // 递增延迟
                        continue;
                    }
                }
                // 其他错误或达到最大重试次数，抛出异常
                throw $e;
            }
        }
    }

}
