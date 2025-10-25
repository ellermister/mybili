<?php
namespace App\Http\Controllers;

use App\Enums\SettingKey;
use App\Services\SettingsService;
use App\Services\TelegramBotService;
use Illuminate\Http\Request;

class SettingsController extends Controller
{

    public function __construct(public SettingsService $settings, public TelegramBotService $telegramBot)
    {

    }

    public function getSettings()
    {
        $presets = [
            SettingKey::DANMAKU_DOWNLOAD_ENABLED->value         => 'off',
            SettingKey::VIDEO_DOWNLOAD_ENABLED->value           => 'off',
            SettingKey::FAVORITE_SYNC_ENABLED->value            => 'off',
            SettingKey::HUMAN_READABLE_NAME_ENABLED->value      => 'off',
            SettingKey::USAGE_ANALYTICS_ENABLED->value          => 'on',
            SettingKey::FAVORITE_EXCLUDE->value                 => [
                'enabled'  => false,
                'selected' => [],
            ],
            SettingKey::MULTI_PARTITION_DOWNLOAD_ENABLED->value => 'off',
            SettingKey::NAME_EXCLUDE->value                     => [
                'contains' => '',
                'regex'    => '',
                'type'     => 'off',
            ],
            SettingKey::SIZE_EXCLUDE->value                     => [
                'custom_size' => 0,
                'type'        => 'off',
            ],
            SettingKey::DURATION_VIDEO_EXCLUDE->value           => [
                'custom_duration' => 0,
                'type'            => 'off',
            ],
            SettingKey::DURATION_VIDEO_PART_EXCLUDE->value      => [
                'custom_duration' => 0,
                'type'            => 'off',
            ],
            SettingKey::TELEGRAM_BOT_API_URL->value             => '',
            SettingKey::TELEGRAM_BOT_ENABLED->value             => 'off',
            SettingKey::TELEGRAM_BOT_TOKEN->value               => '',
            SettingKey::TELEGRAM_CHAT_ID->value                 => '',
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
        if (config('services.bilibili.setting_only')) {
            abort(403);
        }
        $data = $request->validate([
            SettingKey::MULTI_PARTITION_DOWNLOAD_ENABLED->value                 => 'required|string|in:on,off',
            SettingKey::DANMAKU_DOWNLOAD_ENABLED->value                         => 'required|string|in:on,off',
            SettingKey::VIDEO_DOWNLOAD_ENABLED->value                           => 'required|string|in:on,off',
            SettingKey::FAVORITE_SYNC_ENABLED->value                            => 'required|string|in:on,off',
            SettingKey::HUMAN_READABLE_NAME_ENABLED->value                      => 'required|string|in:on,off',
            SettingKey::USAGE_ANALYTICS_ENABLED->value                          => 'required|string|in:on,off',

            SettingKey::NAME_EXCLUDE->value                                     => 'required|array',
            SettingKey::NAME_EXCLUDE->value . '.contains'                       => 'required_if:name_exclude.type,contains|string',
            SettingKey::NAME_EXCLUDE->value . '.regex'                          => 'required_if:name_exclude.type,regex|string',
            SettingKey::NAME_EXCLUDE->value . '.type'                           => 'required|string|in:off,contains,regex',

            SettingKey::SIZE_EXCLUDE->value                                     => 'required|array',
            SettingKey::SIZE_EXCLUDE->value . '.custom_size'                    => 'required_if:size_exclude.type,custom|integer',
            SettingKey::SIZE_EXCLUDE->value . '.type'                           => 'required|string|in:off,1GB,2GB,custom',

            SettingKey::DURATION_VIDEO_EXCLUDE->value                           => 'required|array',
            SettingKey::DURATION_VIDEO_EXCLUDE->value . '.custom_duration'      => 'required_if:duration_video_exclude.type,custom|integer',
            SettingKey::DURATION_VIDEO_EXCLUDE->value . '.type'                 => 'required|string|in:off,30min,60min,120min,custom',

            SettingKey::DURATION_VIDEO_PART_EXCLUDE->value                      => 'required|array',
            SettingKey::DURATION_VIDEO_PART_EXCLUDE->value . '.custom_duration' => 'required_if:duration_video_part_exclude.type,custom|integer',
            SettingKey::DURATION_VIDEO_PART_EXCLUDE->value . '.type'            => 'required|string|in:off,30min,60min,120min,custom',

            SettingKey::FAVORITE_EXCLUDE->value                                 => 'required|array',
            SettingKey::FAVORITE_EXCLUDE->value . '.enabled'                    => 'required|boolean',
            SettingKey::FAVORITE_EXCLUDE->value . '.selected'                   => 'required_if:fav_exclude.enabled,true|array',

            SettingKey::TELEGRAM_BOT_API_URL->value                             => 'nullable|string',
            SettingKey::TELEGRAM_BOT_ENABLED->value                             => 'required|string|in:on,off',
            SettingKey::TELEGRAM_BOT_TOKEN->value                               => 'required_if:telegram_bot_enabled,on|string',
            SettingKey::TELEGRAM_CHAT_ID->value                                 => 'required_if:telegram_bot_enabled,on|string',
        ]);

        foreach ($data as $key => $value) {
            $this->settings->put($key, $value);
        }

        return response()->json(['message' => 'Settings saved successfully']);
    }

    public function testTelegramConnection(Request $request)
    {
        $data = $request->validate([
            SettingKey::TELEGRAM_BOT_TOKEN->value   => 'required|string',
            SettingKey::TELEGRAM_CHAT_ID->value     => 'required|string',
            SettingKey::TELEGRAM_BOT_API_URL->value => 'nullable|string',
        ]);

        $result = $this->telegramBot->testConnection([
            'bot_token' => $data[SettingKey::TELEGRAM_BOT_TOKEN->value],
            'chat_id'   => $data[SettingKey::TELEGRAM_CHAT_ID->value],
            'bot_url'   => $data[SettingKey::TELEGRAM_BOT_API_URL->value] ?? null,
        ]);

        return response()->json(['message' => 'Telegram connection test successful', 'result' => $result]);
    }
}
