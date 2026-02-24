<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('download_queue', function (Blueprint $table) {
            $table->unsignedTinyInteger('retry_count')->default(0)->after('priority');
        });
    }

    public function down(): void
    {
        Schema::table('download_queue', function (Blueprint $table) {
            $table->dropColumn('retry_count');
        });
    }
};
