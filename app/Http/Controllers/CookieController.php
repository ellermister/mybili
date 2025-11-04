<?php
namespace App\Http\Controllers;

use App\Enums\SettingKey;
use App\Services\BilibiliService;
use App\Services\CookieControlService;
use App\Services\SettingsService;
use Illuminate\Http\Request;

class CookieController extends Controller
{

    public function __construct(
        public SettingsService $settingsService,
        public BilibiliService $bilibiliService,
        public CookieControlService $cookieControlService
    ) {
    }

    public function uploadCookieFile(Request $request)
    {
        if (config('services.bilibili.setting_read_only')) {
            abort(403);
        }

        $file = $request->file('file');
        if ($file) {
            if ($file->getMimeType() == 'text/plain') {
                // 读取文件内容并存储到设置中
                $this->settingsService->put(SettingKey::COOKIES_CONTENT, file_get_contents($file->getRealPath()));
                
                // 上传成功后，校验 Cookie 有效性
                $jar      = parse_netscape_cookie_content($this->settingsService->get(SettingKey::COOKIES_CONTENT));
                $isValid = $this->bilibiliService->checkCookieExpired($jar);
                if($isValid) {
                    $this->cookieControlService->clearCookieExpiredNotification();
                }
                
                return response()->json([
                    'success' => true,
                    'valid'   => $isValid,
                ]);
            }
        }
        abort(400);
    }

    public function checkFileExist()
    {
        return response()->json([
            'exist' => empty($this->settingsService->get(SettingKey::COOKIES_CONTENT)) ? false : true,
        ]);
    }

    public function checkCookieValid()
    {
        $cookies      = parse_netscape_cookie_content($this->settingsService->get(SettingKey::COOKIES_CONTENT));
        $isValid = $this->bilibiliService->checkCookieExpired($cookies);
        return response()->json([
            'logged' => $isValid,
        ]);
    }
}
