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
        $data = $this->videoManagerService->getFavList();
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
        $data = $this->videoManagerService->getVideoListByFav($id);
        if ($data && is_array($data)) {
            usort($data, function ($a, $b) {
                if ($a['fav_time'] == $b['fav_time']) {
                    return 0;
                }
                return $a['fav_time'] > $b['fav_time'] ? -1 : 1;
            });
            return response()->json($data);
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
