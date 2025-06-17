<?php

namespace App\Http\Middleware;

use App\Enums\SettingKey;
use App\Services\SettingsService;
use App\Services\SystemUpgradeService;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class CheckCanUpgradedFromRedisMiddleware
{
    /**
     * Handle an incoming request.
     *
     * @param  \Closure(\Illuminate\Http\Request): (\Symfony\Component\HttpFoundation\Response)  $next
     */
    public function handle(Request $request, Closure $next): Response
    {
        $systemUpgradeService = new SystemUpgradeService();
        $settingService = app(SettingsService::class);
        $isUsingOldDatabase = $systemUpgradeService->checkIsUsingOldDatabase();
        if ($isUsingOldDatabase && $settingService->get(SettingKey::INSTALLED_DATETIME) == null) {
            return response()->view('upgrade.upgrade_from_redis');
        }
        return $next($request);
    }
}
