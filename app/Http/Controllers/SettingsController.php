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
            'danmakuCache'        => 'on',
            'favExclude'          => [
                'enabled'  => false,
                'selected' => [],
            ],
            'MultiPartitionCache' => 'off',
            'nameExclude'         => [
                'contains' => '',
                'regex'    => '',
                'type'     => 'off',
            ],
            'sizeExclude'         => [
                'customSize' => 0,
                'type'       => 'off',
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
            'MultiPartitionCache'    => 'required|string|in:on,off',
            'danmakuCache'           => 'required|string|in:on,off',

            'nameExclude'            => 'required|array',
            'nameExclude.contains'   => 'required_if:nameExclude.type,contains|string',
            'nameExclude.regex'      => 'required_if:nameExclude.type,regex|string',
            'nameExclude.type'       => 'required|string|in:off,contains,regex',

            'sizeExclude'            => 'required|array',
            'sizeExclude.customSize' => 'required_if:sizeExclude.type,custom|integer',
            'sizeExclude.type'       => 'required|string|in:off,1GB,2GB,custom',

            'favExclude'             => 'required|array',
            'favExclude.enabled'     => 'required|boolean',
            'favExclude.selected'    => 'required_if:favExclude.enabled,true|array',
        ]);

        foreach ($data as $key => $value) {
            $this->settings->put($key, $value);
        }

        return response()->json(['message' => 'Settings saved successfully']);
    }
}
