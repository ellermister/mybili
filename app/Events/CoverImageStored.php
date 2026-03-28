<?php

namespace App\Events;

use App\Models\Cover;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

class CoverImageStored
{
    use Dispatchable, SerializesModels;

    public function __construct(public Cover $cover) {}
}
