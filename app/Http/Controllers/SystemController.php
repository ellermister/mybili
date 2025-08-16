<?php
namespace App\Http\Controllers;

use App\Services\SystemService;

class SystemController extends Controller
{

    public function __construct(public SystemService $systemService)
    {
    }

    public function getSystemInfo()
    {
        return response()->json($this->systemService->getSystemInfo());
    }
}
