package usage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/model"
	"mybili/internal/version"
)

const (
	// signKeyID
	signKeyID = "v1"
	// signSecret
	signSecret = "7H70hrv1ww3fFHkonpKzUpc02wbRvjGm"
	// sessionURL
	sessionURL = "https://mjj-go.mjj.ee/api/public/software/mybili/session"

	heartbeatEvery = 5 * time.Minute
	installSetting = "usage_install_id"
)

var clientIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,64}$`)

// Reporter 在 server 进程里上报一次会话。session_id 只活在这个进程里。
type Reporter struct {
	installID string
	sessionID string
	started   time.Time
	client    *http.Client
	url       string
	secret    string
	version   string
	platform  string

	mu sync.Mutex
}

// Start 准备一次会话。安装编号会写入设置表，会话编号不落库。
// 密钥还没填时返回 nil，调用方可以照常启动。
func Start(db *gorm.DB) *Reporter {
	if strings.TrimSpace(signSecret) == "" {
		log.Printf("使用统计未配置签名密钥，已跳过")
		return nil
	}
	installID, err := loadInstallID(db)
	if err != nil {
		log.Printf("使用统计读取安装编号失败: %v", err)
		return nil
	}
	return &Reporter{
		installID: installID,
		sessionID: newID(),
		started:   time.Now(),
		client:    &http.Client{Timeout: 10 * time.Second},
		url:       sessionURL,
		secret:    signSecret,
		version:   version.Version,
		platform:  platformName(),
	}
}

// Loop 先发 start，之后每 5 分钟发一次 heartbeat，直到 ctx 结束。
func (r *Reporter) Loop(ctx context.Context) {
	if r == nil {
		return
	}
	r.send(ctx, "start")
	ticker := time.NewTicker(heartbeatEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.send(ctx, "heartbeat")
		}
	}
}

// End 在进程退出前发送 end。
func (r *Reporter) End() {
	if r == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r.send(ctx, "end")
}

func (r *Reporter) send(ctx context.Context, event string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	duration := int64(time.Since(r.started).Seconds())
	if duration < 0 {
		duration = 0
	}
	if event == "start" {
		duration = 0
	}
	body, err := json.Marshal(sessionBody{
		InstallID:   r.installID,
		SessionID:   r.sessionID,
		Platform:    r.platform,
		Version:     r.version,
		DurationSec: duration,
		Event:       event,
	})
	if err != nil {
		log.Printf("使用统计编码失败: %v", err)
		return
	}
	if err := postSession(ctx, r.client, r.url, r.secret, body); err != nil {
		log.Printf("使用统计 %s 发送失败: %v", event, err)
	}
}

type sessionBody struct {
	InstallID   string `json:"install_id"`
	SessionID   string `json:"session_id"`
	Platform    string `json:"platform"`
	Version     string `json:"version"`
	DurationSec int64  `json:"duration_sec"`
	Event       string `json:"event"`
}

func postSession(ctx context.Context, client *http.Client, rawURL, secret string, body []byte) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	ts := fmt.Sprintf("%d", time.Now().Unix())
	nonce, err := nonce()
	if err != nil {
		return err
	}
	payload := signPayload(http.MethodPost, parsed.Path, ts, nonce, string(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MJJ-Key-Id", signKeyID)
	req.Header.Set("X-MJJ-Timestamp", ts)
	req.Header.Set("X-MJJ-Nonce", nonce)
	req.Header.Set("X-MJJ-Signature", signHex(secret, payload))
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payloadBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http %d %s", resp.StatusCode, strings.TrimSpace(string(payloadBody)))
	}
	return nil
}

func signPayload(method, path, timestamp, nonce, body string) string {
	return strings.ToUpper(method) + "\n" + path + "\n" + timestamp + "\n" + nonce + "\n" + body
}

func signHex(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func nonce() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func platformName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func loadInstallID(db *gorm.DB) (string, error) {
	if db == nil {
		return "", fmt.Errorf("database is missing")
	}
	var row model.Setting
	err := db.Where("name = ?", installSetting).Limit(1).Find(&row).Error
	if err != nil {
		return "", err
	}
	if id := settingText(row); clientIDPattern.MatchString(id) {
		return id, nil
	}
	id := newID()
	encoded, err := json.Marshal(id)
	if err != nil {
		return "", err
	}
	text := string(encoded)
	now := dbtime.DateTime{Time: dbtime.Now()}
	if row.ID == 0 {
		row = model.Setting{Name: installSetting, Value: &text, CreatedAt: &now, UpdatedAt: &now}
		return id, db.Create(&row).Error
	}
	return id, db.Model(&row).Updates(map[string]any{"value": text, "updated_at": now}).Error
}

func settingText(row model.Setting) string {
	if row.ID == 0 || row.Value == nil {
		return ""
	}
	var text string
	if json.Unmarshal([]byte(*row.Value), &text) == nil {
		return text
	}
	return strings.Trim(*row.Value, `"`)
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("install-%d", time.Now().UnixNano())
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:])
}
