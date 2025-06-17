<?php
namespace App\Http\Controllers;

use App\Enums\SettingKey;
use App\Services\BilibiliService;
use App\Services\SettingsService;
use GuzzleHttp\Client;
use Illuminate\Http\Request;

class CookieController extends Controller
{

    public function __construct(public SettingsService $settingsService, public BilibiliService $bilibiliService)
    {
    }

    public function uploadCookieFile(Request $request)
    {
        $file = $request->file('file');
        if ($file) {
            if ($file->getMimeType() == 'text/plain') {
                // 将 $file 写入并删除文件
                $this->settingsService->put(SettingKey::COOKIES_CONTENT, file_get_contents($file->getRealPath()));
                $file->delete();
                return response()->json([
                    'success' => true,
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
        $jar      = parse_netscape_cookie_content($this->settingsService->get(SettingKey::COOKIES_CONTENT));
        $client   = new Client();
        $response = $client->request('GET', 'https://api.bilibili.com/x/web-interface/nav', [
            'headers' => [
                'user-agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36',
                'referer'    => 'https://space.bilibili.com/',
            ],
            'cookies' => $jar,
        ]);

        $body = $response->getBody()->getContents();
        $data = json_decode($body, true);

        $isLogin = false;
        if ($data['data']['isLogin'] === true) {
            $isLogin = true;
        }

        return response()->json([
            'logged' => $isLogin,
        ]);
    }
}
