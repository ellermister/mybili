<?php
namespace App\Services;

use App\Events\SubscriptionUpdated;
use App\Events\VideoUpdated;
use App\Jobs\PullVideoInfoJob;
use App\Jobs\UpdateSubscriptionJob;
use App\Models\Subscription;
use App\Models\SubscriptionVideo;
use App\Models\Video;
use App\Services\BilibiliService;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Http;
use Log;

class SubscriptionService
{
    public function __construct(public BilibiliService $bilibiliService, public VideoServiceInterface $videoService)
    {
    }

    public function deleteSubscription(Subscription $subscription)
    {
        // 预加载关联数据，避免 N+1 查询
        $subscription->load(['videos.favorite', 'videos.subscriptions']);
        
        $removeIds = [];
        foreach($subscription->videos as $video) {
            // 判断视频是否被收藏夹引用
            if($video->favorite->count() > 0){
                Log::info('video is referenced by favorite', [
                    'video_id' => $video->id, 
                    'subscription_id' => $subscription->id, 
                    'title' => $video->title
                ]);
                continue;
            }

            // 判断视频是否被其他订阅引用（大于1表示除了当前订阅外还有其他订阅）
            if($video->subscriptions->count() > 1){
                Log::info('video is referenced by other subscription', [
                    'video_id' => $video->id, 
                    'subscription_id' => $subscription->id,
                    'subscriptions_count' => $video->subscriptions->count(),
                    'title' => $video->title
                ]);
                continue;
            }
            
            // 只有当视频没有被任何收藏夹或其他订阅引用时，才加入删除列表
            $removeIds[] = $video->id;
        }
        
        // 删除未被引用的视频
        if (!empty($removeIds)) {
            $this->videoService->deleteVideos($removeIds);
        }
        
        // 删除订阅与视频的关联关系
        SubscriptionVideo::where('subscription_id', $subscription->id)->delete();
        
        // 删除订阅与封面的关联关系（不删除封面本身）
        $subscription->coverImage()->detach();
        
        // 删除订阅记录
        $subscription->delete();
    }


    /**
     * 获取重定向后的URL
     */
    public function getRedirectURL($url)
    {
        $response = Http::get($url);
        return empty($response->effectiveUri()->__toString()) ? $url : $response->effectiveUri()->__toString();
    }

    public function addSubscription($type, $url)
    {
        if(preg_match("#https://b23.tv#", $url)){
            $url = $this->getRedirectURL($url);
        }
        
        if ($type == 'seasons') {
            if (! preg_match('#/(\d+)/lists/(\d+)#', $url, $matches)) {
                throw new \Exception('invalid seasons url');
            }
            $mid                   = $matches[1];
            $seasonId              = $matches[2];
            $subscription          = Subscription::query()->where('mid', $mid)->where('list_id', $seasonId)->firstOrNew();
            $subscription->type    = $type;
            $subscription->mid     = $mid;
            $subscription->list_id = $seasonId;
            $subscription->url     = $url;
            $subscription->save();

            UpdateSubscriptionJob::dispatch($subscription, true);
            return $subscription;
        } else if ($type == 'series') {
            if (! preg_match('#/(\d+)/lists/(\d+)#', $url, $matches)) {
                throw new \Exception('invalid series url');
            }
            $mid                   = $matches[1];
            $seriesId              = $matches[2];
            $subscription          = Subscription::query()->where('mid', $mid)->where('list_id', $seriesId)->firstOrNew();
            $subscription->type    = $type;
            $subscription->mid     = $mid;
            $subscription->list_id = $seriesId;
            $subscription->url     = $url;
            $subscription->save();
            UpdateSubscriptionJob::dispatch($subscription, true);
            return $subscription;
        } else {
            if (! preg_match('#/(\d+)/upload#', $url, $matches) && !preg_match('#https://space.bilibili.com/(\d+)#', $url, $matches)) {
                throw new \Exception('invalid up url');
            }
            $mid                = $matches[1];
            $subscription       = Subscription::query()->where('mid', $mid)->where('type', 'up')->firstOrNew();
            $subscription->type = 'up';
            $subscription->mid  = $mid;
            $subscription->url  = $url;
            $subscription->save();
            UpdateSubscriptionJob::dispatch($subscription, true);
            return $subscription;
        }
    }

