<?php

namespace App\Console\Commands;

use App\Services\ScheduledRateLimiterService;
use App\Services\RateLimitConfig;
use Illuminate\Console\Command;

class ScheduledRateLimitStatusCommand extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'scheduled-rate-limit:status 
                            {--reset= : Reset specific scheduled rate limit key}
                            {--list : List all scheduled rate limit keys}
                            {--schedule= : Show schedule for specific key}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = '查看预安排执行时间的频率限制状态';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $scheduledRateLimiter = app(ScheduledRateLimiterService::class);

        if ($this->option('reset')) {
            $key = $this->option('reset');
            if ($scheduledRateLimiter->resetSchedule($key)) {
                $this->info("已重置预安排频率限制键: {$key}");
            } else {
                $this->error("重置预安排频率限制键失败: {$key}");
            }
            return;
        }

        if ($this->option('schedule')) {
            $key = $this->option('schedule');
            $this->showScheduleForKey($scheduledRateLimiter, $key);
            return;
        }

        if ($this->option('list')) {
            $this->listAllScheduledRateLimits($scheduledRateLimiter);
            return;
        }

        $this->showScheduledRateLimitStatus($scheduledRateLimiter);
    }

    /**
     * 显示预安排频率限制状态
     */
    private function showScheduledRateLimitStatus(ScheduledRateLimiterService $scheduledRateLimiter): void
    {
        $this->info('=== 预安排执行时间频率限制状态 ===');
        
        $jobConfig = RateLimitConfig::getJobConfig();
        foreach ($jobConfig as $jobType => $config) {
            $stats = $scheduledRateLimiter->getScheduleStats($jobType, $config['max_requests'], $config['window_seconds']);
            
            $status = $stats['can_execute_now'] ? '可立即执行' : '已安排延迟';
            $color = $stats['can_execute_now'] ? 'green' : 'yellow';
            
            $this->line("<fg={$color}>{$jobType}: {$stats['scheduled_count']}/{$stats['max_requests']} ({$status})</>");
            $this->line("  利用率: {$stats['utilization_percent']}% | 剩余槽位: {$stats['remaining_slots']}");
            
            if (!$stats['can_execute_now']) {
                $delayMinutes = round($stats['delay_seconds'] / 60, 1);
                $this->line("  下次执行时间: " . date('Y-m-d H:i:s', $stats['next_available_time']) . " (延迟 {$delayMinutes} 分钟)");
            }
            $this->line('');
        }
    }

    /**
     * 显示特定键的执行计划
     */
    private function showScheduleForKey(ScheduledRateLimiterService $scheduledRateLimiter, string $key): void
    {
        $this->info("=== {$key} 的执行计划 ===");
        
        $jobConfig = RateLimitConfig::getJobConfig();
        if (!isset($jobConfig[$key])) {
            $this->error("未找到键 {$key} 的配置");
            return;
        }
        
        $config = $jobConfig[$key];
        $stats = $scheduledRateLimiter->getScheduleStats($key, $config['max_requests'], $config['window_seconds']);
        
        $this->line("配置信息:");
        $this->line("  最大请求数: {$stats['max_requests']}");
        $this->line("  时间窗口: {$stats['window_seconds']} 秒");
        $this->line("  当前计划数: {$stats['scheduled_count']}");
        $this->line("  利用率: {$stats['utilization_percent']}%");
        $this->line("  剩余槽位: {$stats['remaining_slots']}");
        $this->line("");
        
        $this->line("执行计划:");
        $scheduledTimes = $scheduledRateLimiter->getScheduledTimes($key, $config['window_seconds']);
        
        if (empty($scheduledTimes)) {
            $this->line("  暂无执行计划");
        } else {
            foreach ($scheduledTimes as $index => $time) {
                $this->line("  " . ($index + 1) . ". {$time['formatted']}");
            }
        }
        
        $this->line("");
        $this->line("状态信息:");
        if ($stats['can_execute_now']) {
            $this->line("  ✅ 可以立即执行");
        } else {
            $delayMinutes = round($stats['delay_seconds'] / 60, 1);
            $this->line("  ⏰ 需要延迟 {$delayMinutes} 分钟");
            $this->line("  📅 下次执行时间: " . date('Y-m-d H:i:s', $stats['next_available_time']));
        }
    }

    /**
     * 列出所有预安排频率限制键
     */
    private function listAllScheduledRateLimits(ScheduledRateLimiterService $scheduledRateLimiter): void
    {
        $this->info('=== 所有预安排频率限制键 ===');
        
        $jobConfig = RateLimitConfig::getJobConfig();
        foreach ($jobConfig as $jobType => $config) {
            $this->line("Job: {$jobType}");
            $this->line("  配置: {$config['max_requests']} 个请求 / {$config['window_seconds']} 秒");
        }

        $this->info('');
        $this->info('使用 --reset=key 来重置特定的预安排频率限制键');
        $this->info('使用 --schedule=key 来查看特定键的执行计划');
    }
} 