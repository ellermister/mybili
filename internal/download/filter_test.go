package download

import (
	"testing"
	"time"

	"mybili/internal/dbtime"
)

func TestRulesAllow(t *testing.T) {
	rule := rules{
		multi:         false,
		nameType:      "on",
		nameContains:  "广告",
		videoLimitSec: 30 * 60,
		excludeFav:    true,
		excluded:      map[int64]bool{3: true, -8: true},
		favCutoff:     time.Date(2026, 1, 1, 0, 0, 0, 0, dbtime.Location()),
	}
	ok := Candidate{ID: 1, VideoID: 9, Page: 1, Duration: 20, Title: "正常", VideoDuration: 20, Pubtime: dbtime.DateTime{Time: time.Date(2026, 6, 1, 0, 0, 0, 0, dbtime.Location())}}
	if !rule.allow(ok, []int64{1}, nil) {
		t.Fatal("expected allow")
	}
	if rule.allow(ok, []int64{3}, nil) {
		t.Fatal("excluded favorite should block")
	}
	paged := ok
	paged.Page = 2
	if rule.allow(paged, nil, nil) {
		t.Fatal("page 2 should be blocked when multi is off")
	}
	named := ok
	named.Title = "这是广告"
	if rule.allow(named, nil, nil) {
		t.Fatal("title should be blocked")
	}
	old := ok
	old.Pubtime = dbtime.DateTime{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, dbtime.Location())}
	if rule.allow(old, nil, nil) {
		t.Fatal("old favorite time should be blocked")
	}
}
