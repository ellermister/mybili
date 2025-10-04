<?php
namespace App\Services;

use App\Enums\DplayerDanmaku;
use App\Models\Danmaku;

class DPlayerDanmakuService
{
    /**
     * 转换弹幕数据为dplayer格式
     * @param array<Danmaku> $danmaku 弹幕数据
     * @return array
     */
    public function convertDanmaku(array $danmaku)
    {
        // time: item[0],   时间秒
        // type: item[1],   位置需要编号
        // color: item[2],  颜色10进制
        // author: item[3],
        // text: item[4],
        return array_map(function ($item) {
            $content = $this->downgradeAdvancedDanmaku(isset($item['content']) ? $item['content'] : '');
            return [
                isset($item['progress']) ? $item['progress'] / 1000 : 0,
                isset($item['mode']) ? $this->covertMode($item['mode']) : 'right',
                isset($item['color']) ? $item['color'] : 16777215,
                '',
                $content,
            ];
        }, $danmaku);
    }

    protected function covertMode($mode)
    {
        return DplayerDanmaku::getMode($mode);
    }

    protected function downgradeAdvancedDanmaku($content)
    {
        $item = json_decode($content, true);
        if($item !== null && is_array($item) && isset($item[4])) {
            return $item[4];
        }
        return $content;
    }
}
