package covercolor

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

const (
	sampleSize = 100
	maxLevel   = 5
)

// Dominant 用八叉树从图片里取主色，返回 #rrggbb。文件打不开或没有可用颜色时 ok 为 false。
func Dominant(path string) (string, bool) {
	img, err := openImage(path)
	if err != nil {
		return "", false
	}
	sample := image.NewRGBA(image.Rect(0, 0, sampleSize, sampleSize))
	xdraw.BiLinear.Scale(sample, sample.Bounds(), img, img.Bounds(), draw.Src, nil)
	root := newNode(0)
	for y := 0; y < sampleSize; y++ {
		for x := 0; x < sampleSize; x++ {
			r, g, b, _ := sample.At(x, y).RGBA()
			rr, gg, bb := int(r>>8), int(g>>8), int(b>>8)
			if !validColor(rr, gg, bb) {
				continue
			}
			root.add(rr, gg, bb)
		}
	}
	palette := root.palette()
	if len(palette) == 0 {
		return "", false
	}
	color := palette[0]
	return fmt.Sprintf("#%02x%02x%02x", color.r, color.g, color.b), true
}

func openImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if strings.HasSuffix(strings.ToLower(path), ".webp") {
		return webp.Decode(file)
	}
	img, _, err := image.Decode(file)
	if err == nil {
		return img, nil
	}
	if _, seekErr := file.Seek(0, 0); seekErr != nil {
		return nil, err
	}
	return webp.Decode(file)
}

func validColor(r, g, b int) bool {
	brightness := float64(r*299+g*587+b*114) / 1000
	if brightness > 240 || brightness < 30 || saturation(r, g, b) < 0.2 {
		return false
	}
	return true
}

func saturation(r, g, b int) float64 {
	max, min := r, r
	if g > max {
		max = g
	}
	if b > max {
		max = b
	}
	if g < min {
		min = g
	}
	if b < min {
		min = b
	}
	if max == 0 {
		return 0
	}
	return float64(max-min) / float64(max)
}

type rgb struct {
	r, g, b, count int
}

type node struct {
	leaf     bool
	count    int
	r, g, b  int
	level    int
	children [8]*node
}

func newNode(level int) *node {
	n := &node{level: level}
	if level >= maxLevel {
		n.leaf = true
	}
	return n
}

func (n *node) add(r, g, b int) {
	if n.leaf {
		n.count++
		n.r += r
		n.g += g
		n.b += b
		return
	}
	index := colorIndex(n.level, r, g, b)
	if n.children[index] == nil {
		n.children[index] = newNode(n.level + 1)
	}
	n.children[index].add(r, g, b)
}

func colorIndex(level, r, g, b int) int {
	mask := 0x80 >> level
	index := 0
	if r&mask != 0 {
		index |= 4
	}
	if g&mask != 0 {
		index |= 2
	}
	if b&mask != 0 {
		index |= 1
	}
	return index
}

func (n *node) palette() []rgb {
	var leaves []*node
	n.leaves(&leaves)
	out := make([]rgb, 0, len(leaves))
	for _, leaf := range leaves {
		if leaf.count == 0 {
			continue
		}
		out = append(out, rgb{
			r: leaf.r / leaf.count, g: leaf.g / leaf.count, b: leaf.b / leaf.count, count: leaf.count,
		})
	}
	for i := 1; i < len(out); i++ {
		item := out[i]
		j := i
		for j > 0 && out[j-1].count < item.count {
			out[j] = out[j-1]
			j--
		}
		out[j] = item
	}
	return out
}

func (n *node) leaves(dst *[]*node) {
	if n.leaf {
		*dst = append(*dst, n)
		return
	}
	for _, child := range n.children {
		if child != nil {
			child.leaves(dst)
		}
	}
}
