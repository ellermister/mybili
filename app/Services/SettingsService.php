<?php
namespace App\Services;

use App\Models\Setting;

class SettingsService
{
    public function put($name, $values)
    {
        Setting::updateOrCreate(['name' => $name], ['value' => $values]);
    }

    public function get($name)
    {
        $setting = Setting::where('name', $name)->first();
        return $setting ? $setting->value : null;
    }
}