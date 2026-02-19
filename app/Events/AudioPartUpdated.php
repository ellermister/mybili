<?php
namespace App\Events;

use Illuminate\Broadcasting\InteractsWithSockets;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

class AudioPartUpdated
{
    use Dispatchable, InteractsWithSockets, SerializesModels;

    public function __construct(public array $oldAudioPart, public array $newAudioPart)
    {
    }

    public function broadcastOn(): array
    {
        return [];
    }
}
