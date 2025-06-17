<?php

namespace App\Http\Middleware;

use App\Services\SystemUpgradeService;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class CheckDatabaseIsOkMiddleware
{
    /**
     * Handle an incoming request.
     *
     * @param  \Closure(\Illuminate\Http\Request): (\Symfony\Component\HttpFoundation\Response)  $next
     */
    public function handle(Request $request, Closure $next): Response
    {
        $systemUpgradeService = new SystemUpgradeService();
        $isDatabaseConnected = $systemUpgradeService->testDatabaseConnnect();
        if (!$isDatabaseConnected) {
            return response()->view('upgrade.upgrade_database_guide', [
                'database' => $systemUpgradeService->getDatabaseInfo(),
            ]);
        }
        return $next($request);
    }
}
