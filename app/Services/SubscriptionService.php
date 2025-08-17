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
            $subscription      = $this->updateSeasons($mid, $seasonId, true);
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
                if (! $this->lockSubscription($subscription->mid, $subscription->season_id)) {
                    continue;
                }
                try {
                    $this->updateSeasons($subscription->mid, $subscription->season_id, false);
                } finally {
                    $this->unlockSubscription($subscription->mid, $subscription->season_id);
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

    public function unlockSubscription($mid, $seasonId = null)
    {
        redis()->del("subscription:lock:{$mid}:{$seasonId}");
    }

    protected function lockSubscription($mid, $seasonId = null)
    {
        $lock = redis()->setnx("subscription:lock:{$mid}:{$seasonId}", 1);
        if (! $lock) {
            return false;
        }
        redis()->expire("subscription:lock:{$mid}:{$seasonId}", 1200);
        return true;
    }

    public function updateSeasons($mid, $seasonId, $pullAll = false)
    {
        $subscription            = Subscription::where('mid', $mid)->where('season_id', $seasonId)->firstOrNew();
        $subscription->type      = 'seasons';
        $subscription->mid       = $mid;
        $subscription->season_id = $seasonId;
        $subscription->save();

        $oldSubscription = $subscription->toArray();
        $page            = 1;
        $loaded          = 0;
        while (1) {
            $seasonsList = $this->bilibiliService->getSeasonsList($mid, $seasonId, $page);
            if (isset($seasonsList['data']) && is_array($seasonsList['data'])) {

                if (count($seasonsList['data']['archives']) == 0) {
                    break;
                }

                if (! isset($seasonsList['data']['meta'])) {
                    Log::error('get seasons list failed', ['mid' => $mid, 'season_id' => $seasonId, 'page' => $page, 'response' => $seasonsList]);
                    throw new \Exception('get seasons list failed');
                }

                $loaded += count($seasonsList['data']['archives']);
                DB::transaction(function () use ($subscription, $seasonsList) {
                    $subscription->total       = $seasonsList['data']['meta']['total'];
                    $subscription->name        = $seasonsList['data']['meta']['name'];
                    $subscription->description = $seasonsList['data']['meta']['description'];
                    $subscription->cover       = $seasonsList['data']['meta']['cover'];
                    $subscription->save();

                    $archives = $seasonsList['data']['archives'];
                    foreach ($archives as $archive) {
                        $subscriptionVideo                  = SubscriptionVideo::where('subscription_id', $subscription->id)->where('video_id', $archive['aid'])->firstOrNew();
                        $subscriptionVideo->bvid            = $archive['bvid'];
                        $subscriptionVideo->subscription_id = $subscription->id;
                        $subscriptionVideo->video_id        = $archive['aid'];
                        $subscriptionVideo->save();

                        PullVideoInfoJob::dispatchWithRateLimit($archive['bvid']);
                    }

                });

                if ($loaded >= $subscription->total) {
                    break;
                }

                if (! $pullAll) {
                    break;
                }

                $page++;
            } else {
                Log::error('get seasons list failed', ['mid' => $mid, 'season_id' => $seasonId, 'page' => $page, 'response' => $seasonsList]);
                throw new \Exception('get seasons list failed');
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
            DB::transaction(function () use ($subscription, $upVideos) {
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
            });
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
