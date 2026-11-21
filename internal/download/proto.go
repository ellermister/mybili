package download

import "math"

type danmakuElem struct {
	ID       int64
	Progress int
	Mode     int
	Color    int
	Content  string
	Ctime    int64
	hasColor bool
}

// parseDanmakuSeg 解析 seg.so 里的弹幕列表。字段编号沿用 Bilibili 的 protobuf。
func parseDanmakuSeg(buf []byte) []danmakuElem {
	var out []danmakuElem
	i := 0
	for i < len(buf) {
		key, n := readVarint(buf[i:])
		if n == 0 {
			break
		}
		i += n
		field := int(key >> 3)
		wire := int(key & 7)
		if field == 1 && wire == 2 {
			length, n := readVarint(buf[i:])
			if n == 0 || length > uint64(len(buf)-i-n) {
				break
			}
			i += n
			end := i + int(length)
			item := parseDanmakuElem(buf[i:end])
			i = end
			if item.Content != "" && item.ID > 0 {
				out = append(out, item)
			}
			continue
		}
		next := skipWire(buf, i, wire)
		if next < 0 {
			break
		}
		i = next
	}
	return out
}

func parseDanmakuElem(buf []byte) danmakuElem {
	var item danmakuElem
	i := 0
	for i < len(buf) {
		key, n := readVarint(buf[i:])
		if n == 0 {
			break
		}
		i += n
		field := int(key >> 3)
		wire := int(key & 7)
		switch {
		case wire == 0 && (field == 1 || field == 2 || field == 3 || field == 5 || field == 8):
			value, n := readVarint(buf[i:])
			if n == 0 {
				return item
			}
			i += n
			switch field {
			case 1:
				if value <= math.MaxInt64 {
					item.ID = int64(value)
				}
			case 2:
				item.Progress = int(value)
			case 3:
				item.Mode = int(value)
			case 5:
				item.Color = int(value)
				item.hasColor = true
			case 8:
				if value <= math.MaxInt64 {
					item.Ctime = int64(value)
				}
			}
		case wire == 2 && field == 7:
			length, n := readVarint(buf[i:])
			if n == 0 || length > uint64(len(buf)-i-n) {
				return item
			}
			i += n
			end := i + int(length)
			item.Content = string(buf[i:end])
			i = end
		default:
			next := skipWire(buf, i, wire)
			if next < 0 {
				return item
			}
			i = next
		}
	}
	if item.Mode == 0 {
		item.Mode = 1
	}
	return item
}

func readVarint(buf []byte) (uint64, int) {
	var value uint64
	var shift uint
	for i := 0; i < len(buf) && i < 10; i++ {
		b := buf[i]
		value |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return value, i + 1
		}
		shift += 7
	}
	return 0, 0
}

func skipWire(buf []byte, i, wire int) int {
	switch wire {
	case 0:
		_, n := readVarint(buf[i:])
		if n == 0 {
			return -1
		}
		return i + n
	case 1:
		if i+8 > len(buf) {
			return -1
		}
		return i + 8
	case 5:
		if i+4 > len(buf) {
			return -1
		}
		return i + 4
	case 2:
		length, n := readVarint(buf[i:])
		if n == 0 || length > uint64(len(buf)-i-n) {
			return -1
		}
		return i + n + int(length)
	default:
		return -1
	}
}
