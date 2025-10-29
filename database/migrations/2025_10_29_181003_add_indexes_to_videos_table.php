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
        Schema::table('videos', function (Blueprint $table) {
            // 为 fav_time 添加索引，用于排序查询优化
            $table->index('fav_time', 'idx_videos_fav_time');
            
            // 为 created_at 添加索引，用于排序查询优化
            $table->index('created_at', 'idx_videos_created_at');
            
            // 为 video_downloaded_num 添加索引，用于统计查询优化
            $table->index('video_downloaded_num', 'idx_videos_downloaded_num');
            
            // 为 invalid 添加索引，用于统计查询优化
            $table->index('invalid', 'idx_videos_invalid');
            
            // 为 frozen 添加索引，用于统计查询优化
            $table->index('frozen', 'idx_videos_frozen');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('videos', function (Blueprint $table) {
            $table->dropIndex('idx_videos_fav_time');
            $table->dropIndex('idx_videos_created_at');
            $table->dropIndex('idx_videos_downloaded_num');
            $table->dropIndex('idx_videos_invalid');
            $table->dropIndex('idx_videos_frozen');
        });
    }
};

