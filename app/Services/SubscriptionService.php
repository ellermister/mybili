<?php
namespace App\Services;

use App\Events\SubscriptionUpdated;
use App\Jobs\PullVideoInfoJob;
use App\Models\Subscription;
use App\Models\SubscriptionVideo;
use App\Services\BilibiliService;
use DB;
use Log;

class SubscriptionService
{
    public function __construct(public BilibiliService $bilibiliService)
    {
    }

    public function addSubscription($type, $url)
    {
        if ($type == 'seasons') {
            if (! preg_match('#/(\d+)/lists/(\d+)#', $url, $matches)) {
                throw new \Exception('invalid seasons url');
            }
            $mid               = $matches[1];
            $seasonId          = $matches[2];
            $subscription      = $this->updateSeasonsAndSeries('seasons', $mid, $seasonId, true);
            $subscription->url = $url;
            $subscription->save();
            return $subscription;
        } else if ($type == 'series') {
            if (! preg_match('#/(\d+)/lists/(\d+)#', $url, $matches)) {
                throw new \Exception('invalid series url');
            }
            $mid               = $matches[1];
            $seriesId          = $matches[2];
            $subscription      = $this->updateSeasonsAndSeries('series', $mid, $seriesId, true);
            $subscription->url = $url;
            $subscription->save();
            return $subscription;
        } else {
            if (! preg_match('#/(\d+)/upload#', $url, $matches)) {
                throw new \Exception('invalid up url');
            }
            $mid               = $matches[1];
            $subscription      = $this->updateUpVideos($mid, true);
            $subscription->url = $url;
            $subscription->save();
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

    public function deleteSubscription(Subscription $subscription)
    {
        if ($subscription->videos()->count() > 0) {
            throw new \Exception('subscription has videos, cannot delete');
        }
        $subscription->delete();
    }

    public function changeSubscription(Subscription $subscription, array $data)
    {
        $subscription->fill($data);
        $subscription->save();
        return $subscription;
    }

    public function updateSubscription()
    {
        $subscriptions = Subscription::where('status', Subscription::STATUS_ACTIVE)->where('last_check_at', '<', now()->subMinutes(20))->get();
        foreach ($subscriptions as $subscription) {
            if ($subscription->type == 'seasons') {
                if (! $this->lockSubscription($subscription->mid, $subscription->list_id)) {
                    continue;
                }
                try {
                    $this->updateSeasonsAndSeries($subscription->type, $subscription->mid, $subscription->list_id, false);
                } finally {
                    $this->unlockSubscription($subscription->mid, $subscription->list_id);
                }
            } else if ($subscription->type == 'up') {
                if (! $this->lockSubscription($subscription->mid)) {
                    continue;
                }
                try {
                    $this->updateUpVideos($subscription->mid, false);
                } finally {
                    $this->unlockSubscription($subscription->mid);
                }
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
        $subscription            = Subscription::query()->where('mid', $mid)->where('list_id', $listId)->firstOrNew();
        $subscription->type      = $type;
        $subscription->mid       = $mid;
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

                if ($type == "series" && $page == 1&& count($dataList['archives']) > 0) {
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
            $upVideos = $this->bilibiliService->getUpVideos($mid, $offsetAid);
            foreach ($upVideos['list'] as $item) {
                Log::info('up video', ['title' => $item['title']]);
                $aid                                = $item['param'];
                $subscriptionVideo                  = SubscriptionVideo::where('subscription_id', $subscription->id)->where('video_id', $aid)->firstOrNew();
                $subscriptionVideo->bvid            = $item['bvid'];
                $subscriptionVideo->subscription_id = $subscription->id;
                $subscriptionVideo->video_id        = $aid;
                $subscriptionVideo->save();

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
