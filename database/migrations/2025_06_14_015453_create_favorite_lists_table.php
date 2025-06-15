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
        Schema::create('favorite_lists', function (Blueprint $table) {
            // 需要考虑要不要完全使用bilibili 的第三方ID，还是自己保留一个自增ID
            $table->id();
            $table->string('title');
            $table->string('cover');
            $table->datetime('ctime');
            $table->datetime('mtime');
            $table->integer('media_count')->default(0);
            $table->string('cache_image')->default('');
            $table->datetimes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('favorite_lists');
    }
};
