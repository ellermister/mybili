package bilibili

import "testing"

func TestBvToAv(t *testing.T) {
	cases := []struct {
		bvid string
		aid  int64
	}{
		{bvid: "BV17x411w7KC", aid: 170001},
		{bvid: "BV1134y1g7Vq", aid: 832246077},
		{bvid: "BV19x411F7m7", aid: 457094},
		{bvid: "BV1Ax411F7Pe", aid: 287966},
		{bvid: "BV1L9Uoa9EUx", aid: 111298867365120},
		{bvid: "BV1cTWU6KE9R", aid: 117599123802618},
		{bvid: "BV1zYnQ6PEew", aid: 117504919735279},
	}
	for _, item := range cases {
		got, ok := BvToAv(item.bvid)
		if !ok || got != item.aid {
			t.Fatalf("%s = %d ok=%v", item.bvid, got, ok)
		}
	}
	if _, ok := BvToAv("BV12"); ok {
		t.Fatal("short bvid decoded")
	}
	if _, ok := BvToAv(""); ok {
		t.Fatal("empty bvid decoded")
	}
}