    public function getSubscriptions()
    {
        return Subscription::query()->get();
    }

    public function disableSubscription(Subscription $subscription)
    {
        $subscription->status = Subscription::STATUS_DISABLED;
        $subscription->save();
    }

    public function enableSubscription(Subscription $subscription)
    {
        $subscription->status = Subscription::STATUS_ACTIVE;
        $subscription->save();
    }

    public function changeSubscription(Subscription $subscription, array $data)
    {
        $subscription->fill($data);
        $subscription->save();
        return $subscription;
    }

    public function updateSubscriptions($pullAll = false)
    {
        $subscriptions = Subscription::where('status', Subscription::STATUS_ACTIVE)->get();
        foreach ($subscriptions as $subscription) {
            $this->updateSubscription($subscription, $pullAll);
        }
    }

    public function updateSubscription(Subscription $subscription, bool $pullAll = false)
    {
        if ($subscription->type == 'seasons') {
            if (! $this->lockSubscription($subscription->mid, $subscription->list_id)) {
                return;
            }
            try {
                $this->updateSeasonsAndSeries($subscription->type, $subscription->mid, $subscription->list_id, $pullAll);
            } finally {
                $this->unlockSubscription($subscription->mid, $subscription->list_id);
            }
        } else if ($subscription->type == 'up') {
            if (! $this->lockSubscription($subscription->mid)) {
                return;
            }
            try {
                $this->updateUpVideos($subscription->mid, $pullAll);
            } finally {
                $this->unlockSubscription($subscription->mid);
            }
        }
    }

    public function unlockSubscription($mid, $listId = null)
    {
        redis()->del("subscription:lock:{$mid}:{$listId}");
    }

    protected function lockSubscription($mid, $listId = null)
    {
        $lock = redis()->setnx("subscription:lock:{$mid}:{$listId}", 1);
        if (! $lock) {
            return false;
        }
        redis()->expire("subscription:lock:{$mid}:{$listId}", 1200);
        return true;
    }

    public function updateSeasonsAndSeries($type, $mid, $listId, $pullAll = false)
    {
        if (! in_array($type, ['seasons', 'series'])) {
            throw new \Exception('invalid type');
        }
        $subscription          = Subscription::query()->where('mid', $mid)->where('list_id', $listId)->firstOrNew();
        $subscription->type    = $type;
        $subscription->mid     = $mid;
        $subscription->list_id = $listId;
        $subscription->save();

        if ($type == "series") {
            $dataMeta                  = $this->bilibiliService->getSeriesMeta($listId);
            $listMeta                  = $dataMeta['meta'];
            $subscription->total       = $listMeta['total'];
            $subscription->name        = $listMeta['name'];
            $subscription->description = $listMeta['description'];
            $subscription->cover       = $listMeta['cover'] ?? '';
            $subscription->save();
        }

        $oldSubscription = $subscription->toArray();
        $page            = 1;
        $loaded          = 0;
        while (1) {
            if ($type == 'seasons') {
                $dataList = $this->bilibiliService->getSeasonsList($mid, $listId, $page);
            } else {
                $dataList = $this->bilibiliService->getSeriesList($mid, $listId, $page);
            }
            if (is_array($dataList) && isset($dataList['archives']) && is_array($dataList['archives'])) {

                if (count($dataList['archives']) == 0) {
                    break;
                }

                $loaded += count($dataList['archives']);
                if ($type == "seasons" && isset($dataList['meta'])) {
                    $listMeta                  = $dataList['meta'];
                    $subscription->total       = $listMeta['total'];
                    $subscription->name        = $listMeta['name'];
                    $subscription->description = $listMeta['description'];
                    $subscription->cover       = $listMeta['cover'];
                    $subscription->save();
                }

                if ($type == "series" && $page == 1 && count($dataList['archives']) > 0) {
                    $subscription->cover = $dataList['archives'][0]['pic'] ?? '';
                    $subscription->save();
                }

                $archives = $dataList['archives'];
                foreach ($archives as $archive) {
                    $subscriptionVideo                  = SubscriptionVideo::where('subscription_id', $subscription->id)->where('video_id', $archive['aid'])->firstOrNew();
                    $subscriptionVideo->bvid            = $archive['bvid'];
                    $subscriptionVideo->subscription_id = $subscription->id;
                    $subscriptionVideo->video_id        = $archive['aid'];
                    $subscriptionVideo->save();

                    PullVideoInfoJob::dispatchWithRateLimit($archive['bvid']);
                }

                if ($loaded >= $subscription->total) {
                    break;
                }

                if (! $pullAll) {
                    break;
                }

                $page++;
            }
        }
        $subscription->last_check_at = now();
        $subscription->save();

        event(new SubscriptionUpdated($oldSubscription, $subscription->toArray()));
        return $subscription;
    }

