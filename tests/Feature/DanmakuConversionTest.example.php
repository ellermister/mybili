<?php

/**
 * 弹幕转换测试示例
 * 
 * 这是一个示例文件，展示如何测试弹幕转换功能
 * 如需使用，请重命名为 DanmakuConversionTest.php
 */

namespace Tests\Feature;

use App\Services\DanmakuConverterService;
use Tests\TestCase;

class DanmakuConversionTest extends TestCase
{
    protected DanmakuConverterService $converter;

    protected function setUp(): void
    {
        parent::setUp();
        $this->converter = new DanmakuConverterService();
    }

    /**
     * 测试滚动弹幕转换 (mode 1)
     */
    public function test_convert_rolling_danmaku()
    {
        $input = [
            [
                'id' => 1,
                'progress' => 45811,
                'mode' => 1,
                'color' => 16777215,
                'content' => '这是一条滚动弹幕',
            ]
        ];

        $result = $this->converter->convert($input);

        $this->assertEquals([
            [
                'text' => '这是一条滚动弹幕',
                'time' => 45.811,
                'mode' => 0, // 滚动
                'color' => '#FFFFFF',
                'border' => false,
                'style' => new \stdClass(),
            ]
        ], $result);
    }

    /**
     * 测试顶部弹幕转换 (mode 5)
     */
    public function test_convert_top_danmaku()
    {
        $input = [
            [
                'progress' => 10000,
                'mode' => 5,
                'color' => 16711680, // 红色
                'content' => '顶部弹幕',
            ]
        ];

        $result = $this->converter->convert($input);

        $this->assertEquals('#FF0000', $result[0]['color']);
        $this->assertEquals(1, $result[0]['mode']); // 顶部
    }

    /**
     * 测试底部弹幕转换 (mode 4)
     */
    public function test_convert_bottom_danmaku()
    {
        $input = [
            [
                'progress' => 20000,
                'mode' => 4,
                'color' => 65280, // 绿色
                'content' => '底部弹幕',
            ]
        ];

        $result = $this->converter->convert($input);

        $this->assertEquals('#00FF00', $result[0]['color']);
        $this->assertEquals(2, $result[0]['mode']); // 底部
    }

    /**
     * 测试高级弹幕降级
     */
    public function test_downgrade_advanced_danmaku()
    {
        $input = [
            [
                'progress' => 30000,
                'mode' => 7, // 高级弹幕
                'color' => 16777215,
                'content' => '[7, 25, 16777215, 1234567890, "这是高级弹幕", {}]',
            ]
        ];

        $result = $this->converter->convert($input);

        $this->assertEquals('这是高级弹幕', $result[0]['text']);
        $this->assertEquals(0, $result[0]['mode']); // 降级为滚动
    }

    /**
     * 测试颜色转换
     */
    public function test_color_conversion()
    {
        $testCases = [
            16777215 => '#FFFFFF', // 白色
            0 => '#000000',        // 黑色
            16711680 => '#FF0000', // 红色
            65280 => '#00FF00',    // 绿色
            255 => '#0000FF',      // 蓝色
            16776960 => '#FFFF00', // 黄色
        ];

        foreach ($testCases as $decimal => $hex) {
            $input = [[
                'progress' => 0,
                'mode' => 1,
                'color' => $decimal,
                'content' => 'test',
            ]];

            $result = $this->converter->convert($input);
            $this->assertEquals($hex, $result[0]['color'], "Failed for color $decimal");
        }
    }

    /**
     * 测试时间转换精度
     */
    public function test_time_precision()
    {
        $testCases = [
            45811 => 45.811,
            10000 => 10.0,
            1234 => 1.234,
            0 => 0.0,
        ];

        foreach ($testCases as $progress => $expectedTime) {
            $input = [[
                'progress' => $progress,
                'mode' => 1,
                'color' => 16777215,
                'content' => 'test',
            ]];

            $result = $this->converter->convert($input);
            $this->assertEquals($expectedTime, $result[0]['time']);
        }
    }

    /**
     * 测试空数组
     */
    public function test_convert_empty_array()
    {
        $result = $this->converter->convert([]);
        $this->assertEmpty($result);
    }

    /**
     * 测试批量转换
     */
    public function test_batch_conversion()
    {
        $input = [
            [
                'progress' => 1000,
                'mode' => 1,
                'color' => 16777215,
                'content' => '第一条',
            ],
            [
                'progress' => 2000,
                'mode' => 5,
                'color' => 16711680,
                'content' => '第二条',
            ],
            [
                'progress' => 3000,
                'mode' => 4,
                'color' => 65280,
                'content' => '第三条',
            ],
        ];

        $result = $this->converter->convert($input);

        $this->assertCount(3, $result);
        $this->assertEquals('第一条', $result[0]['text']);
        $this->assertEquals('第二条', $result[1]['text']);
        $this->assertEquals('第三条', $result[2]['text']);
    }

    /**
     * 测试 API 端点
     */
    public function test_danmaku_api_endpoint()
    {
        // 这个测试需要有实际的数据库数据
        // 可以使用工厂或种子数据来创建测试数据
        
        $response = $this->getJson('/api/danmaku?id=12345');

        $response->assertStatus(200)
                 ->assertJsonStructure([
                     'code',
                     'data' => [
                         '*' => [
                             'text',
                             'time',
                             'mode',
                             'color',
                             'border',
                             'style',
                         ]
                     ]
                 ]);
    }

    /**
     * 测试缺少 CID 参数
     */
    public function test_danmaku_api_missing_cid()
    {
        $response = $this->getJson('/api/danmaku');

        $response->assertStatus(200)
                 ->assertJson([
                     'code' => 1,
                     'message' => 'CID 参数不能为空',
                     'data' => [],
                 ]);
    }
}

