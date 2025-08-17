<?php
namespace App\Services;

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schema; // Added this import for Schema facade

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

    /**
     * 检测是否有待执行的迁移
     */
    public function hasPendingMigrations()
    {
        try {
            // 方法1: 使用数据库直接查询（推荐）
            if ($this->checkMigrationsViaDatabase()) {
                return true;
            }
            
            // 方法2: 使用改进的 Artisan 调用
            if ($this->checkMigrationsViaArtisan()) {
                return true;
            }
            
            return false;
        } catch (\Exception $e) {
            // 如果检测失败，返回 false
            return false;
        }
    }

    /**
     * 获取迁移状态信息
     */
    public function getMigrationStatus()
    {
        try {
            $databaseMethod = $this->checkMigrationsViaDatabase();
            $artisanMethod = $this->checkMigrationsViaArtisan();
            
            return [
                'has_pending' => $databaseMethod || $artisanMethod,
                'database_method' => $databaseMethod,
                'artisan_method' => $artisanMethod,
                'status_output' => 'Migration check completed'
            ];
        } catch (\Exception $e) {
            return [
                'has_pending' => false,
                'status_output' => 'Unable to check migration status: ' . $e->getMessage()
            ];
        }
    }

    /**
     * 方法1: 通过数据库直接查询检测待执行迁移
     */
    private function checkMigrationsViaDatabase()
    {
        try {
            // 检查 migrations 表是否存在
            if (!Schema::hasTable('migrations')) {
                return true; // 表不存在，需要迁移
            }
            
            // 获取所有迁移文件
            $migrationFiles = $this->getMigrationFiles();
            
            // 获取已执行的迁移
            $executedMigrations = DB::table('migrations')->pluck('migration')->toArray();
            
            // 检查是否有未执行的迁移
            foreach ($migrationFiles as $migration) {
                if (!in_array($migration, $executedMigrations)) {
                    return true; // 发现待执行的迁移
                }
            }
            
            return false;
        } catch (\Exception $e) {
            return false;
        }
    }

    /**
     * 方法2: Artisan 调用方式
     */
    private function checkMigrationsViaArtisan()
    {
        try {
            // 使用 Output 类来捕获输出
            $output = new \Symfony\Component\Console\Output\BufferedOutput;
            Artisan::call('migrate:status', [], $output);
            $content = $output->fetch();
            
            return strpos($content, 'Pending') !== false;
        } catch (\Exception $e) {
            return false;
        }
    }

    /**
     * 获取迁移文件列表
     */
    private function getMigrationFiles()
    {
        $migrationsPath = database_path('migrations');
        $files = glob($migrationsPath . '/*.php');
        
        $migrations = [];
        foreach ($files as $file) {
            $filename = basename($file, '.php');
            // 提取迁移名称（去掉时间戳前缀）
            if (preg_match('/^\d{4}_\d{2}_\d{2}_\d{6}_/', $filename)) {
                $migrations[] = $filename;
            }
        }
        
        return $migrations;
    }
}
