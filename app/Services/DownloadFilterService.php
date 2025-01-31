<?php

namespace App\Services;

use Log;

class DownloadFilterService
{
    public function __construct(public SettingsService $settings)
    {

    }

    public function shouldExcludeByName(string $name)
    {
        $nameExclude = $this->settings->get('nameExclude');
        if ($nameExclude['type'] === 'off') {
            return false;
        }
        if ($nameExclude['contains'] && strpos($name, $nameExclude['contains']) !== false) {
            return true;
        }
        try{
            if ($nameExclude['regex'] && preg_match('/'.$nameExclude['regex'].'/', $name)) {
                return true;
            }
        }catch(\Throwable $e){
            Log::error('Regex error', ['error' => $e->getMessage(), 'name' => $name]);
            return false;
        }
        return false;
    }

    public function shouldExcludeBySize(int $size)
    {
        $sizeExclude = $this->settings->get('sizeExclude');
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
            return $size > ($sizeExclude['customSize'] * 1024 * 1024 * 1024);
        }
        return false;
    }

    public function shouldExcludeByFav(int $favId)
    {
        $favExclude = $this->settings->get('favExclude');
        if (!$favExclude || $favExclude['enabled'] === false) {
            return false;
        }
        return in_array($favId, $favExclude['selected']);
    }

    public function isMultiPEnabled()
    {
        $multiP = $this->settings->get('MultiPartitionCache');
        return $multiP === 'on';
    }
}   
