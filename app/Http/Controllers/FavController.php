<?php
namespace App\Http\Controllers;

use App\Services\VideoManager\Contracts\FavoriteServiceInterface;
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
}
