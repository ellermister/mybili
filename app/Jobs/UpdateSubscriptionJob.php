<?php
namespace App\Jobs;

use App\Models\Subscription;
use App\Services\SubscriptionService;

class UpdateSubscriptionJob extends BaseScheduledRateLimitedJob
{
    protected $subscription;
    protected $pullAll;

    public function __construct(Subscription $subscription, bool $pullAll = false)
    {
        $this->subscription = $subscription;
        $this->pullAll      = $pullAll;
    }

    protected function getRateLimitKey(): string
    {
        return 'update_job';
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
