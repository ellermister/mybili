package syncbili

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/image/webp"
	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/model"
	"mybili/internal/taskrun"
)

const (
	CoverKindVideo        = "video"
	CoverKindFavorite     = "favorite"
	CoverKindAvatar       = "avatar"
	CoverKindSubscription = "subscription"

	coverMaxBytes = 20 << 20
)

var coverClient = &http.Client{Timeout: 30 * time.Second}

type coverPayload struct {
	Kind string `json:"kind"`
	ID   int64  `json:"id"`
	URL  string `json:"url"`
}

type coverFileInfo struct {
	mime   string
	size   int
	width  int
	height int
}

const coverDedupTTL = 6 * time.Hour

// CoverPusher 把封面消息直接推进 Redis。worker 不经过 tasks 表。
type CoverPusher interface {
	PushDedup(ctx context.Context, dedupKey, payload string) (bool, error)
	Forget(ctx context.Context, dedupKey string) error
}

// coverRedis 把封面去重推进 cover 队列。队列名和消息类型留在这里。
type coverRedis struct {
	tasks *taskrun.Tasks
}

// CoverRedis 返回封面使用的去重队列。tasks 为空时返回 nil。
func CoverRedis(tasks *taskrun.Tasks) CoverPusher {
	if tasks == nil {
		return nil
	}
	return coverRedis{tasks: tasks}
}

func (c coverRedis) PushDedup(ctx context.Context, dedupKey, payload string) (bool, error) {
	return c.tasks.PushDedup(ctx, taskrun.QueueCover, taskrun.Message{
		Type:    TypeCover,
		Payload: payload,
	}, dedupKey, coverDedupTTL)
}

func (c coverRedis) Forget(ctx context.Context, dedupKey string) error {
	return c.tasks.Forget(ctx, dedupKey)
}

// EnqueueCover 在封面地址变了、还没有关联或图片文件缺失时推进 cover 队列。
// 不写 tasks，也不在当前进程下载。同一地址还在去重窗口内时不重复推送。
// 返回值表示这次是否新推进，或换成了另一个地址。
func (s Sync) EnqueueCover(db *gorm.DB, kind string, id int64, rawURL string) (bool, error) {
	if db == nil {
		db = s.DB
	}
	if _, _, err := coverKindMeta(kind); err != nil {
		return false, err
	}
	normalized, ok := httpCoverURL(rawURL)
	if !ok || id <= 0 {
		return false, nil
	}
	stale, err := s.coverStale(db, kind, id, normalized)
	if err != nil || !stale {
		return false, err
	}
	if s.CoverPush == nil {
		return false, fmt.Errorf("cover queue is not configured")
	}
	encoded, err := json.Marshal(coverPayload{Kind: kind, ID: id, URL: normalized})
	if err != nil {
		return false, err
	}
	pushed, err := s.CoverPush.PushDedup(context.Background(), coverTaskKey(kind, id), string(encoded))
	if err != nil {
		return false, err
	}
	if pushed {
		s.noteCoverJob()
	}
	return pushed, nil
}

// CoverJobs 返回这次同步新写入或重新打开的封面任务数。
func (s Sync) CoverJobs() int {
	if s.coverJobs == nil {
		return 0
	}
	return *s.coverJobs
}

// noteCoverJob 给本次写入的封面任务计数。
func (s Sync) noteCoverJob() {
	if s.coverJobs != nil {
		*s.coverJobs++
	}
}

