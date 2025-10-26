<?php

namespace App\Services;

use Exception;
use Illuminate\Support\Facades\Log;

/**
 * 颜色提取服务
 * 支持八叉树法和K-Means聚类法
 */
class ColorExtractionService
{
    const METHOD_OCTREE = 'octree';
    const METHOD_KMEANS = 'kmeans';
    
    private string $method;
    private int $sampleSize;
    private int $kmeansK;
    
    /**
     * @param string $method 提取方法: 'octree' 或 'kmeans'
     * @param int $sampleSize 采样大小，默认100x100
     * @param int $kmeansK K-Means算法的K值，默认5
     */
    public function __construct(string $method = self::METHOD_OCTREE, int $sampleSize = 100, int $kmeansK = 5)
    {
        $this->method = $method;
        $this->sampleSize = $sampleSize;
        $this->kmeansK = $kmeansK;
    }
    
    /**
     * 提取图片主色调
     * 
     * @param string $imagePath 图片路径
     * @return array|null ['r' => int, 'g' => int, 'b' => int, 'hex' => string] 或 null
     */
    public function extractDominantColor(string $imagePath): ?array
    {
        if (!file_exists($imagePath)) {
            Log::warning("颜色提取失败：图片不存在", ['path' => $imagePath]);
            return null;
        }
        
        try {
            $result = match($this->method) {
                self::METHOD_OCTREE => $this->extractByOctree($imagePath),
                self::METHOD_KMEANS => $this->extractByKMeans($imagePath),
                default => null,
            };
            
            return $result;
        } catch (Exception $e) {
            Log::error("颜色提取异常", [
                'path' => $imagePath,
                'method' => $this->method,
                'error' => $e->getMessage()
            ]);
            return null;
        }
    }
    
    /**
     * 八叉树颜色提取法
     */
    private function extractByOctree(string $imagePath): ?array
    {
        $img = $this->loadImage($imagePath);
        if (!$img) {
            return null;
        }
        
        [$width, $height] = [imagesx($img), imagesy($img)];
        
        // 缩放图片
        $sampleImg = imagecreatetruecolor($this->sampleSize, $this->sampleSize);
        imagecopyresampled($sampleImg, $img, 0, 0, 0, 0, $this->sampleSize, $this->sampleSize, $width, $height);
        
        // 构建八叉树
        $octree = new OctreeColorQuantizer();
        
        for ($x = 0; $x < $this->sampleSize; $x++) {
            for ($y = 0; $y < $this->sampleSize; $y++) {
                $rgb = imagecolorat($sampleImg, $x, $y);
                $r = ($rgb >> 16) & 0xFF;
                $g = ($rgb >> 8) & 0xFF;
                $b = $rgb & 0xFF;
                
                // 过滤极端颜色
                if (!$this->isValidColor($r, $g, $b)) {
                    continue;
                }
                
                $octree->addColor($r, $g, $b);
            }
        }
        
        $palette = $octree->makePalette();
        
        imagedestroy($img);
        imagedestroy($sampleImg);
        
        if (!empty($palette)) {
            $color = $palette[0];
            return $this->formatColor($color['r'], $color['g'], $color['b']);
        }
        
        return null;
    }
    
    /**
     * K-Means 聚类颜色提取法
     */
    private function extractByKMeans(string $imagePath): ?array
    {
        $img = $this->loadImage($imagePath);
        if (!$img) {
            return null;
        }
        
        [$width, $height] = [imagesx($img), imagesy($img)];
        
        // 缩放图片
        $sampleImg = imagecreatetruecolor($this->sampleSize, $this->sampleSize);
        imagecopyresampled($sampleImg, $img, 0, 0, 0, 0, $this->sampleSize, $this->sampleSize, $width, $height);
        
        // 收集颜色点
        $points = [];
        for ($x = 0; $x < $this->sampleSize; $x++) {
            for ($y = 0; $y < $this->sampleSize; $y++) {
                $rgb = imagecolorat($sampleImg, $x, $y);
                $r = ($rgb >> 16) & 0xFF;
                $g = ($rgb >> 8) & 0xFF;
                $b = $rgb & 0xFF;
                
                // 过滤极端颜色
                if (!$this->isValidColor($r, $g, $b)) {
                    continue;
                }
                
                $saturation = $this->calculateSaturation($r, $g, $b);
                $points[] = ['r' => $r, 'g' => $g, 'b' => $b, 'saturation' => $saturation];
            }
        }
        
        imagedestroy($img);
        imagedestroy($sampleImg);
        
        if (empty($points)) {
            return null;
        }
        
        // K-Means 聚类
        $centers = $this->kMeansClustering($points, $this->kmeansK);
        
        if (!empty($centers)) {
            $dominant = $centers[0];
            return $this->formatColor($dominant['r'], $dominant['g'], $dominant['b']);
        }
        
        return null;
    }
    
