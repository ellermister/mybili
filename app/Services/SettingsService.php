<?php
namespace App\Services;
class SettingsService
{
    public function __construct()
    {
        
    }

    public function put($name, $values)
    {
        redis()->hset('settings', $name, json_encode($values));
    }

    public function get($name)
    {
        $value = redis()->hget('settings', $name);
        return json_decode($value, true);
    }
}