// Cover 下载一个对象的图片，并写入 covers / coverables。
func (s Sync) Cover(ctx context.Context, kind string, id int64, rawURL string) error {
	coverableType, coverType, err := coverKindMeta(kind)
	if err != nil {
		return err
	}
	normalized, ok := httpCoverURL(rawURL)
	if !ok || id <= 0 {
		return fmt.Errorf("cover target is empty")
	}
	filename, err := safeCoverFilename(coverFilename(normalized))
	if err != nil {
		return err
	}
	var existing model.Cover
	err = s.DB.Where("filename = ?", filename).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rel, info, err := s.downloadCoverFile(ctx, normalized, filename)
		if err != nil {
			return err
		}
		created := dbtime.DateTime{Time: s.now()}
		existing = model.Cover{
			Type:      coverType,
			URL:       normalized,
			Filename:  filename,
			Path:      rel,
			MimeType:  info.mime,
			Size:      info.size,
			Width:     info.width,
			Height:    info.height,
			CreatedAt: &created,
		}
		if err := s.DB.Create(&existing).Error; err != nil {
			if err2 := s.DB.Where("filename = ?", filename).First(&existing).Error; err2 != nil {
				return err
			}
		}
	} else if err != nil {
		return err
	} else if !s.coverFileReady(filename) {
		if _, _, err := s.downloadCoverFile(ctx, normalized, filename); err != nil {
			return err
		}
	}
	return bindCover(s.DB, existing.ID, id, coverableType, s.now())
}

// Covers 扫描还没有对上当前地址的封面并入队，只给人工查漏。
// kind 为空表示视频、收藏夹、头像、订阅都看。limit 为 0 表示不限制条数。
func (s Sync) Covers(ctx context.Context, kind string, limit int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if limit < 0 {
		return 0, fmt.Errorf("limit must be >= 1")
	}
	kinds, err := coverKinds(kind)
	if err != nil {
		return 0, err
	}
	var queued int
	for _, item := range kinds {
		if limit > 0 && queued >= limit {
			break
		}
		remain := 0
		if limit > 0 {
			remain = limit - queued
		}
		n, err := s.enqueueMissingCovers(item, remain)
		if err != nil {
			return queued, err
		}
		queued += n
	}
	return queued, nil
}

// enqueueMissingCovers 按 ID 分批找出还没对上当前地址的封面并入队。limit 为 0 时不限制条数。
func (s Sync) enqueueMissingCovers(kind string, limit int) (int, error) {
	const batchSize = 100
	var queued int
	var after int64
	for {
		rows, err := missingCoverRows(s.DB, kind, after, batchSize)
		if err != nil {
			return queued, err
		}
		if len(rows) == 0 {
			return queued, nil
		}
		for _, row := range rows {
			after = row.id
			ok, err := s.EnqueueCover(nil, kind, row.id, row.rawURL)
			if err != nil {
				return queued, err
			}
			if !ok {
				continue
			}
			queued++
			if limit > 0 && queued >= limit {
				return queued, nil
			}
		}
		if len(rows) < batchSize {
			return queued, nil
		}
	}
}

type missingCoverRow struct {
	id     int64
	rawURL string
}

