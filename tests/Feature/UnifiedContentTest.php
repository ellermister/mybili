<?php

namespace Tests\Feature;

use App\Contracts\VideoManagerServiceInterface;
use App\Models\FavoriteList;
use App\Models\Subscription;
use App\Models\Video;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class UnifiedContentTest extends TestCase
{
    use RefreshDatabase;

    protected function setUp(): void
    {
        parent::setUp();
        
        // 创建测试数据
        $this->createTestData();
    }

    private function createTestData()
    {
        // 创建测试收藏夹
        $favorite = FavoriteList::create([
            'id' => 1,
            'title' => '测试收藏夹',
            'intro' => '这是一个测试收藏夹',
            'cover' => 'https://example.com/cover1.jpg',
            'media_count' => 5,
        ]);

        // 创建测试订阅
        $subscription = Subscription::create([
            'id' => 1,
            'name' => '测试UP主',
            'description' => '这是一个测试UP主',
            'cover' => 'https://example.com/cover2.jpg',
            'type' => 'up',
            'mid' => 12345,
            'total' => 10,
            'status' => 'active',
        ]);

        // 创建测试视频
        $video1 = Video::create([
            'id' => 1,
            'title' => '测试视频1',
            'bvid' => 'BV1234567890',
        ]);

        $video2 = Video::create([
            'id' => 2,
            'title' => '测试视频2',
            'bvid' => 'BV0987654321',
        ]);

        // 关联视频到收藏夹
        $favorite->videos()->attach([1, 2]);

        // 关联视频到订阅
        $subscription->videos()->attach([1, 2]);
    }

    public function test_get_unified_content_list()
    {
        $service = app(VideoManagerServiceInterface::class);
        $result = $service->getUnifiedContentList();

        $this->assertIsArray($result);
        $this->assertCount(2, $result);

        // 检查收藏夹
        $favorite = collect($result)->firstWhere('id', 1);
        $this->assertNotNull($favorite);
        $this->assertEquals('favorite', $favorite['type']);
        $this->assertEquals('测试收藏夹', $favorite['name']);
        
        // 验证时间字段为时间戳格式
        $this->assertIsInt($favorite['created_at']);
        $this->assertIsInt($favorite['updated_at']);

        // 检查订阅
        $subscription = collect($result)->firstWhere('id', -1);
        $this->assertNotNull($subscription);
        $this->assertEquals('subscription', $subscription['type']);
        $this->assertEquals('测试UP主', $subscription['name']);
        
        // 验证时间字段为时间戳格式
        $this->assertIsInt($subscription['created_at']);
        $this->assertIsInt($subscription['updated_at']);
    }

    public function test_get_unified_content_detail_favorite()
    {
        $service = app(VideoManagerServiceInterface::class);
        $result = $service->getUnifiedContentDetail(1);

        $this->assertNotNull($result);
        $this->assertEquals(1, $result->id);
        $this->assertEquals('favorite', $result->type);
        $this->assertEquals('测试收藏夹', $result->name);
        $this->assertCount(2, $result->videos);
        
        // 验证时间字段为时间戳格式
        $this->assertIsInt($result->created_at);
        $this->assertIsInt($result->updated_at);
    }

    public function test_get_unified_content_detail_subscription()
    {
        $service = app(VideoManagerServiceInterface::class);
        $result = $service->getUnifiedContentDetail(-1);

        $this->assertNotNull($result);
        $this->assertEquals(-1, $result->id);
        $this->assertEquals('subscription', $result->type);
        $this->assertEquals('测试UP主', $result->name);
        $this->assertCount(2, $result->videos);
        
        // 验证时间字段为时间戳格式
        $this->assertIsInt($result->created_at);
        $this->assertIsInt($result->updated_at);
        $this->assertIsInt($result->last_check_at);
    }

    public function test_is_subscription()
    {
        $service = app(VideoManagerServiceInterface::class);
        
        $this->assertTrue($service->isSubscription(-1));
        $this->assertTrue($service->isSubscription(-100));
        $this->assertFalse($service->isSubscription(1));
        $this->assertFalse($service->isSubscription(0));
    }

    public function test_is_favorite()
    {
        $service = app(VideoManagerServiceInterface::class);
        
        $this->assertTrue($service->isFavorite(1));
        $this->assertTrue($service->isFavorite(100));
        $this->assertFalse($service->isFavorite(-1));
        $this->assertFalse($service->isFavorite(0));
    }
}
