<?php
namespace App\Services;

use App\Enums\SettingKey;
use App\Models\Video;
use Carbon\Carbon;
use Log;

class DownloadFilterService
{
    public function __construct(public SettingsService $settings)
    {

    }

    public function shouldExcludeByName(string $name)
    {
        $nameExclude = $this->settings->get(SettingKey::NAME_EXCLUDE);
        if (is_null($nameExclude)) {
            throw new \Exception('name_exclude is not set');
        }
        if ($nameExclude['type'] === 'off') {
            return false;
        }
        if ($nameExclude['contains'] && strpos($name, $nameExclude['contains']) !== false) {
            return true;
        }
        try {
            if ($nameExclude['regex'] && preg_match('/' . $nameExclude['regex'] . '/', $name)) {
                return true;
            }
        } catch (\Throwable $e) {
            Log::error('Regex error', ['error' => $e->getMessage(), 'name' => $name]);
            return false;
        }
        return false;
    }

    public function shouldExcludeBySize(int $size)
    {
        $sizeExclude = $this->settings->get(SettingKey::SIZE_EXCLUDE);
        if ($sizeExclude['type'] === 'off') {
            return false;
        }
        if ($sizeExclude['type'] === '1GB') {
            return $size > 1024 * 1024 * 1024;
        }
        if ($sizeExclude['type'] === '2GB') {
            return $size > 2 * 1024 * 1024 * 1024;
        }
        if ($sizeExclude['type'] === 'custom') {
            return $size > ($sizeExclude['custom_size'] * 1024 * 1024 * 1024);
        }
        return false;
    }

    public function shouldExcludeByFav(int $favId)
    {
        $favExclude = $this->settings->get(SettingKey::FAVORITE_EXCLUDE);
        if (! $favExclude || $favExclude['enabled'] === false) {
            return false;
        }
        return in_array($favId, $favExclude['selected']);
    }

    /**
     * 只要视频关联的收藏夹或订阅中，有任意一个未被排除，就允许下载。
     * 若存在关联但全部被排除，则返回 true（应排除下载）。
     */
    public function shouldExcludeByVideo(Video $video): bool
    {
        $video->loadMissing('favorite', 'subscriptions');

        $hasAnyRelation = $video->favorite->isNotEmpty() || $video->subscriptions->isNotEmpty();
        if (! $hasAnyRelation) {
            return false;
        }

        $hasFavNotExcluded = $video->favorite->contains(
            fn ($fav) => ! $this->shouldExcludeByFav($fav->id)
        );
        $hasSubNotExcluded = $video->subscriptions->contains(
            fn ($sub) => ! $this->shouldExcludeByFav(-$sub->id)
        );

        return ! $hasFavNotExcluded && ! $hasSubNotExcluded;
    }

    public function isMultiPEnabled()
    {
        $multiP = $this->settings->get(SettingKey::MULTI_PARTITION_DOWNLOAD_ENABLED);
        return $multiP === 'on';
    }

    public function shouldExcludeByDuration(int $duration)
    {
        $durationExclude = $this->settings->get(SettingKey::DURATION_VIDEO_EXCLUDE);
        if (!$durationExclude) {
            return false;
        }
        if ($durationExclude['type'] === 'off') {
            return false;
        }
        if ($durationExclude['type'] === '30min') {
            return $duration > 30 * 60;
        }
        if ($durationExclude['type'] === '60min') {
            return $duration > 60 * 60;
        }
        if ($durationExclude['type'] === 'custom') {
            return $duration > $durationExclude['custom_duration'] * 60;
        }
        return false;
    }

    public function shouldExcludeByDurationPart(int $duration)
    {
        $durationPartExclude = $this->settings->get(SettingKey::DURATION_VIDEO_PART_EXCLUDE);
        if (!$durationPartExclude) {
            return false;
        }
        if ($durationPartExclude['type'] === 'off') {
            return false;
        }
        if ($durationPartExclude['type'] === '30min') {
            return $duration > 30 * 60;
        }
        if ($durationPartExclude['type'] === '60min') {
            return $duration > 60 * 60;
        }
        if ($durationPartExclude['type'] === 'custom') {
            return $duration > $durationPartExclude['custom_duration'] * 60;
        }
        return false;
    }

    public function shouldExcludeByFavTime(Video $video): bool
    {
        $config = $this->settings->get(SettingKey::FAV_TIME_EXCLUDE);
        if (! $config || ($config['type'] ?? 'off') === 'off') {
            return false;
        }

        $favTime = $video->fav_time; // accessor: timestamp|null
        if (! $favTime) {
            // 如果收藏时间为空，则使用视频发布时间
            $favTime = $video->pubtime;
            if (! $favTime) {
                return false;
            }
        }

        $type = (string) ($config['type'] ?? 'off');
        $cutoff = match ($type) {
            '1m' => Carbon::now()->subMonthNoOverflow()->startOfDay(),
            '3m' => Carbon::now()->subMonthsNoOverflow(3)->startOfDay(),
            '6m' => Carbon::now()->subMonthsNoOverflow(6)->startOfDay(),
            'custom' => isset($config['custom_date']) && $config['custom_date']
                ? Carbon::createFromFormat('Y-m-d', (string) $config['custom_date'])->startOfDay()
                : null,
            default => null,
        };

        if (! $cutoff) {
            return false;
        }

        return Carbon::createFromTimestamp((int) $favTime)->lt($cutoff);
    }
}