// missingCoverRows 查出某一类对象里，封面地址还没有对应 coverables 的记录。
func missingCoverRows(db *gorm.DB, kind string, after int64, limit int) ([]missingCoverRow, error) {
	switch kind {
	case CoverKindVideo:
		return scanMissing(db, limit, func(q *gorm.DB) *gorm.DB {
			q = q.Model(&model.Video{}).
				Select("videos.id as id, videos.cover as raw_url").
				Joins("LEFT JOIN coverables ON coverables.coverable_id = videos.id AND coverables.coverable_type = ?", model.CoverableVideo).
				Joins("LEFT JOIN covers ON covers.id = coverables.cover_id AND covers.url = videos.cover").
				Where("videos.deleted_at IS NULL AND videos.invalid = 0 AND videos.cover <> '' AND covers.id IS NULL")
			if after > 0 {
				q = q.Where("videos.id > ?", after)
			}
			return q.Order("videos.id")
		})
	case CoverKindFavorite:
		return scanMissing(db, limit, func(q *gorm.DB) *gorm.DB {
			q = q.Model(&model.FavoriteList{}).
				Select("favorite_lists.id as id, favorite_lists.cover as raw_url").
				Joins("LEFT JOIN coverables ON coverables.coverable_id = favorite_lists.id AND coverables.coverable_type = ?", model.CoverableFavoriteList).
				Joins("LEFT JOIN covers ON covers.id = coverables.cover_id AND covers.url = favorite_lists.cover").
				Where("favorite_lists.cover <> '' AND covers.id IS NULL")
			if after > 0 {
				q = q.Where("favorite_lists.id > ?", after)
			}
			return q.Order("favorite_lists.id")
		})
	case CoverKindAvatar:
		return scanMissing(db, limit, func(q *gorm.DB) *gorm.DB {
			q = q.Model(&model.Upper{}).
				Select(`"upper".id as id, "upper".face as raw_url`).
				Joins(`LEFT JOIN coverables ON coverables.coverable_id = "upper".id AND coverables.coverable_type = ?`, model.CoverableUpper).
				Joins(`LEFT JOIN covers ON covers.id = coverables.cover_id AND covers.url = "upper".face`).
				Where(`"upper".face <> '' AND covers.id IS NULL`)
			if after > 0 {
				q = q.Where(`"upper".id > ?`, after)
			}
			return q.Order(`"upper".id`)
		})
	case CoverKindSubscription:
		return scanMissing(db, limit, func(q *gorm.DB) *gorm.DB {
			q = q.Model(&model.Subscription{}).
				Select("subscriptions.id as id, subscriptions.cover as raw_url").
				Joins("LEFT JOIN coverables ON coverables.coverable_id = subscriptions.id AND coverables.coverable_type = ?", model.CoverableSubscription).
				Joins("LEFT JOIN covers ON covers.id = coverables.cover_id AND covers.url = subscriptions.cover").
				Where("subscriptions.cover <> '' AND covers.id IS NULL")
			if after > 0 {
				q = q.Where("subscriptions.id > ?", after)
			}
			return q.Order("subscriptions.id")
		})
	default:
		return nil, fmt.Errorf("unknown cover kind %q", kind)
	}
}

type missingCoverScan struct {
	ID     int64  `gorm:"column:id"`
	RawURL string `gorm:"column:raw_url"`
}

// scanMissing 执行一类封面的缺失查询，并收成 id 和地址。
func scanMissing(db *gorm.DB, limit int, build func(*gorm.DB) *gorm.DB) ([]missingCoverRow, error) {
	q := build(db)
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []missingCoverScan
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]missingCoverRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, missingCoverRow{id: row.ID, rawURL: row.RawURL})
	}
	return out, nil
}

// coverStale 判断这个对象要不要重新入队。没有关联、地址变了或图片文件缺失时为 true。
func (s Sync) coverStale(db *gorm.DB, kind string, id int64, normalized string) (bool, error) {
	coverableType, _, err := coverKindMeta(kind)
	if err != nil {
		return false, err
	}
	filename, err := safeCoverFilename(coverFilename(normalized))
	if err != nil {
		return false, nil
	}
	var link model.Coverable
	err = db.Where("coverable_id = ? AND coverable_type = ?", id, coverableType).First(&link).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	var cover model.Cover
	if err := db.First(&cover, link.CoverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}
	if cover.URL == normalized || cover.Filename == filename {
		return !s.coverFileReady(filename), nil
	}
	return true, nil
}

// coverFileReady 判断图片文件是否已经在磁盘上，且不是空文件。
func (s Sync) coverFileReady(filename string) bool {
	info, err := os.Stat(s.imagesDir() + "/" + filename)
	return err == nil && info.Size() > 0 && !info.IsDir()
}

