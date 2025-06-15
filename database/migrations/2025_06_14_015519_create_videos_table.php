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
        Schema::create('videos', function (Blueprint $table) {
            $table->id();
            $table->string('link');
            $table->string('title');
            $table->text('intro');
            $table->string('cover');
            $table->string('bvid');
            $table->datetime('pubtime');
            $table->integer('attr')->default(0);
            $table->boolean('invalid')->default(false);
            $table->boolean('frozen')->default(false);
            $table->string('cache_image');
            $table->integer('page')->default(1);
            $table->datetime('fav_time');
            $table->datetimes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('videos');
    }
};
