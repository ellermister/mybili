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

        // 首先检测数据库连接
        $isDatabaseConnected = $systemUpgradeService->testDatabaseConnnect();
        if (! $isDatabaseConnected) {
            return response()->view('upgrade.upgrade_database_guide', [
                'database' => $systemUpgradeService->getDatabaseInfo(),
            ]);
        }

        // 检测是否有待执行的迁移
        $hasPendingMigrations = $systemUpgradeService->hasPendingMigrations();
        if ($hasPendingMigrations) {
            return response()->view('upgrade.upgrade_database_guide', [
                'database' => $systemUpgradeService->getDatabaseInfo(),
            ]);
        }

        return $next($request);
    }
}
