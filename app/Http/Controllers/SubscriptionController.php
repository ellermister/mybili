<?php

namespace App\Http\Controllers;

use App\Models\Subscription;
use App\Services\SubscriptionService;
use App\Services\ColorExtractionService;
use Illuminate\Http\Request;
use Storage;

class SubscriptionController extends Controller
{
    public function __construct(
        public SubscriptionService $subscriptionService,
        public ColorExtractionService $colorExtractionService
    ) {
    }

    public function index()
    {
        $sub = $this->subscriptionService->getSubscriptions()->toArray();
        
        foreach ($sub as $key => $subscription) {
            // 只为 UP 主类型提取颜色
            if ($subscription['type'] === 'up' && isset($subscription['cover_info']['path'])) {
                $localPath = Storage::disk('public')->path($subscription['cover_info']['path']);

                
                // 提取颜色
                $color = $this->colorExtractionService->extractDominantColor($localPath);
                
                if ($color) {
                    $sub[$key]['dominant_color'] = $color['hex'];
                } else {
                    $sub[$key]['dominant_color'] = null;
                }
            } else {
                $sub[$key]['dominant_color'] = null;
            }
        }
        
        return response()->json($sub);
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
