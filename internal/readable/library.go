package readable

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/taskrun"
)

const (
	seasonDirName = "Season 01"
	posterName    = "poster.jpg"
)

type linkPayload struct {
	PartID int64 `json:"part_id"`
}

// EnqueuePart 给一个已经落盘的分 P 写入增量链接任务。开关关闭时不写。
func EnqueuePart(ctx context.Context, db *gorm.DB, wake func(context.Context, string) error, partID int64) error {
	if partID <= 0 {
		return nil
	}
	return enqueuePartIDs(ctx, db, wake, []int64{partID})
}

// EnqueueVideos 给这些稿件里已经落盘的分 P 各写一条增量链接任务。
func EnqueueVideos(ctx context.Context, db *gorm.DB, wake func(context.Context, string) error, videoIDs []int64) error {
	if db == nil || len(videoIDs) == 0 {
		return nil
	}
	on, err := readableEnabled(ctx, db)
	if err != nil || !on {
		return err
	}
	var ids []int64
	err = db.WithContext(ctx).Model(&model.VideoPart{}).
		Where("video_id IN ? AND video_download_path IS NOT NULL AND video_download_path <> ''", videoIDs).
		Pluck("id", &ids).Error
	if err != nil {
		return err
	}
	return enqueuePartIDs(ctx, db, wake, ids)
}

func enqueuePartIDs(ctx context.Context, db *gorm.DB, wake func(context.Context, string) error, partIDs []int64) error {
	if db == nil || len(partIDs) == 0 {
		return nil
	}
	on, err := readableEnabled(ctx, db)
	if err != nil || !on {
		return err
	}
	store := jobs.Store{DB: db}
	now := time.Now()
	for _, id := range partIDs {
		if id <= 0 {
			continue
		}
		body, err := json.Marshal(linkPayload{PartID: id})
		if err != nil {
			return err
		}
		_, err = store.Enqueue(ctx, jobs.Draft{
			Queue:      taskrun.QueueLocal,
			Type:       TypeLink,
			UniqueKey:  fmt.Sprintf("local:readable-link:%d", id),
			Payload:    string(body),
			MaxRetries: jobs.DefaultMaxRetries,
		}, now, jobs.SaveOptions{}, wake)
		if err != nil {
			return err
		}
	}
	return nil
}

func runLink(ctx context.Context, db *gorm.DB, dataDir, readableDir, payload string) error {
	var body linkPayload
	if err := json.Unmarshal([]byte(payload), &body); err != nil || body.PartID <= 0 {
		return taskrun.ErrCannotRetry
	}
	on, err := readableEnabled(ctx, db)
	if err != nil || !on {
		return err
	}
	n, err := LinkPart(ctx, db, dataDir, readableDir, body.PartID)
	if err != nil {
		return err
	}
	log.Printf("可读链接已写入 %d 个", n)
	return nil
}

// LinkPart 只为这一个分 P 建立硬链接。路径由页码决定，和下载顺序无关。
func LinkPart(ctx context.Context, db *gorm.DB, dataDir, readableDir string, partID int64) (int, error) {
	if db == nil || partID <= 0 {
		return 0, fmt.Errorf("readable link is incomplete")
	}
	root, err := readableRoot(dataDir, readableDir)
	if err != nil {
		return 0, err
	}
	var part model.VideoPart
	if err := db.WithContext(ctx).First(&part, partID).Error; err != nil {
		return 0, err
	}
	var video model.Video
	if err := db.WithContext(ctx).First(&video, part.VideoID).Error; err != nil {
		return 0, err
	}
	if video.DeletedAt != nil {
		return 0, nil
	}
	return linkDownloaded(ctx, db, dataDir, root, video, []model.VideoPart{part})
}

// UnlinkVideos 按稿件删掉各收藏夹和订阅目录下的分集链接。
func UnlinkVideos(ctx context.Context, db *gorm.DB, dataDir, readableDir string, videoIDs []int64) error {
	if db == nil || len(videoIDs) == 0 {
		return nil
	}
	on, err := readableEnabled(ctx, db)
	if err != nil || !on {
		return err
	}
	root := Dir(dataDir, readableDir)
	for _, id := range videoIDs {
		if id <= 0 {
			continue
		}
		var video model.Video
		if err := db.WithContext(ctx).First(&video, id).Error; err != nil {
			continue
		}
		var parts []model.VideoPart
		if err := db.WithContext(ctx).Where("video_id = ?", id).Find(&parts).Error; err != nil {
			return err
		}
		names, err := folderNames(ctx, db, id)
		if err != nil {
			return err
		}
		for _, name := range names {
			series := findSeriesDir(root, name, video)
			if series == "" {
				continue
			}
			for _, part := range parts {
				removeEpisode(filepath.Join(series, seasonDirName), episodeToken(part.Page))
			}
			pruneSeries(root, series)
			parent := filepath.Dir(series)
			if inside(root, parent) {
				removeDirIfEmpty(parent)
			}
		}
	}
	return nil
}

