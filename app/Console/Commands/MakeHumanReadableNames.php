<?php
namespace App\Console\Commands;

use App\Services\HumanReadableNameService;
use Illuminate\Console\Command;

class MakeHumanReadableNames extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:make-human-readable-names';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = '生成可读文件名和目录结构';

    protected HumanReadableNameService $humanReadableNameService;

    public function __construct(HumanReadableNameService $humanReadableNameService)
    {
        parent::__construct();
        $this->humanReadableNameService = $humanReadableNameService;
    }

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $this->info('开始生成可读文件名和目录结构...');
        
        $this->humanReadableNameService->generateHumanReadableNames();
        $this->info('可读文件名和目录结构生成完成！');
        
        return 0;
    }
}
