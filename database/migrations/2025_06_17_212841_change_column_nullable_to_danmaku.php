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
        Schema::table('danmaku', function (Blueprint $table) {
            $table->integer('color')->nullable()->change();
            $table->integer('mode')->default(1)->change();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('danmaku', function (Blueprint $table) {
            $table->integer('color')->nullable(false)->change();
            // 移除默认值
            $table->integer('mode')->default(null)->change();
        });
    }
};
