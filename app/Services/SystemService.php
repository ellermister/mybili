<?php

namespace App\Services;

use App\Models\Danmaku;
use App\Models\FavoriteList;
use App\Models\Video;
use App\Models\VideoPart;
use Carbon\Carbon;
use Illuminate\Support\Facades\DB;
use PDO;

class SystemService
{
    public function getSystemInfo(): array
    {
        $info = [
            'app_version'      => config('app.version'),
            'php_version'      => phpversion(),
            'laravel_version'  => app()->version(),
            'database_version' => DB::connection()->getPdo()->getAttribute(PDO::ATTR_SERVER_VERSION),
            'timezone'         => config('app.timezone'),
            'time_now'         => Carbon::now()->toDateTimeString(),

            //usage
            'database_usage'   => [
                'favorite_lists' => FavoriteList::count(),
                'videos'         => Video::count(),
                'video_parts'    => VideoPart::count(),
                'danmaku'        => Danmaku::count(),
                'db_size'        => $this->getDatabaseSize(),
            ],
            'media_usage'      => [
                'videos_size' => $this->getMediaSize('videos'),
                'images_size' => $this->getMediaSize('images'),
            ],
        ];
        return $info;
    }

    public function getDatabaseSize(): int
    {
        // 如果是sqlite
        if (DB::connection()->getPdo()->getAttribute(PDO::ATTR_DRIVER_NAME) == 'sqlite') {
            return DB::table('sqlite_master')->where('type', 'table')->sum('rootpage') * 1024;
        }
        return 0;
    }

    public function getMediaSize(string $directory): string
    {
        $mediasPath = storage_path("app/public/".$directory);
        exec(sprintf('du -sh %s', escapeshellarg($mediasPath)), $output, $result);
        if ($result !== 0) {
            return '0';
        }
        $output = explode("\t", $output[0]);
        return $output[0] ?? '0';
    }
}