// bindCover 把封面记录关联到视频、收藏夹、UP 主或订阅。已有关联就改 cover_id。
func bindCover(db *gorm.DB, coverID, ownerID int64, coverableType string, now time.Time) error {
	stamp := dbtime.DateTime{Time: now}
	var link model.Coverable
	err := db.Where("coverable_id = ? AND coverable_type = ?", ownerID, coverableType).First(&link).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		link = model.Coverable{
			CoverID:       coverID,
			CoverableID:   ownerID,
			CoverableType: coverableType,
			CreatedAt:     &stamp,
			UpdatedAt:     &stamp,
		}
		if err := db.Create(&link).Error; err != nil {
			if err2 := db.Where("coverable_id = ? AND coverable_type = ?", ownerID, coverableType).First(&link).Error; err2 != nil {
				return err
			}
		} else {
			return nil
		}
	} else if err != nil {
		return err
	}
	if link.CoverID == coverID {
		return nil
	}
	return db.Model(&link).Updates(map[string]any{
		"cover_id":   coverID,
		"updated_at": stamp,
	}).Error
}

// downloadCoverFile 把图片写到 DATA_DIR/public/images，并生成同名 .hash。文件已在且哈希一致时不再请求。
func (s Sync) downloadCoverFile(ctx context.Context, rawURL, filename string) (string, coverFileInfo, error) {
	dir := s.imagesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", coverFileInfo{}, err
	}
	dest := dir + "/" + filename
	rel := "images/" + filename
	if info, ok, err := reusableCoverFile(dest); err != nil {
		return "", coverFileInfo{}, err
	} else if ok {
		return rel, info, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", coverFileInfo{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	req.Header.Set("Accept", "image/avif,image/webp,image/*,*/*;q=0.8")
	resp, err := coverClient.Do(req)
	if err != nil {
		return "", coverFileInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusPreconditionFailed {
		return "", coverFileInfo{}, ErrThrottled
	}
	if resp.StatusCode != http.StatusOK {
		return "", coverFileInfo{}, fmt.Errorf("cover http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, coverMaxBytes+1))
	if err != nil {
		return "", coverFileInfo{}, err
	}
	if len(body) == 0 {
		return "", coverFileInfo{}, fmt.Errorf("cover image is empty")
	}
	if len(body) > coverMaxBytes {
		return "", coverFileInfo{}, fmt.Errorf("cover image is too large")
	}
	tmp := dest + ".part"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return "", coverFileInfo{}, err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return "", coverFileInfo{}, err
	}
	sum := sha256.Sum256(body)
	if err := os.WriteFile(dest+".hash", []byte(hex.EncodeToString(sum[:])), 0o644); err != nil {
		return "", coverFileInfo{}, err
	}
	info, err := readCoverInfo(dest, filename)
	if err != nil {
		return "", coverFileInfo{}, err
	}
	return rel, info, nil
}

// reusableCoverFile 在图片和 .hash 都在、且哈希一致时复用本地文件。
func reusableCoverFile(path string) (coverFileInfo, bool, error) {
	hashText, err := os.ReadFile(path + ".hash")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return coverFileInfo{}, false, nil
		}
		return coverFileInfo{}, false, err
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return coverFileInfo{}, false, nil
		}
		return coverFileInfo{}, false, err
	}
	defer file.Close()
	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return coverFileInfo{}, false, err
	}
	if hex.EncodeToString(sum.Sum(nil)) != strings.TrimSpace(string(hashText)) {
		return coverFileInfo{}, false, nil
	}
	info, err := readCoverInfo(path, path)
	if err != nil {
		return coverFileInfo{}, false, err
	}
	return info, true, nil
}

// readCoverInfo 读取图片大小、类型和宽高。svg 不解码，宽高记为 0。
func readCoverInfo(path, filename string) (coverFileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return coverFileInfo{}, err
	}
	info := coverFileInfo{size: int(stat.Size())}
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".svg") {
		info.mime = "image/svg+xml"
		return info, nil
	}
	format, width, height, err := decodeCoverConfig(path)
	if err != nil {
		return coverFileInfo{}, err
	}
	info.width = width
	info.height = height
	info.mime = mimeFromFormat(format)
	return info, nil
}

