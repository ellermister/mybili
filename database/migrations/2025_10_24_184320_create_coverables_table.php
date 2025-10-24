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
        Schema::create('coverables', function (Blueprint $table) {
            $table->id();
            $table->unsignedBigInteger('cover_id');
            $table->morphs('coverable'); // coverable_id + coverable_type
            $table->datetimes();

            // 确保同一实体只关联一个封面
            $table->unique(['coverable_id', 'coverable_type']);
            // 索引优化查询
            $table->index('cover_id');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('coverables');
    }
};
