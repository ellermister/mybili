<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('favorite_list_videos', function (Blueprint $table) {
            $table->integer('favorite_list_id');
            $table->integer('video_id');
            $table->datetimes();
            $table->primary(['favorite_list_id', 'video_id']);
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('favorite_list_videos');
    }
};
