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
        Schema::create('video_parts', function (Blueprint $table) {
            $table->id();
            $table->integer('video_id');
            $table->integer('cid');
            $table->integer('page')->default(1);
            $table->string('from');
            $table->string('part');
            $table->integer('duration');
            $table->string('vid');
            $table->string('weblink');
            $table->integer('width')->default(0);
            $table->integer('height')->default(0);
            $table->integer('rotate')->default(0);
            $table->string('first_frame')->default('');
            $table->datetimes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('video_parts');
    }
};
