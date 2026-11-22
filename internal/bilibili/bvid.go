package bilibili

import "strings"

const (
	bvAlphabet = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
	bvXOR      = 23442827791579
	bvMask     = 2251799813685247 // 2^51 - 1
)

// BvToAv 把 BV 号解成 aid。格式不对或结果不是正数时返回 false。
// 现行算法使用 BV1 后面的 9 位，先交换下标 3↔9、4↔7，再按 58 进制解码。
func BvToAv(bvid string) (int64, bool) {
	bvid = strings.TrimSpace(bvid)
	if len(bvid) != 12 || !strings.HasPrefix(bvid, "BV1") {
		return 0, false
	}
	buf := []byte(bvid)
	buf[3], buf[9] = buf[9], buf[3]
	buf[4], buf[7] = buf[7], buf[4]
	var packed int64
	for i := 3; i < 12; i++ {
		digit := strings.IndexByte(bvAlphabet, buf[i])
		if digit < 0 {
			return 0, false
		}
		packed = packed*58 + int64(digit)
	}
	aid := (packed & bvMask) ^ bvXOR
	if aid <= 0 {
		return 0, false
	}
	return aid, true
}
