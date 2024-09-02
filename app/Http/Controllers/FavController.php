<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

class FavController extends Controller
{
    /**
     * Display a listing of the resource.
     */
    public function index()
    {
        $result = redis()->get('fav_list');
        $data   = json_decode($result, true);
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
        $result = redis()->get(sprintf('fav_list:%d', $id));
        $data   = json_decode($result, true);
        if ($data) {
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