    /**
     * K-Means 聚类算法
     */
    private function kMeansClustering(array $points, int $k): array
    {
        mt_srand(42); // 固定种子
        
        $pointCount = count($points);
        $step = max(1, floor($pointCount / $k));
        
        // 初始化中心点
        $centers = [];
        for ($i = 0; $i < $k && $i * $step < $pointCount; $i++) {
            $centers[] = $points[$i * $step];
        }
        
        // K-Means 迭代
        $maxIterations = 10;
        for ($iter = 0; $iter < $maxIterations; $iter++) {
            // 分配点到最近的中心
            $clusters = array_fill(0, $k, []);
            
            foreach ($points as $point) {
                $minDist = PHP_FLOAT_MAX;
                $minIndex = 0;
                
                foreach ($centers as $ci => $center) {
                    $dist = abs($point['r'] - $center['r']) + 
                           abs($point['g'] - $center['g']) + 
                           abs($point['b'] - $center['b']);
                    
                    if ($dist < $minDist) {
                        $minDist = $dist;
                        $minIndex = $ci;
                    }
                }
                
                $clusters[$minIndex][] = $point;
            }
            
            // 更新中心点
            $newCenters = [];
            foreach ($clusters as $cluster) {
                if (empty($cluster)) {
                    $newCenters[] = $centers[count($newCenters)] ?? $points[0];
                    continue;
                }
                
                $sumR = $sumG = $sumB = $sumSat = 0;
                foreach ($cluster as $point) {
                    $sumR += $point['r'];
                    $sumG += $point['g'];
                    $sumB += $point['b'];
                    $sumSat += $point['saturation'];
                }
                
                $count = count($cluster);
                $newCenters[] = [
                    'r' => (int)($sumR / $count),
                    'g' => (int)($sumG / $count),
                    'b' => (int)($sumB / $count),
                    'saturation' => $sumSat / $count,
                    'count' => $count
                ];
            }
            
            $centers = $newCenters;
        }
        
        // 按像素数量排序
        usort($centers, function($a, $b) {
            return ($b['count'] ?? 0) <=> ($a['count'] ?? 0);
        });
        
        return $centers;
    }
    
    /**
     * 加载图片
     */
    private function loadImage(string $path)
    {
        $imageInfo = @getimagesize($path);
        if (!$imageInfo) {
            return false;
        }
        
        $mimeType = $imageInfo['mime'];
        
        return match($mimeType) {
            'image/jpeg' => @imagecreatefromjpeg($path),
            'image/png' => @imagecreatefrompng($path),
            'image/gif' => @imagecreatefromgif($path),
            'image/webp' => @imagecreatefromwebp($path),
            default => false,
        };
    }
    
    /**
     * 验证颜色是否有效（过滤白色、黑色、灰色）
     */
    private function isValidColor(int $r, int $g, int $b): bool
    {
        // 计算亮度
        $brightness = ($r * 299 + $g * 587 + $b * 114) / 1000;
        
        // 计算饱和度
        $saturation = $this->calculateSaturation($r, $g, $b);
        
        // 过滤条件
        return !($brightness > 240 || $brightness < 30 || $saturation < 0.2);
    }
    
    /**
     * 计算饱和度
     */
    private function calculateSaturation(int $r, int $g, int $b): float
    {
        $max = max($r, $g, $b);
        $min = min($r, $g, $b);
        return $max == 0 ? 0 : ($max - $min) / $max;
    }
    
    /**
     * 格式化颜色输出
     */
    private function formatColor(int $r, int $g, int $b): array
    {
        return [
            'r' => $r,
            'g' => $g,
            'b' => $b,
            'hex' => sprintf('#%02x%02x%02x', $r, $g, $b)
        ];
    }
}

/**
 * 八叉树节点类
 */
class OctreeNode
{
    public bool $isLeaf = false;
    public int $pixelCount = 0;
    public int $red = 0;
    public int $green = 0;
    public int $blue = 0;
    public array $children = [];
    public int $level = 0;
    private static int $maxLevel = 5;
    
    public function __construct(int $level)
    {
        $this->level = $level;
        if ($level >= self::$maxLevel) {
            $this->isLeaf = true;
        }
    }
    
    public function addColor(int $r, int $g, int $b): void
    {
        if ($this->isLeaf) {
            $this->pixelCount++;
            $this->red += $r;
            $this->green += $g;
            $this->blue += $b;
        } else {
            $index = $this->getColorIndex($r, $g, $b);
            
            if (!isset($this->children[$index])) {
                $this->children[$index] = new OctreeNode($this->level + 1);
            }
            
            $this->children[$index]->addColor($r, $g, $b);
        }
    }
    
    private function getColorIndex(int $r, int $g, int $b): int
    {
        $index = 0;
        $mask = 0b10000000 >> $this->level;
        
        if ($r & $mask) $index |= 4;
        if ($g & $mask) $index |= 2;
        if ($b & $mask) $index |= 1;
        
        return $index;
    }
    
    public function getLeafNodes(array &$leafNodes = []): array
    {
        if ($this->isLeaf) {
            $leafNodes[] = $this;
        } else {
            foreach ($this->children as $child) {
                if ($child !== null) {
                    $child->getLeafNodes($leafNodes);
                }
            }
        }
        return $leafNodes;
    }
}

/**
 * 八叉树颜色量化器
 */
class OctreeColorQuantizer
{
    private OctreeNode $root;
    
    public function __construct()
    {
        $this->root = new OctreeNode(0);
    }
    
    public function addColor(int $r, int $g, int $b): void
    {
        $this->root->addColor($r, $g, $b);
    }
    
    public function makePalette(): array
    {
        $leafNodes = [];
        $this->root->getLeafNodes($leafNodes);
        
        $palette = [];
        foreach ($leafNodes as $node) {
            if ($node->pixelCount > 0) {
                $palette[] = [
                    'r' => (int)($node->red / $node->pixelCount),
                    'g' => (int)($node->green / $node->pixelCount),
                    'b' => (int)($node->blue / $node->pixelCount),
                    'count' => $node->pixelCount
                ];
            }
        }
        
        // 按像素数量排序
        usort($palette, function($a, $b) {
            return $b['count'] <=> $a['count'];
        });
        
        return $palette;
    }
}

