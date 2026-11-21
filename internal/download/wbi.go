package download

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

var mixinTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

func mixinKey(imgKey, subKey string) string {
	orig := imgKey + subKey
	var b strings.Builder
	for _, n := range mixinTab {
		if n >= 0 && n < len(orig) {
			b.WriteByte(orig[n])
		}
	}
	key := b.String()
	if len(key) > 32 {
		key = key[:32]
	}
	return key
}

func imageKey(rawURL string) string {
	name := rawURL
	if slash := strings.LastIndex(name, "/"); slash >= 0 {
		name = name[slash+1:]
	}
	return strings.TrimSuffix(name, ".png")
}

// signWBI 给弹幕分段参数加上 w_rid。
func signWBI(params map[string]string, imgKey, subKey string) string {
	keys := make([]string, 0, len(params))
	clean := make(map[string]string, len(params))
	for key, value := range params {
		value = strings.NewReplacer("!", "", "'", "", "(", "", ")", "", "*", "").Replace(value)
		clean[key] = value
		keys = append(keys, key)
	}
	sort.Strings(keys)
	query := url.Values{}
	for _, key := range keys {
		query.Set(key, clean[key])
	}
	encoded := query.Encode()
	sum := md5.Sum([]byte(encoded + mixinKey(imgKey, subKey)))
	return encoded + "&w_rid=" + hex.EncodeToString(sum[:])
}
