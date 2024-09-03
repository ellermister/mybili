<?php

namespace App\Http\Controllers;

use GuzzleHttp\Client;
use Illuminate\Http\Request;

class CookieController extends Controller
{

    protected $path = null;
    public function __construct()
    {
        $this->path = storage_path('app/cookie.txt');
    }

    public function uploadCookieFile(Request $request)
    {
        $file = $request->file('file');
        if ($file) {
            if ($file->getMimeType() == 'text/plain') {
                $file->move(storage_path('app/'), 'cookie.txt');
                return response()->json();
            }
        }
        abort(400);
    }

    public function checkFileExist()
    {
        return response()->json([
            'exist' => is_file($this->path),
        ]);
    }

    public function checkCookieValid()
    {
        $jar      = parse_netscape_cookie_file($this->path);
        $client   = new Client();
        $response = $client->request('GET', 'https://api.bilibili.com/x/web-interface/nav', [
            'headers' => [
                'user-agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36',
                'referer'    => 'https://space.bilibili.com/',
            ],
            'cookies' => $jar,
        ]);

        $body = $response->getBody()->getContents();
        // dump($response->getStatusCode());
        // dd($body);

        $data    = json_decode($body, true);
        $isLogin = false;
        if ($data['data']['isLogin'] === true) {
            $isLogin = true;
        }
        return response()->json([
            'logged' => $isLogin,
        ]);
    }
}