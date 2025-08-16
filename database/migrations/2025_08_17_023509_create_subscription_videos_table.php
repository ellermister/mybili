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
        Schema::create('subscription_videos', function (Blueprint $table) {
            $table->integer('subscription_id');
            $table->integer('video_id');
            $table->string('bvid');
            $table->primary(['subscription_id', 'video_id']);
            $table->datetimes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('subscription_videos');
    }
};
