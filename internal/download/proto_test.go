package download

import "testing"

func TestParseDanmakuSeg(t *testing.T) {
	inner := appendVarint(nil, uint64(1<<3))
	inner = appendVarint(inner, 7)
	inner = appendVarint(inner, uint64(2<<3))
	inner = appendVarint(inner, 100)
	inner = appendVarint(inner, uint64(7<<3|2))
	inner = appendVarint(inner, uint64(len("你好")))
	inner = append(inner, []byte("你好")...)
	outer := appendVarint(nil, uint64(1<<3|2))
	outer = appendVarint(outer, uint64(len(inner)))
	outer = append(outer, inner...)

	items := parseDanmakuSeg(outer)
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
	if items[0].ID != 7 || items[0].Progress != 100 || items[0].Content != "你好" || items[0].Mode != 1 {
		t.Fatalf("item = %+v", items[0])
	}
}

func appendVarint(dst []byte, value uint64) []byte {
	for value >= 0x80 {
		dst = append(dst, byte(value)|0x80)
		value >>= 7
	}
	return append(dst, byte(value))
}
