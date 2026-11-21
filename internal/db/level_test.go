package db

import (
	"testing"

	"gorm.io/gorm/logger"
)

func TestSQLLogLevel(t *testing.T) {
	cases := []struct {
		in   string
		want logger.LogLevel
	}{
		{in: "", want: logger.Warn},
		{in: "warn", want: logger.Warn},
		{in: " WARN ", want: logger.Warn},
		{in: "silent", want: logger.Silent},
		{in: "error", want: logger.Error},
		{in: "info", want: logger.Info},
		{in: "nope", want: logger.Warn},
	}
	for _, tc := range cases {
		if got := SQLLogLevel(tc.in); got != tc.want {
			t.Fatalf("SQLLogLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
