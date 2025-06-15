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
        Schema::create('danmaku', function (Blueprint $table) {
            $table->id();
            $table->integer('video_id');
            $table->integer('cid');
            $table->integer('progress')->default(0);
            $table->integer('mode');
            $table->integer('color');
            $table->string('content');
            $table->datetime('created_at');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('danmaku');
    }
};
