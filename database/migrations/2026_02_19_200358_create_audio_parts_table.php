<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('audio_parts', function (Blueprint $table) {
            $table->id();
            $table->unsignedBigInteger('video_id')->comment('关联 videos.id，对音频而言即 sid');
            $table->unsignedBigInteger('sid')->comment('B站音频 sid');
            $table->integer('duration')->default(0)->comment('时长（秒）');
            $table->datetime('audio_downloaded_at')->nullable();
            $table->string('audio_download_path')->nullable();
            $table->timestamps();

            $table->unique('video_id');
            $table->index('sid');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('audio_parts');
    }
};