    public function updateUpVideos($mid, $pullAll = false)
    {
        $subscription       = Subscription::where('mid', $mid)->where('type', 'up')->firstOrNew();
        $subscription->type = 'up';
        $subscription->mid  = $mid;
        $subscription->save();
        $oldSubscription = $subscription->toArray();

        $uperCard                  = $this->bilibiliService->getUperCard($mid);
        $subscription->name        = $uperCard['name'] ?? '';
        $subscription->cover       = $uperCard['face'] ?? '';
        $subscription->description = $uperCard['sign'] ?? '';
        $subscription->save();

        $offsetAid = null;
        $loaded    = 0;
        while (1) {
            Log::info('get up videos', ['offsetAid' => $offsetAid, 'loaded' => $loaded]);
            while (1) {
                $retry = 0;
                try {
                    $upVideos = $this->bilibiliService->getUpVideos($mid, $offsetAid);
                } catch (\Exception $e) {
                    Log::error('get up videos error', ['error' => $e->getMessage()]);
                    $retry++;
                    if ($retry > 3) {
                        Log::error('get up videos error: ' . $e->getMessage());
                        throw new \Exception('get up videos error: ' . $e->getMessage());
                    }
                    continue;
                }
                break;
            }
            foreach ($upVideos['list'] as $item) {
                Log::info('up video', ['title' => $item['title']]);
                $aid                                = $item['param'];
                $subscriptionVideo                  = SubscriptionVideo::where('subscription_id', $subscription->id)->where('video_id', $aid)->firstOrNew();
                $subscriptionVideo->bvid            = $item['bvid'];
                $subscriptionVideo->subscription_id = $subscription->id;
                $subscriptionVideo->video_id        = $aid;
                $subscriptionVideo->save();

                // 快速填写一个视频信息
                // 这里获取到的视频都是有效的，所以可以忽略 invalid 处理和封面判断
                $video = Video::withTrashed()->where('id', $aid)->firstOrNew();
                $video->fill([
                    'id'       => $aid,
                    'upper_id' => $mid,
                    'bvid'     => $item['bvid'],
                    'title'    => $item['title'],
                    'cover'    => $item['cover'],
                    'duration' => $item['duration'],
                    'page'     => intval($item['videos']),
                    'pubtime'  => date('Y-m-d H:i:s', $item['ctime']),
                    'link'     => sprintf('https://www.bilibili.com/video/%s', $item['bvid']),
                    'intro'    => '',
                ]);
                $video->save();
                if($video->trashed()){
                    $video->restore();
                }
                event(new VideoUpdated([], $video->getAttributes()));

                PullVideoInfoJob::dispatchWithRateLimit($item['bvid']);
            }
            $loaded += count($upVideos['list']);

            if (! $pullAll) {
                break;
            }

            if (count($upVideos['list']) == 0) {
                break;
            }

            if (! $upVideos['has_next']) {
                break;
            }
            $offsetAid = $upVideos['last_aid'];
        }
        $subscription->total         = $loaded;
        $subscription->last_check_at = now();
        $subscription->save();
        event(new SubscriptionUpdated($oldSubscription, $subscription->toArray()));
        return $subscription;
    }

}
