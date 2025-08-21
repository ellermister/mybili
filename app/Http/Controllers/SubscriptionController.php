<?php

namespace App\Http\Controllers;

use App\Models\Subscription;
use App\Services\SubscriptionService;
use Illuminate\Http\Request;

class SubscriptionController extends Controller
{
    public function __construct(public SubscriptionService $subscriptionService)
    {
    }

    public function index()
    {
        return response()->json($this->subscriptionService->getSubscriptions());
    }

    public function store(Request $request)
    {
        if (config('services.bilibili.setting_only')) {
            abort(403);
        }
        $this->subscriptionService->addSubscription($request->type, $request->url);
        return response()->json(['message' => 'Subscription added successfully']);
    }

    public function update(Request $request, int $id)
    {
        if (config('services.bilibili.setting_only')) {
            abort(403);
        }
        $subscription = Subscription::find($id);
        if (!$subscription) {
            return response()->json(['message' => 'Subscription not found'], 404);
        }
        $subscription = $this->subscriptionService->changeSubscription($subscription, $request->all());
        return response()->json($subscription);
    }

    public function destroy(int $id)
    {
        if (config('services.bilibili.setting_only')) {
            abort(403);
        }
        $subscription = Subscription::find($id);
        if (!$subscription) {
            return response()->json(['message' => 'Subscription not found'], 404);
        }
        $this->subscriptionService->deleteSubscription($subscription);
        return response()->json(['message' => 'Subscription deleted successfully']);
    }

    public function show(int $id)
    {
        $subscription = Subscription::find($id);
        if (!$subscription) {
            return response()->json(['message' => 'Subscription not found'], 404);
        }
        $subscription->load(['videos' => function($query) {
            $query->orderBy('pubtime', 'desc');
        }]);
        return response()->json($subscription);
    }
}
