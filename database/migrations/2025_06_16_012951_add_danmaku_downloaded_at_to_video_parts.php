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
        Schema::table('video_parts', function (Blueprint $table) {
            $table->datetime('danmaku_downloaded_at')->nullable()->after('duration');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('video_parts', function (Blueprint $table) {
            $table->dropColumn('danmaku_downloaded_at');
        });
    }
};
