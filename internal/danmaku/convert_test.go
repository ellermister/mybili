package danmaku

import "testing"

func TestConvertTimeAndMode(t *testing.T) {
	color := 16711680
	items := Convert([]Row{{
		Content:  `["","","","","你好"]`,
		Progress: 1500,
		Mode:     5,
		Color:    &color,
	}})
	if len(items) != 1 {
		t.Fatal(len(items))
	}
	if items[0].Text != "你好" || items[0].Time != 1.5 || items[0].Mode != 1 || items[0].Color != "#FF0000" {
		t.Fatalf("%+v", items[0])
	}
}
