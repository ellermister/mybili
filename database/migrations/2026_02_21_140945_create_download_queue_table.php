<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('download_queue', function (Blueprint $table) {
            $table->id();

            // video / audio
            $table->string('type', 20)->default('video');

            $table->unsignedBigInteger('video_id');
            // 视频分P的 video_parts.id；音频时为 null
            $table->unsignedBigInteger('video_part_id')->nullable();

            // pending / running / done / failed / cancelled
            $table->string('status', 20)->default('pending');

            // 数值越大优先级越高
            $table->integer('priority')->default(0);

            $table->text('error_msg')->nullable();

            // 被 Scheduler 取出开始执行的时间
            $table->datetime('scheduled_at')->nullable();
            $table->datetime('completed_at')->nullable();
            $table->timestamps();

            // 用于防重复入队：video:{video_part_id} 或 audio:{video_id}
            $table->string('unique_key')->unique();

            $table->index(['status', 'priority', 'id']);
            $table->index('video_id');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('download_queue');
    }
};
