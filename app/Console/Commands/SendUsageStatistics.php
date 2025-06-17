<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use App\Services\UsageStatisticsService;

class SendUsageStatistics extends Command
{
    protected $signature = 'stats:send';
    protected $description = 'Send anonymous usage statistics';

    private UsageStatisticsService $statsService;

    public function __construct(UsageStatisticsService $statsService)
    {
        parent::__construct();
        $this->statsService = $statsService;
    }

    public function handle()
    {
        $this->info('Sending usage statistics...');
        $this->statsService->sendStats();
        $this->info('Usage statistics sent successfully.');
    }
} 