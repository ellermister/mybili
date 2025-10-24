<?php
namespace App\Jobs;

use App\Models\Subscription;
use App\Services\SubscriptionService;

class UpdateSubscriptionJob extends BaseScheduledRateLimitedJob
{

    public $queue = 'default';

    protected $subscription;
    protected $pullAll;
    /**
     * Create a new job instance.
     */
    public function __construct(Subscription $subscription, bool $pullAll = false)
    {
        $this->subscription = $subscription;
        $this->pullAll      = $pullAll;
    }

    /**
     * 获取限流键名
     */
    protected function getRateLimitKey(): string
    {
        return 'update_subscription_job';
    }

    /**
     * 获取最大处理数量 - 每分钟最多5个订阅更新
     */
    protected function getMaxProcessCount(): int
    {
        return 2;
    }

    /**
     * 获取时间窗口 - 2分钟
     */
    protected function getTimeWindow(): int
    {
        return 120;
    }

    protected function process(): void
    {
        app(SubscriptionService::class)->updateSubscription($this->subscription, $this->pullAll);
    }

    public function displayName(): string
    {
        return 'Update Subscription Job';
    }
}
