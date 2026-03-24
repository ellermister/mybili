<?php
namespace App\Http\Controllers;

use App\Services\VideoManager\Contracts\FavoriteServiceInterface;
use App\Services\VideoManager\Contracts\VideoServiceInterface;
use Illuminate\Http\Request;

class FavController extends Controller
{
    public function __construct(public FavoriteServiceInterface $favoriteService)
    {

    }

    /**
     * Display a listing of the resource.
     */
    public function index()
    {
        $data = $this->favoriteService->getUnifiedContentList();
        if ($data) {
            return response()->json($data);
        } else {
            return response()->json([]);
        }
    }

    /**
     * Store a newly created resource in storage.
     */
    public function store(Request $request)
    {

    }

    /**
     * Display the specified resource.
     */
    public function show(string $id)
    {
        $id      = intval($id);
        $content = $this->favoriteService->getUnifiedContentDetail($id);

        if ($content) {
            return response()->json($content);
        } else {
            return response()->json([]);
        }
    }

    /**
     * Update the specified resource in storage.
     */
    public function update(Request $request, string $id)
    {
        //
    }

    /**
     * Remove the specified resource from storage.
     */
    public function destroy(string $id)
    {
        //
    }

    /**
     * 获取收藏夹/订阅的轻量视频列表（走 Redis 缓存）
     */
    public function videos(string $id)
    {
        $id = intval($id);
        $videoService = app(VideoServiceInterface::class);

        if ($id > 0) {
            $data = $videoService->getFavVideosLightweight($id);
        } elseif ($id < 0) {
            $data = $videoService->getSubVideosLightweight(abs($id));
        } else {
            return response()->json([]);
        }

        return response()->json($data);
    }
}
