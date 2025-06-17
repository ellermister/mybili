<?php
namespace App\Services;

use Illuminate\Support\Facades\DB;

class SystemUpgradeService
{
    public function checkIsUsingOldDatabase()
    {
        // 检测是否存在 settings, 以及扫描是否存在 video:* 的 key，如果存在则就是使用旧redis数据库
        $existsSettings = redis()->exists('settings');
        if (! $existsSettings) {
            return false;
        }
        $keys = redis()->keys('video:*');
        if (count($keys) == 0) {
            return false;
        }
        return true;
    }

    public function testDatabaseConnnect()
    {
        try {
            DB::connection()->getPdo();
            return true;
        } catch (\Exception $e) {
            return false;
        }
    }

    public function getDatabaseInfo()
    {
        $databaseDefault = config('database.default');
        $databaseConnection = config('database.connections.' . $databaseDefault);
        $databaseDriver     = $databaseConnection['driver'];
        $databaseUrl        = $databaseConnection['url'];
        $databaseDatabase   = $databaseConnection['database'];
        return [
            'driver'   => $databaseDriver,
            'url'      => $databaseUrl,
            'database' => $databaseDatabase,
        ];
    }
}