// UnlinkFolder 只删除这个收藏夹或订阅目录下这些稿件的系列目录。
func UnlinkFolder(ctx context.Context, db *gorm.DB, dataDir, readableDir, folder string, videoIDs []int64) error {
	if db == nil || len(videoIDs) == 0 || strings.TrimSpace(folder) == "" {
		return nil
	}
	on, err := readableEnabled(ctx, db)
	if err != nil || !on {
		return err
	}
	root := Dir(dataDir, readableDir)
	for _, id := range videoIDs {
		if id <= 0 {
			continue
		}
		var video model.Video
		if err := db.WithContext(ctx).First(&video, id).Error; err != nil {
			continue
		}
		series := findSeriesDir(root, folder, video)
		if series == "" || !inside(root, series) {
			continue
		}
		if err := os.RemoveAll(series); err != nil {
			return err
		}
	}
	parent := filepath.Join(root, sanitize(folder))
	if inside(root, parent) {
		removeDirIfEmpty(parent)
	}
	return nil
}

func linkDownloaded(ctx context.Context, db *gorm.DB, dataDir, root string, video model.Video, parts []model.VideoPart) (int, error) {
	ready := make([]model.VideoPart, 0, len(parts))
	for _, part := range parts {
		if partFile(dataDir, part) != "" {
			ready = append(ready, part)
		}
	}
	if len(ready) == 0 || video.DeletedAt != nil {
		return 0, nil
	}
	names, err := folderNames(ctx, db, video.ID)
	if err != nil {
		return 0, err
	}
	cover := coverPath(ctx, db, dataDir, video.ID)
	links := 0
	for _, name := range names {
		series, err := ensureSeriesDir(root, name, video)
		if err != nil {
			return links, err
		}
		if cover != "" {
			if err := createLink(cover, filepath.Join(series, posterName)); err != nil {
				log.Printf("可读封面链接失败 %s: %v", video.Title, err)
			} else {
				links++
			}
		}
		season := filepath.Join(series, seasonDirName)
		if err := os.MkdirAll(season, 0o777); err != nil {
			return links, err
		}
		for _, part := range ready {
			source := partFile(dataDir, part)
			fileName := episodeFileName(video, part)
			removeOtherEpisodeNames(season, episodeToken(part.Page), fileName)
			target := filepath.Join(season, fileName)
			if err := createLink(source, target); err != nil {
				log.Printf("可读视频链接失败 %s: %v", fileName, err)
				continue
			}
			links++
		}
	}
	return links, nil
}

func readableRoot(dataDir, readableDir string) (string, error) {
	root := Dir(dataDir, readableDir)
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("可读目录不存在: %s", root)
	}
	if unsafeDir(dataDir, root) {
		return "", fmt.Errorf("可读目录不能盖住媒体文件: %s", root)
	}
	return root, nil
}

func seriesFolder(video model.Video) string {
	name := sanitizeLimit(video.Title, 60)
	if year := video.Pubtime.Year(); year >= 2 {
		name += " (" + strconv.Itoa(year) + ")"
	}
	return name + " [" + libraryID(video) + "]"
}

func libraryID(video model.Video) string {
	id := strings.TrimSpace(video.Bvid)
	if id == "" {
		id = "av" + strconv.FormatInt(video.ID, 10)
	}
	return sanitizeLimit(id, 40)
}

func episodeToken(page int) string {
	if page < 1 {
		page = 1
	}
	return fmt.Sprintf("S01E%02d", page)
}

func episodeFileName(video model.Video, part model.VideoPart) string {
	token := episodeToken(part.Page)
	partName := sanitize(part.Part)
	if partName == "" || partName == "untitled" || partName == sanitize(video.Title) {
		return token + ".mp4"
	}
	return token + " - " + partName + ".mp4"
}

func ensureSeriesDir(root, folder string, video model.Video) (string, error) {
	parent := filepath.Join(root, sanitize(folder))
	if err := os.MkdirAll(parent, 0o777); err != nil {
		return "", err
	}
	target := filepath.Join(parent, seriesFolder(video))
	current := findSeriesDir(root, folder, video)
	if current == "" || current == target {
		if err := os.MkdirAll(target, 0o777); err != nil {
			return "", err
		}
		return target, nil
	}
	if _, err := os.Stat(target); err == nil {
		return target, nil
	}
	if err := os.Rename(current, target); err != nil {
		if _, statErr := os.Stat(target); statErr == nil {
			return target, nil
		}
		return "", err
	}
	return target, nil
}

func findSeriesDir(root, folder string, video model.Video) string {
	parent := filepath.Join(root, sanitize(folder))
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	suffix := " [" + libraryID(video) + "]"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), suffix) {
			return filepath.Join(parent, entry.Name())
		}
	}
	return ""
}

func matchesEpisode(name, token string) bool {
	if !strings.HasPrefix(name, token) {
		return false
	}
	rest := strings.TrimPrefix(name, token)
	return rest == "" || strings.HasPrefix(rest, ".") || strings.HasPrefix(rest, " - ")
}

func removeOtherEpisodeNames(season, token, keep string) {
	entries, err := os.ReadDir(season)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == keep || !matchesEpisode(entry.Name(), token) {
			continue
		}
		_ = os.Remove(filepath.Join(season, entry.Name()))
	}
}

func removeEpisode(season, token string) {
	entries, err := os.ReadDir(season)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !matchesEpisode(entry.Name(), token) {
			continue
		}
		_ = os.Remove(filepath.Join(season, entry.Name()))
	}
}

func pruneSeries(root, series string) {
	if !inside(root, series) {
		return
	}
	removeDirIfEmpty(filepath.Join(series, seasonDirName))
	entries, err := os.ReadDir(series)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.Name() == posterName {
			continue
		}
		return
	}
	_ = os.RemoveAll(series)
}

func removeDirIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}
	_ = os.Remove(dir)
}
