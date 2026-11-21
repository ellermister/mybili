package dbtime

import (
	"testing"
	"time"
)

func TestScanDriverUTCWallClock(t *testing.T) {
	var parsed DateTime
	driverTime := time.Date(2025, 8, 17, 6, 0, 0, 0, time.UTC)
	if err := parsed.Scan(driverTime); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2025, 8, 17, 6, 0, 0, 0, Location())
	if got := FormatISO(parsed.Time); got != "2025-08-16T22:00:00.000000Z" {
		t.Fatalf("iso = %v", got)
	}
	if parsed.Unix() != want.Unix() {
		t.Fatalf("unix = %d, want %d", parsed.Unix(), want.Unix())
	}
}

func TestScanUnixSeconds(t *testing.T) {
	var parsed DateTime
	if err := parsed.Scan(int64(1499900654)); err != nil {
		t.Fatal(err)
	}
	if parsed.Unix() != int64(1499900654) {
		t.Fatalf("unix = %d", parsed.Unix())
	}
	if got := FormatLocal(parsed.Time); got != "2017-07-13 07:04:14" {
		t.Fatalf("local = %v", got)
	}
}

func TestParseLocalDatetime(t *testing.T) {
	parsed, err := Parse("2025-08-17 04:00:00")
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatLocal(parsed); got != "2025-08-17 04:00:00" {
		t.Fatalf("local = %v", got)
	}
	if got := FormatISO(parsed); got != "2025-08-16T20:00:00.000000Z" {
		t.Fatalf("iso = %v", got)
	}
}
