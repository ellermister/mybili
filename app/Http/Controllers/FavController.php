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
        $fav = $this->videoManagerService->getFavDetail(intval($id));
        if ($fav) {
            $fav->load('videos');
            return response()->json($fav);
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
