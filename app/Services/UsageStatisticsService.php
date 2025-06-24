<?php
namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;

class UsageStatisticsService
{
    private const UMANI_ENDPOINT = 'https://api-gateway.umami.dev/api/send';
    private $websiteId;
    private const CACHE_KEY      = 'usage_stats_id';
    private const CACHE_TTL      = 86400 * 365; // 1 year

    public function __construct()
    {
        $this->websiteId = config('app.website_id', '');
    }

    private function ensureAnonymousId(): void
    {
        if (! Cache::has(self::CACHE_KEY)) {
            Cache::put(self::CACHE_KEY, Str::uuid()->toString(), self::CACHE_TTL);
        }
    }

    public function getAnonymousId(): string
    {
        return (string) Cache::get(self::CACHE_KEY);
    }

    public function collectStats(): array
    {
        return [
            'type'    => 'event',
            'payload' => [
                'website'  => $this->websiteId,
                'hostname' => config('app.url'),
                'name'     => 'app_usage',
                'url'      => '/',
                'data'     => [
                    'anonymous_id' => $this->getAnonymousId(),
                    'version'      => config('app.version', '1.0.0'),
                    'environment'  => config('app.env'),
                    'is_docker'    => $this->isRunningInDocker(),
                    'os'           => PHP_OS,
                    'php_version'  => PHP_VERSION,
                ],
            ],
        ];
    }

    private function isRunningInDocker(): bool
    {
        return file_exists('/.dockerenv');
    }

    public function sendStats(): void
    {
        try {
            if (empty($this->websiteId)) {
                return;
            }

            $this->ensureAnonymousId();

            $stats = $this->collectStats();

            $response = Http::withHeaders([
                'User-Agent'   => 'Mozilla/5.0 (Apple-iPhone7C2/1202.466; U; CPU like Mac OS X; en) AppleWebKit/420+ (KHTML, like Gecko) Version/3.0 Mobile/1A543 Safari/419.3 U/' . $this->getAnonymousId(),
                'Content-Type' => 'application/json',
                'Accept'       => 'application/json',
            ])->post(self::UMANI_ENDPOINT, $stats);

            if ($response->successful()) {
                Log::info('Usage statistics sent successfully', [
                    'anonymous_id' => $stats['payload']['data']['anonymous_id'],
                    'response'     => $response->body(),
                ]);
            } else {
                Log::error('Failed to send usage statistics', [
                    'status'       => $response->status(),
                    'body'         => $response->body(),
                    'anonymous_id' => $stats['payload']['data']['anonymous_id'],
                ]);
            }
        } catch (\Exception $e) {
            Log::error('Failed to send usage statistics', [
                'error'        => $e->getMessage(),
                'anonymous_id' => $this->getAnonymousId(),
            ]);
        }
    }
}
