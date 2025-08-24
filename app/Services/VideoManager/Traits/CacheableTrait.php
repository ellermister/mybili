<?php
namespace App\Services\VideoManager\Traits;

trait CacheableTrait
{
    protected function cacheKey(string $prefix, ...$params): string
    {
        return $prefix . ':' . implode(':', $params);
    }

    protected function cacheRemember(string $key, int $ttl, callable $callback)
    {
        return cache()->remember($key, $ttl, $callback);
    }

    protected function cacheForget(string $pattern): void
    {
        // 实现模式匹配的缓存清理
    }
}