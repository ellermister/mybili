<?php
namespace App\Services\VideoManager\Actions\Favorite;

use App\Events\FavoriteUpdated;
use App\Models\FavoriteList;
use App\Services\BilibiliService;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class UpdateFavoritesAction
{
    public function __construct(
        public BilibiliService $bilibiliService
    ) {
    }

    /**
     * 更新收藏夹列表
     */
    public function execute(): void
    {
        Log::info('Update favorites start');
        $favorites = $this->bilibiliService->pullFav();

        DB::transaction(function () use ($favorites) {
            array_map(function ($item) {
                $favorite = FavoriteList::query()->where('id', $item['id'])->first();
                if (! $favorite) {
                    $favorite = new FavoriteList();
                }

                $oldFav = $favorite->toArray();

                $favorite->fill($item);
                $favorite->save();

                event(new FavoriteUpdated($oldFav, $item));

                return $item;
            }, $favorites);
        });

        Log::info('Update favorites success');
    }
}
