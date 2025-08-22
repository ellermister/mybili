<?php

namespace App\Http\Controllers;

use App\Contracts\VideoManagerServiceInterface;
use Illuminate\Http\Request;

class FavController extends Controller
{
    public function __construct(public VideoManagerServiceInterface $videoManagerService)
    {

    }

    /**
     * Display a listing of the resource.
     */
    public function index()
    {
        $data = $this->videoManagerService->getUnifiedContentList();
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
        $id = intval($id);
        $content = $this->videoManagerService->getUnifiedContentDetail($id);
        
        if ($content) {
            // 确保视频关联已加载
            if (isset($content->videos) && method_exists($content->videos, 'load')) {
                $content->videos->load('parts');
            }
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
}
