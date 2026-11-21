package readable

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"

	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/taskrun"
)

const (
	// TypeNames 是手动全量重建可读文件名的任务。
	TypeNames = "readable-names"
	// TypeLink 是单个分 P 的增量硬链接。
	TypeLink = "readable-link"

	settingReadable = "human_readable_name_enabled"
)

var controlRE = regexp.MustCompile(`[\x00-\x1F\x7F]`)

// Register 把可读文件名登记到 worker。全量重建只在任务被手动写入时执行。
func Register(tasks *taskrun.Tasks, db *gorm.DB, dataDir, readableDir string) {
	if tasks == nil {
		return
	}
	store := jobs.Store{DB: db}
	tasks.Handle(TypeNames, func(ctx context.Context, msg taskrun.Message) error {
		return runSaved(ctx, store, tasks, msg, func(ctx context.Context) error {
			n, err := Generate(ctx, db, dataDir, readableDir)
			if err != nil {
				return err
			}
			log.Printf("可读文件名已生成 %d 个链接", n)
			return nil
		})
	})
	tasks.Handle(TypeLink, func(ctx context.Context, msg taskrun.Message) error {
		return runSaved(ctx, store, tasks, msg, func(ctx context.Context) error {
			return runLink(ctx, db, dataDir, readableDir, msg.Payload)
		})
	})
}

// Generate 清空可读目录下的子目录，再按收藏夹和订阅为已下载视频建立硬链接。
// 目录不存在时返回错误，不会创建根目录。
func Generate(ctx context.Context, db *gorm.DB, dataDir, readableDir string) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("readable names are incomplete")
	}
	root := Dir(dataDir, readableDir)
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return 0, fmt.Errorf("可读目录不存在: %s", root)
	}
	if unsafeDir(dataDir, root) {
		return 0, fmt.Errorf("可读目录不能盖住媒体文件: %s", root)
	}
	if err := clearChildren(root); err != nil {
		return 0, err
	}
	var videos []model.Video
	err = db.WithContext(ctx).
		Where("deleted_at IS NULL AND video_downloaded_num > 0").
		Order("id").
		Find(&videos).Error
	if err != nil {
		return 0, err
	}
	links := 0
	for _, video := range videos {
		n, err := linkVideo(ctx, db, dataDir, root, video)
		if err != nil {
			return links, err
		}
		links += n
	}
	return links, nil
}

// Dir 返回可读目录。配置为空时用数据目录下的 public/medias。
func Dir(dataDir, configured string) string {
	configured = strings.TrimSpace(configured)
	if configured != "" {
		return configured
	}
	return filepath.Join(dataDir, "public", "medias")
}

func linkVideo(ctx context.Context, db *gorm.DB, dataDir, root string, video model.Video) (int, error) {
	var parts []model.VideoPart
	if err := db.WithContext(ctx).Where("video_id = ?", video.ID).Order("page, id").Find(&parts).Error; err != nil {
		return 0, err
	}
	return linkDownloaded(ctx, db, dataDir, root, video, parts)
}