// decodeCoverConfig 只读图片头，得到格式和宽高。标准库失败时再按 webp 试一次。
func decodeCoverConfig(path string) (string, int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, 0, err
	}
	defer file.Close()
	cfg, format, err := image.DecodeConfig(file)
	if err == nil {
		return format, cfg.Width, cfg.Height, nil
	}
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		return "", 0, 0, err
	}
	cfg, webpErr := webp.DecodeConfig(file)
	if webpErr != nil {
		return "", 0, 0, err
	}
	return "webp", cfg.Width, cfg.Height, nil
}

// mimeFromFormat 把图片格式转成写入 covers.mime_type 的值。
func mimeFromFormat(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// imagesDir 返回封面文件目录 DATA_DIR/public/images。
func (s Sync) imagesDir() string {
	return s.dataDir() + "/public/images"
}

// dataDir 返回数据目录。没配置时用 storage/app。
func (s Sync) dataDir() string {
	if strings.TrimSpace(s.DataDir) != "" {
		return strings.TrimRight(s.DataDir, "/")
	}
	return "storage/app"
}

// coverKinds 展开要扫描的封面类型。kind 为空时四种都扫。
func coverKinds(kind string) ([]string, error) {
	if kind == "" {
		return []string{CoverKindVideo, CoverKindFavorite, CoverKindAvatar, CoverKindSubscription}, nil
	}
	if _, _, err := coverKindMeta(kind); err != nil {
		return nil, err
	}
	return []string{kind}, nil
}

// coverKindMeta 把封面类型换成 coverables 使用的 PHP 类名，以及 covers.type。
func coverKindMeta(kind string) (string, string, error) {
	switch kind {
	case CoverKindVideo:
		return model.CoverableVideo, CoverKindVideo, nil
	case CoverKindFavorite:
		return model.CoverableFavoriteList, CoverKindFavorite, nil
	case CoverKindAvatar:
		return model.CoverableUpper, CoverKindAvatar, nil
	case CoverKindSubscription:
		return model.CoverableSubscription, CoverKindSubscription, nil
	default:
		return "", "", fmt.Errorf("unknown cover kind %q", kind)
	}
}

// coverTaskKey 生成封面任务的唯一键，格式是 cover:类型:ID。
func coverTaskKey(kind string, id int64) string {
	return fmt.Sprintf("cover:%s:%d", kind, id)
}

// payloadURL 从任务 payload 里取出待下载地址。解析失败时返回空字符串。
func payloadURL(payload string) string {
	var decoded coverPayload
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return ""
	}
	return decoded.URL
}

// httpCoverURL 只接受 http 或 https 地址。以 // 开头的补成 https。
func httpCoverURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "", false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false
	}
	return raw, true
}

// safeCoverFilename 拒绝空文件名和带路径的名字。
func safeCoverFilename(name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("cover filename is empty")
	}
	return name, nil
}

// coverFilename 用地址最后一段做文件名。取不到时改用地址的 base64。
func coverFilename(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	parts := strings.Split(rawURL, "/")
	filename := parts[len(parts)-1]
	if filename == "" {
		filename = base64.StdEncoding.EncodeToString([]byte(rawURL)) + ".jpg"
	}
	filename = sanitizeCoverFilename(filename)
	if filename == "" {
		filename = sanitizeCoverFilename(base64.StdEncoding.EncodeToString([]byte(rawURL)) + ".jpg")
	}
	return filename
}

// sanitizeCoverFilename 去掉文件名里的符号，紧挨着 jpg、png、gif、svg、webp 的符号保留。
func sanitizeCoverFilename(filename string) string {
	exts := []string{"jpg", "png", "gif", "svg", "webp"}
	var b strings.Builder
	for i := 0; i < len(filename); i++ {
		c := filename[i]
		if isCoverNameChar(c) {
			b.WriteByte(c)
			continue
		}
		rest := filename[i+1:]
		keep := false
		for _, ext := range exts {
			if strings.HasPrefix(rest, ext) {
				keep = true
				break
			}
		}
		if keep {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// isCoverNameChar 判断这个字节是不是字母或数字。
func isCoverNameChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
