package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/model"
	"mybili/internal/taskrun"
)

const (
	// TypeTelegram 是通知队列里的消息类型。
	TypeTelegram = "telegram"
	// CookieNoticeKey 挡住 24 小时内重复的 Cookie 过期通知。不沿用旧程序的 Redis 键。
	CookieNoticeKey = "mybili:notify:cookie-expired"

	cookieKind   = "cookie-expired"
	freezeKind   = "video-frozen"
	downloadKind = "video-downloaded"
)

// Queue 是通知入队需要的 Redis 操作。*taskrun.Tasks 实现了它。
type Queue interface {
	Push(ctx context.Context, queue string, msg taskrun.Message) error
	MarkOnce(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unmark(ctx context.Context, key string) error
}

type payload struct {
	Kind string `json:"kind"`
	ID   int64  `json:"id,omitempty"`
}

// EnqueueCookie 写入一条 Cookie 过期通知。24 小时内已经写过就不再写。
func EnqueueCookie(ctx context.Context, q Queue) error {
	if q == nil {
		return nil
	}
	ok, err := q.MarkOnce(ctx, CookieNoticeKey, 24*time.Hour)
	if err != nil || !ok {
		return err
	}
	body, err := json.Marshal(payload{Kind: cookieKind})
	if err != nil {
		_ = q.Unmark(ctx, CookieNoticeKey)
		return err
	}
	err = q.Push(ctx, taskrun.QueueNotify, taskrun.Message{Type: TypeTelegram, Payload: string(body)})
	if err != nil {
		_ = q.Unmark(ctx, CookieNoticeKey)
	}
	return err
}

// EnqueueFrozen 写入一条视频冻结通知。
func EnqueueFrozen(ctx context.Context, q Queue, videoID int64) error {
	if q == nil || videoID <= 0 {
		return nil
	}
	body, err := json.Marshal(payload{Kind: freezeKind, ID: videoID})
	if err != nil {
		return err
	}
	return q.Push(ctx, taskrun.QueueNotify, taskrun.Message{Type: TypeTelegram, Payload: string(body)})
}

// EnqueueDownloaded 写入一条分 P 下载完成通知。每次下载成功都发，不按视频去重。
func EnqueueDownloaded(ctx context.Context, q Queue, partID int64) error {
	if q == nil || partID <= 0 {
		return nil
	}
	body, err := json.Marshal(payload{Kind: downloadKind, ID: partID})
	if err != nil {
		return err
	}
	return q.Push(ctx, taskrun.QueueNotify, taskrun.Message{Type: TypeTelegram, Payload: string(body)})
}

// ClearCookieNotice 在 Cookie 更新后允许再次通知过期。
func ClearCookieNotice(ctx context.Context, q Queue) error {
	if q == nil {
		return nil
	}
	return q.Unmark(ctx, CookieNoticeKey)
}

// Register 把通知发送登记到 worker。
func Register(tasks *taskrun.Tasks, db *gorm.DB, dataDir string) {
	if tasks == nil {
		return
	}
	tasks.Handle(TypeTelegram, func(ctx context.Context, msg taskrun.Message) error {
		var body payload
		if err := json.Unmarshal([]byte(msg.Payload), &body); err != nil || body.Kind == "" {
			return taskrun.ErrCannotRetry
		}
		switch body.Kind {
		case cookieKind:
			return SendHTML(ctx, db, cookieMessage(dbtime.Now()))
		case freezeKind:
			text, err := freezeMessage(ctx, db, dataDir, body.ID)
			if err != nil {
				return err
			}
			return SendHTML(ctx, db, text)
		case downloadKind:
			text, err := downloadedMessage(ctx, db, dataDir, body.ID)
			if err != nil {
				return err
			}
			return SendHTML(ctx, db, text)
		default:
			return taskrun.ErrCannotRetry
		}
	})
}

// Test 先确认机器人身份，再发一条测试消息。不看启用开关。
func Test(ctx context.Context, token, chatID, apiURL string) (bool, error) {
	token = strings.TrimSpace(token)
	chatID = strings.TrimSpace(chatID)
	if token == "" || chatID == "" {
		return false, nil
	}
	base, err := botBase(apiURL)
	if err != nil {
		return false, nil
	}
	if _, err := call(ctx, http.MethodGet, base+"/bot"+token+"/getMe", nil); err != nil {
		return false, nil
	}
	sent, err := call(ctx, http.MethodPost, base+"/bot"+token+"/sendMessage", map[string]any{
		"chat_id": chatID,
		"text":    "Test Message - Your Telegram Bot has been connected successfully!",
	})
	if err != nil || !sent {
		return false, nil
	}
	return true, nil
}

// SendHTML 用设置里的机器人发送一条 HTML 消息。没启用时直接成功，避免反复重试。
func SendHTML(ctx context.Context, db *gorm.DB, message string) error {
	enabled, token, chatID, apiURL, err := botSettings(db)
	if err != nil {
		return err
	}
	if !enabled || token == "" || chatID == "" {
		return nil
	}
	base, err := botBase(apiURL)
	if err != nil {
		return taskrun.ErrCannotRetry
	}
	ok, err := call(ctx, http.MethodPost, base+"/bot"+token+"/sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("telegram did not accept the message")
	}
	return nil
}

func cookieMessage(now time.Time) string {
	when := now.In(dbtime.Location()).Format("2006-01-02 15:04:05")
	return `
🤖 <b>Mybili Notification</b>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ <b>Cookie Expired</b>

🍪 <b>Status</b>: Cookie has expired
⏰ <b>Time</b>: ` + when + `

🔧 <i>Please update your cookie as soon as possible to continue using the service.</i>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`
}

func freezeMessage(ctx context.Context, db *gorm.DB, dataDir string, videoID int64) (string, error) {
	if db == nil || videoID <= 0 {
		return "", taskrun.ErrCannotRetry
	}
	var video model.Video
	if err := db.WithContext(ctx).First(&video, videoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", taskrun.ErrCannotRetry
		}
		return "", err
	}
	var parts []model.VideoPart
	if err := db.WithContext(ctx).Where("video_id = ?", videoID).Find(&parts).Error; err != nil {
		return "", err
	}
	var bytes int64
	var seconds int
	for _, part := range parts {
		seconds += part.Duration
		bytes += partFileSize(dataDir, part)
	}
	link := "https://www.bilibili.com/video/av" + fmt.Sprint(video.ID)
	if strings.TrimSpace(video.Bvid) != "" {
		link = "https://www.bilibili.com/video/" + video.Bvid
	}
	return `
🤖 <b>Mybili Notification</b>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎬 <b>Video Frozen</b>

📺 <b>Title</b>: ` + html.EscapeString(video.Title) + `
⏱️ <b>Duration</b>: ` + formatDuration(seconds) + `
💾 <b>File Size</b>: ` + formatFileSize(bytes) + `
🔗 <b>Source</b>: <a href='` + link + `'>View on Bilibili</a>

🧊 <i>Video frozen successfully!</i>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, nil
}

func downloadedMessage(ctx context.Context, db *gorm.DB, dataDir string, partID int64) (string, error) {
	if db == nil || partID <= 0 {
		return "", taskrun.ErrCannotRetry
	}
	var part model.VideoPart
	if err := db.WithContext(ctx).First(&part, partID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", taskrun.ErrCannotRetry
		}
		return "", err
	}
	var video model.Video
	if err := db.WithContext(ctx).First(&video, part.VideoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", taskrun.ErrCannotRetry
		}
		return "", err
	}
	partInfo := "Single Part"
	if part.Page > 1 {
		partInfo = "Part " + fmt.Sprint(part.Page)
	}
	link := "https://www.bilibili.com/video/av" + fmt.Sprint(video.ID)
	if strings.TrimSpace(video.Bvid) != "" {
		link = "https://www.bilibili.com/video/" + video.Bvid
	}
	return `
🤖 <b>Mybili Notification</b>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🎬 <b>Video Download Completed</b>

📺 <b>Title</b>: ` + html.EscapeString(video.Title) + `
🎥 <b>Part</b>: ` + html.EscapeString(part.Part) + ` (` + partInfo + `)
⏱️ <b>Duration</b>: ` + formatDuration(part.Duration) + `
💾 <b>File Size</b>: ` + formatFileSize(partFileSize(dataDir, part)) + `
🔗 <b>Source</b>: <a href='` + link + `'>View on Bilibili</a>

✅ <i>Download completed successfully!</i>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, nil
}

