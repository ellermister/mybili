<?php
namespace App\Http\Controllers;

use App\Services\SettingsService;
use Illuminate\Http\Request;

class SettingsController extends Controller
{

    public function __construct(public SettingsService $settings)
    {

    }

    public function getSettings()
    {
        $presets = [
            'danmaku_cache'         => 'on',
            'fav_exclude'           => [
                'enabled'  => false,
                'selected' => [],
            ],
            'multi_partition_cache' => 'off',
            'name_exclude'          => [
                'contains' => '',
                'regex'    => '',
                'type'     => 'off',
            ],
            'size_exclude'          => [
                'custom_size' => 0,
                'type'        => 'off',
            ],
        ];

        foreach ($presets as $key => $value) {
            $got = $this->settings->get($key);
            if ($got) {
                $presets[$key] = $got;
            }
        }

        return response()->json($presets);
    }

    public function saveSettings(Request $request)
    {
        $data = $request->validate([
            'multi_partition_cache'    => 'required|string|in:on,off',
            'danmaku_cache'            => 'required|string|in:on,off',

            'name_exclude'             => 'required|array',
            'name_exclude.contains'    => 'required_if:name_exclude.type,contains|string',
            'name_exclude.regex'       => 'required_if:name_exclude.type,regex|string',
            'name_exclude.type'        => 'required|string|in:off,contains,regex',

            'size_exclude'             => 'required|array',
            'size_exclude.custom_size' => 'required_if:size_exclude.type,custom|integer',
            'size_exclude.type'        => 'required|string|in:off,1GB,2GB,custom',

            'fav_exclude'              => 'required|array',
            'fav_exclude.enabled'      => 'required|boolean',
            'fav_exclude.selected'     => 'required_if:fav_exclude.enabled,true|array',
        ]);

        foreach ($data as $key => $value) {
            $this->settings->put($key, $value);
        }

        return response()->json(['message' => 'Settings saved successfully']);
    }
}
