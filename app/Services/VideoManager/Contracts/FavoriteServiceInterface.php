<?php
namespace App\Services\VideoManager\Contracts;

use Illuminate\Database\Eloquent\Collection;

interface FavoriteServiceInterface
{
    public function getUnifiedContentList(): array;
    
    public function getUnifiedContentDetail(int $id, array $columns = ['*']): ?object;

    public function getFavorites(): Collection;
}