func partFileSize(dataDir string, part model.VideoPart) int64 {
	path := ""
	if part.VideoDownloadPath != nil {
		path = resolveStored(dataDir, *part.VideoDownloadPath)
	}
	if path == "" || !fileSizeOf(path) {
		name := fmt.Sprintf("%d.%d.mp4", part.VideoID, part.Cid)
		path = filepath.Join(dataDir, "public", "videos", name)
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}

func fileSizeOf(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}

func resolveStored(dataDir, stored string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	if strings.HasPrefix(stored, "/storage/") {
		return filepath.Join(dataDir, "public", strings.TrimPrefix(stored, "/storage/"))
	}
	if filepath.IsAbs(stored) {
		return stored
	}
	return filepath.Join(dataDir, "public", stored)
}

func formatFileSize(n int64) string {
	if n <= 0 {
		return "0 MB"
	}
	const mb = 1024 * 1024
	const gb = mb * 1024
	if n >= gb {
		return fmt.Sprintf("%.2f GB", float64(n)/float64(gb))
	}
	return fmt.Sprintf("%.2f MB", float64(n)/float64(mb))
}

func formatDuration(sec int) string {
	if sec < 0 {
		sec = 0
	}
	return time.Unix(int64(sec), 0).UTC().Format("15:04:05")
}

func botSettings(db *gorm.DB) (bool, string, string, string, error) {
	if db == nil {
		return false, "", "", "", fmt.Errorf("database is missing")
	}
	enabled, err := readSetting(db, "telegram_bot_enabled")
	if err != nil {
		return false, "", "", "", err
	}
	token, err := readSetting(db, "telegram_bot_token")
	if err != nil {
		return false, "", "", "", err
	}
	chatID, err := readSetting(db, "telegram_chat_id")
	if err != nil {
		return false, "", "", "", err
	}
	apiURL, err := readSetting(db, "telegram_bot_api_url")
	if err != nil {
		return false, "", "", "", err
	}
	return enabled == "on", strings.TrimSpace(token), strings.TrimSpace(chatID), apiURL, nil
}

func readSetting(db *gorm.DB, name string) (string, error) {
	var row model.Setting
	err := db.Where("name = ?", name).Limit(1).Find(&row).Error
	if err != nil || row.ID == 0 || row.Value == nil {
		return "", err
	}
	var text string
	if err := json.Unmarshal([]byte(*row.Value), &text); err == nil {
		return text, nil
	}
	return *row.Value, nil
}

func botBase(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "https://api.telegram.org", nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("telegram api url is invalid")
	}
	return strings.TrimRight(raw, "/"), nil
}

func call(ctx context.Context, method, endpoint string, body map[string]any) (bool, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return false, err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return false, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("telegram http %d", resp.StatusCode)
	}
	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return false, err
	}
	return result.OK, nil
}
