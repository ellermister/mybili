package dbtime

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

var shanghai = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}()

func Location() *time.Location {
	return shanghai
}

func Now() time.Time {
	return time.Now().In(shanghai)
}

// DateTime 对应 SQLite 的 datetime，读写都是不带时区的本地时间。
type DateTime struct {
	time.Time
}

func (DateTime) GormDataType() string {
	return "datetime"
}

func (d DateTime) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.In(shanghai).Format("2006-01-02 15:04:05"), nil
}

func (d *DateTime) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		// SQLite 没有时区。驱动有时会把墙上时间标成 UTC，这里只取年月日时分秒，再按上海时间解释。
		d.Time = time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond(), shanghai)
		return nil
	case string:
		parsed, err := Parse(v)
		if err != nil {
			return err
		}
		d.Time = parsed
		return nil
	case []byte:
		parsed, err := Parse(string(v))
		if err != nil {
			return err
		}
		d.Time = parsed
		return nil
	case int64:
		d.Time = fromUnixStored(v)
		return nil
	case int:
		d.Time = fromUnixStored(int64(v))
		return nil
	default:
		return fmt.Errorf("cannot read %T as datetime", value)
	}
}

// fromUnixStored 读取库里直接存的 Unix 时间。秒和毫秒都接受。
// datetime 这个类型名里带 int，SQLite 会把它当成整数亲和，收藏夹的 ctime 就是这样存的。
func fromUnixStored(n int64) time.Time {
	if n == 0 {
		return time.Time{}
	}
	if n > 1_000_000_000_000 || n < -1_000_000_000_000 {
		return time.UnixMilli(n).In(shanghai)
	}
	return time.Unix(n, 0).In(shanghai)
}

func Parse(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05.000000Z",
		time.RFC3339Nano,
		time.RFC3339,
	}
	var last error
	for _, layout := range layouts {
		if strings.Contains(layout, "Z") || strings.Contains(layout, "Z07") {
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return parsed.In(shanghai), nil
			}
			last = err
			continue
		}
		parsed, err := time.ParseInLocation(layout, value, shanghai)
		if err == nil {
			return parsed, nil
		}
		last = err
	}
	return time.Time{}, last
}

func FormatLocal(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.In(shanghai).Format("2006-01-02 15:04:05")
}

func FormatMinute(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.In(shanghai).Format("2006-01-02 15:04")
}

func FormatISO(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format("2006-01-02T15:04:05.000000Z")
}

func Unix(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.Unix()
}

func FromPtr(value *DateTime) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.Time
}
