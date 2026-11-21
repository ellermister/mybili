package syncbili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

const (
	suspendSetting = "bilibili_suspend_until"
	suspendFor     = 2 * time.Hour
	pageInterval   = time.Second
)

// SuspendedError 表示处于风控暂停，Until 之前不要再请求 Bilibili。
type SuspendedError struct {
	Until time.Time
}

// ErrThrottled 表示下载图片时被 CDN 限速。worker 会暂停 cover 队列。
var ErrThrottled = errors.New("remote throttled")

// Error 返回暂停结束时间，供调用方展示。
func (e *SuspendedError) Error() string {
	return "Bilibili risk control suspended, please try again after " + e.Until.In(dbtime.Location()).Format("2006-01-02 15:04:05")
}

// Sync 把 Bilibili 的收藏夹、视频和封面同步到本地。
// HTTP 细节在 API 里，这里负责写库和封面文件。
type Sync struct {
	DB           *gorm.DB
	API          *bilibili.Client
	Cookies      []bilibili.Cookie
	PageSize     int
	PageInterval time.Duration
	Now          func() time.Time
	DataDir      string
	CoverPush    CoverPusher
	coverJobs    *int
	// QueueVideo 在订阅发现本地还没有的稿件时写入详情任务。为空则只记关联。
	QueueVideo func(ctx context.Context, bvid string) error
	// Wake 在下载任务落成 pending 时通知 Feeder。为空则只写 tasks。
	Wake func(ctx context.Context, queue string) error
	// NotifyFrozen 在视频第一次被标成冻结时调用。为空则不通知。
	NotifyFrozen func(ctx context.Context, videoID int64)
	// AfterLinked 在收藏夹或订阅新关联到稿件后调用。为空则跳过。
	AfterLinked func(ctx context.Context, videoIDs []int64)
	// AfterUnlinked 在稿件从某个收藏夹移除后调用。folder 是收藏夹标题。为空则跳过。
	AfterUnlinked func(ctx context.Context, folder string, videoIDs []int64)
}

// Hooks 是收藏和订阅关系变化后的旁路。字段为空则不调用。
type Hooks struct {
	AfterLinked   func(ctx context.Context, videoIDs []int64)
	AfterUnlinked func(ctx context.Context, folder string, videoIDs []int64)
}

// New 用库里的 Cookie 组合同步器。pageSize 是收藏夹视频每页条数。
func New(db *gorm.DB, pageSize int) (Sync, error) {
	client, cookies, err := loadClient(db)
	if err != nil {
		return Sync{}, err
	}
	return Sync{
		DB:           db,
		API:          client,
		Cookies:      cookies,
		PageSize:     pageSize,
		PageInterval: pageInterval,
		Now:          dbtime.Now,
		coverJobs:    new(int),
	}, nil
}

// WithoutClient 只写库和封面任务，不请求 Bilibili。
func WithoutClient(db *gorm.DB) Sync {
	return Sync{DB: db, coverJobs: new(int)}
}

// now 返回这次同步使用的当前时间。测试可以替换 Now。
func (s Sync) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return dbtime.Now()
}

// suspend 把风控暂停写进设置，并返回 SuspendedError。
func (s Sync) suspend(now time.Time) error {
	until := now.Add(suspendFor)
	if err := saveSuspend(s.DB, until, now); err != nil {
		return err
	}
	return &SuspendedError{Until: until}
}

// loadClient 用设置里的 Netscape Cookie 创建 Bilibili 客户端。
func loadClient(db *gorm.DB) (*bilibili.Client, []bilibili.Cookie, error) {
	content, err := readString(db, "cookies_content")
	if err != nil {
		return nil, nil, err
	}
	cookies, err := bilibili.ParseNetscape(content)
	if err != nil {
		if errors.Is(err, bilibili.ErrEmptyCookie) {
			return nil, nil, fmt.Errorf("no cookie uploaded")
		}
		return nil, nil, err
	}
	return bilibili.NewClient(cookies), cookies, nil
}

// suspendUntil 读取暂停截止时间。还没到时 active 为 true。
func suspendUntil(db *gorm.DB, now time.Time) (time.Time, bool, error) {
	text, err := readString(db, suspendSetting)
	if err != nil || text == "" {
		return time.Time{}, false, err
	}
	var unix int64
	if err := json.Unmarshal([]byte(text), &unix); err != nil {
		return time.Time{}, false, nil
	}
	until := time.Unix(unix, 0)
	if !until.After(now) {
		return time.Time{}, false, nil
	}
	return until, true, nil
}

// saveSuspend 把暂停截止时间写成 Unix 秒，存进设置。
func saveSuspend(db *gorm.DB, until, now time.Time) error {
	encoded, err := json.Marshal(until.Unix())
	if err != nil {
		return err
	}
	return writeString(db, suspendSetting, string(encoded), now)
}

// readString 读取一条设置。值是 JSON 字符串或数字时先解开。没有记录时返回空字符串。
func readString(db *gorm.DB, name string) (string, error) {
	var row model.Setting
	err := db.Where("name = ?", name).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if row.Value == nil {
		return "", nil
	}
	var decoded string
	if err := json.Unmarshal([]byte(*row.Value), &decoded); err == nil {
		return decoded, nil
	}
	var number json.Number
	if err := json.Unmarshal([]byte(*row.Value), &number); err == nil {
		return number.String(), nil
	}
	return *row.Value, nil
}

// writeString 写入或更新一条设置。
func writeString(db *gorm.DB, name, value string, now time.Time) error {
	text := value
	stamp := dbtime.DateTime{Time: now}
	var existing model.Setting
	err := db.Where("name = ?", name).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := model.Setting{Name: name, Value: &text, CreatedAt: &stamp, UpdatedAt: &stamp}
		return db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&existing).Updates(map[string]any{"value": text, "updated_at": stamp}).Error
}

// unixDateTime 把 Unix 秒转成上海时区的时间字段。
func unixDateTime(unix int64) dbtime.DateTime {
	return dbtime.DateTime{Time: time.Unix(unix, 0).In(dbtime.Location())}
}

// waitInterval 在两次 Bilibili 请求之间等待。间隔小于等于 0 时直接返回。
func waitInterval(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
