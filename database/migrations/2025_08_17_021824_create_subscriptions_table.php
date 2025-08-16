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
        Schema::create('subscriptions', function (Blueprint $table) {
            $table->id();
            $table->string('type')->comment('1: up, 2: seasons');
            $table->string('mid');
            $table->string('season_id')->nullable();
            $table->string('name')->default('');
            $table->string('url')->default('');
            $table->string('cover')->default('');
            $table->string('description')->default('');
            $table->string('total')->default(0);
            $table->tinyInteger('status')->default(1)->comment('1: active, 0: inactive');
            $table->datetime('last_check_at')->nullable();
            $table->datetimes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('subscriptions');
    }
};
