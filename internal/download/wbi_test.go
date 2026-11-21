package download

import (
	"strings"
	"testing"
)

func TestSignWBI(t *testing.T) {
	img := "7cd084941338484aae1ad9425b84077c"
	sub := "4932caff0ff746eab6f01bf08b70ac45"
	if len(mixinKey(img, sub)) != 32 {
		t.Fatalf("mixin = %s", mixinKey(img, sub))
	}
	query := signWBI(map[string]string{
		"type":          "1",
		"oid":           "9",
		"segment_index": "1",
		"wts":           "1700000000",
	}, img, sub)
	if !strings.Contains(query, "oid=9") || !strings.Contains(query, "&w_rid=") {
		t.Fatalf("query = %s", query)
	}
	if signWBI(map[string]string{"oid": "9", "wts": "1"}, img, sub) != signWBI(map[string]string{"wts": "1", "oid": "9"}, img, sub) {
		t.Fatal("sign should be stable")
	}
}
