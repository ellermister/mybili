<?php
namespace App\Services;

/**
 * 弹幕转换服务
 * 将 Bilibili 格式的弹幕转换为播放器格式
 */
class DanmakuConverterService
{
    /**
     * 转换弹幕数据为播放器格式
     * 
     * @param array $danmakuList 弹幕数据数组
     * @return array
     */
    public function convert(array $danmakuList): array
    {
        return array_map(function ($danmaku) {
            return $this->convertSingle($danmaku);
        }, $danmakuList);
    }

    /**
     * 转换单条弹幕
     * 
     * @param array $danmaku 弹幕数据
     * @return array
     */
    protected function convertSingle(array $danmaku): array
    {
        // 降级高级弹幕，获取纯文本内容
        $text = $this->downgradeAdvancedDanmaku($danmaku['content'] ?? '');
        
        return [
            'text'   => $text,
            'time'   => $this->convertTime($danmaku['progress'] ?? 0),
            'mode'   => $this->convertMode($danmaku['mode'] ?? 1),
            'color'  => $this->convertColor($danmaku['color'] ?? 16777215),
            'border' => false,
            'style'  => new \stdClass(), // 空对象
        ];
    }

    /**
     * 将进度时间从毫秒转换为秒
     * 
     * @param int $progress 进度（毫秒）
     * @return float
     */
    protected function convertTime(int $progress): float
    {
        return round($progress / 1000, 3);
    }

    /**
     * 转换弹幕模式
     * Bilibili 模式 -> 播放器模式
     * 
     * Bilibili:
     * 1-3: 滚动弹幕
     * 4: 底部弹幕
     * 5: 顶部弹幕
     * 6-9: 高级弹幕（降级为滚动）
     * 
     * 播放器:
     * 0: 滚动
     * 1: 顶部
     * 2: 底部
     * 
     * @param int $mode Bilibili 弹幕模式
     * @return int
     */
    protected function convertMode(int $mode): int
    {
        return match ($mode) {
            5 => 1,       // 顶部
            4 => 2,       // 底部
            default => 0, // 1-3 滚动, 6-9 高级弹幕降级为滚动
        };
    }

    /**
     * 将十进制颜色转换为十六进制颜色字符串
     * 
     * @param int $color 十进制颜色值
     * @return string 十六进制颜色字符串 (如 #FFFFFF)
     */
    protected function convertColor(int $color): string
    {
        return '#' . strtoupper(str_pad(dechex($color), 6, '0', STR_PAD_LEFT));
    }

    /**
     * 降级高级弹幕
     * 高级弹幕为 JSON 数组格式，索引 4 为文本内容
     * 
     * @param string $content 弹幕内容
     * @return string 纯文本内容
     */
    protected function downgradeAdvancedDanmaku(string $content): string
    {
        // 尝试解析 JSON
        $decoded = json_decode($content, true);
        
        // 如果是有效的 JSON 数组且包含索引 4（文本内容）
        if (is_array($decoded) && isset($decoded[4])) {
            return (string) $decoded[4];
        }
        
        // 否则返回原始内容
        return $content;
    }
}