func folderNames(ctx context.Context, db *gorm.DB, videoID int64) ([]string, error) {
	var favs []string
	err := db.WithContext(ctx).Table("favorite_lists").
		Joins("JOIN favorite_list_videos ON favorite_list_videos.favorite_list_id = favorite_lists.id").
		Where("favorite_list_videos.video_id = ?", videoID).
		Order("favorite_lists.id").
		Pluck("favorite_lists.title", &favs).Error
	if err != nil {
		return nil, err
	}
	var subs []string
	err = db.WithContext(ctx).Table("subscriptions").
		Joins("JOIN subscription_videos ON subscription_videos.subscription_id = subscriptions.id").
		Where("subscription_videos.video_id = ?", videoID).
		Order("subscriptions.id").
		Pluck("subscriptions.name", &subs).Error
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	names := make([]string, 0, len(favs)+len(subs))
	for _, name := range append(favs, subs...) {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names, nil
}

func partFile(dataDir string, part model.VideoPart) string {
	if part.VideoDownloadPath != nil {
		if path := existingFile(resolveStored(dataDir, *part.VideoDownloadPath)); path != "" {
			return path
		}
	}
	if path := existingFile(resolveStored(dataDir, legacyVideoPath(part.VideoID, int64(part.Page)))); path != "" {
		return path
	}
	if part.Cid > 0 {
		if path := existingFile(resolveStored(dataDir, legacyVideoPath(part.VideoID, part.Cid))); path != "" {
			return path
		}
	}
	name := fmt.Sprintf("%d.%d.mp4", part.VideoID, part.Cid)
	return existingFile(filepath.Join(dataDir, "public", "videos", name))
}

func coverPath(ctx context.Context, db *gorm.DB, dataDir string, videoID int64) string {
	var cover model.Cover
	err := db.WithContext(ctx).Table("covers").
		Joins("JOIN coverables ON coverables.cover_id = covers.id").
		Where("coverables.coverable_type = ? AND coverables.coverable_id = ?", model.CoverableVideo, videoID).
		Order("covers.id").
		Limit(1).
		Find(&cover).Error
	if err != nil || cover.ID == 0 {
		return ""
	}
	return existingFile(resolveStored(dataDir, cover.Path))
}

func legacyVideoPath(videoID, part int64) string {
	id := fmt.Sprint(videoID)
	if part <= 1 {
		return "/storage/videos/" + id + ".mp4"
	}
	return "/storage/videos/" + id + ".part" + fmt.Sprint(part) + ".mp4"
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

func existingFile(path string) string {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() == 0 {
		return ""
	}
	return path
}

func createLink(source, target string) error {
	if source == "" || target == "" {
		return fmt.Errorf("link path is empty")
	}
	src, err := os.Stat(source)
	if err != nil {
		return err
	}
	if dst, err := os.Stat(target); err == nil && os.SameFile(src, dst) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o777); err != nil {
		return err
	}
	if info, err := os.Lstat(target); err == nil && !info.IsDir() {
		if err := os.Remove(target); err != nil {
			return err
		}
	}
	if err := os.Link(source, target); err == nil {
		return nil
	} else if out, err2 := exec.Command("ln", source, target).CombinedOutput(); err2 != nil {
		return fmt.Errorf("%v; ln: %v %s", err, err2, bytesTrim(out))
	}
	return nil
}

func bytesTrim(b []byte) string {
	return strings.TrimSpace(string(b))
}

func sanitize(name string) string {
	return sanitizeLimit(name, 80)
}

func sanitizeLimit(name string, limit int) string {
	replacer := strings.NewReplacer(
		"/", " ", "\\", " ", ":", " ", "*", " ", "?", " ", "\"", " ", "<", " ", ">", " ", "|", " ",
		"#", " ", "%", " ", "&", " ", "{", " ", "}", " ",
		"`", " ", "$", " ", "!", " ", "@", " ", "+", " ", "=", " ",
	)
	name = replacer.Replace(name)
	name = controlRE.ReplaceAllString(name, "")
	name = strings.Join(strings.FieldsFunc(name, unicode.IsSpace), " ")
	name = strings.TrimSpace(name)
	runes := []rune(name)
	if limit > 0 && len(runes) > limit {
		name = strings.TrimSpace(string(runes[:limit]))
	}
	if name == "" {
		return "untitled"
	}
	return name
}

func clearChildren(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if !inside(root, path) {
			continue
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(path); err != nil {
				return err
			}
			continue
		}
		if !info.IsDir() {
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func unsafeDir(dataDir, root string) bool {
	root = filepath.Clean(root)
	public := filepath.Clean(filepath.Join(dataDir, "public"))
	videos := filepath.Clean(filepath.Join(public, "videos"))
	images := filepath.Clean(filepath.Join(public, "images"))
	if root == filepath.Clean(dataDir) || root == public || root == videos || root == images {
		return true
	}
	rel, err := filepath.Rel(root, videos)
	if err != nil {
		return true
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func inside(root, path string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil || rel == "." {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func settingOn(db *gorm.DB, name string) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("database is missing")
	}
	var row model.Setting
	err := db.Where("name = ?", name).Limit(1).Find(&row).Error
	if err != nil || row.ID == 0 || row.Value == nil {
		return false, err
	}
	var text string
	if json.Unmarshal([]byte(*row.Value), &text) == nil {
		return text == "on", nil
	}
	return strings.Trim(*row.Value, `"`) == "on", nil
}

func readableEnabled(ctx context.Context, db *gorm.DB) (bool, error) {
	if db == nil {
		return false, nil
	}
	return settingOn(db.WithContext(ctx), settingReadable)
}

func runSaved(ctx context.Context, store jobs.Store, tasks *taskrun.Tasks, msg taskrun.Message, work func(context.Context) error) error {
	job, err := store.Load(ctx, msg.ID)
	if err != nil {
		return err
	}
	switch job.Status {
	case jobs.StatusCancelled, jobs.StatusDone, jobs.StatusFailed:
		return nil
	}
	ok, err := store.MarkRunning(ctx, msg.ID, time.Now())
	if err != nil || !ok {
		return err
	}
	err = work(ctx)
	if err == nil {
		_, err = store.MarkDone(ctx, msg.ID, time.Now())
		return err
	}
	if errors.Is(err, taskrun.ErrCannotRetry) || job.RetryCount+1 >= maxRetries(job) {
		_, failErr := store.MarkFailed(ctx, msg.ID, err.Error(), job.RetryCount+1, time.Now())
		if failErr != nil {
			return failErr
		}
		return taskrun.ErrCannotRetry
	}
	if _, pendingErr := store.MarkPending(ctx, msg.ID, job.RetryCount+1, err.Error(), time.Now()); pendingErr != nil {
		return pendingErr
	}
	if tasks != nil {
		if wakeErr := tasks.Wake(ctx, job.Queue); wakeErr != nil {
			log.Printf("通知投喂失败 %s: %v", job.Queue, wakeErr)
		}
	}
	return err
}

func maxRetries(job model.Job) int {
	if job.MaxRetries > 0 {
		return job.MaxRetries
	}
	return jobs.DefaultMaxRetries
}
