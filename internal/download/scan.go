package download

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/schedule"
	"mybili/internal/taskrun"
)

// EnqueueLost 检查已经记过下载的视频和音频。文件不在时写成 bili-default。
// 没下过的行不入队。开关关掉时返回 0。
func EnqueueLost(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, wake func(context.Context, string) error) (int, error) {
	on, err := settingOn(db, "video_download_enabled")
	if err != nil || !on {
		return 0, err
	}
	rule, err := loadRules(db, now)
	if err != nil {
		return 0, err
	}
	videos, err := enqueueLostVideos(ctx, db, dataDir, videoID, now, rule, wake)
	if err != nil {
		return videos, err
	}
	audios, err := enqueueLostAudios(ctx, db, dataDir, videoID, now, rule, wake)
	return videos + audios, err
}

func settingOn(db *gorm.DB, name string) (bool, error) {
	text, err := readSettingString(db, name)
	if err != nil || text == "" {
		return false, err
	}
	return text == "on", nil
}

// multiPartOn 读取多分 P 下载开关。关掉时视频和弹幕都只处理第 1 P。
func multiPartOn(db *gorm.DB) (bool, error) {
	text, err := readSettingString(db, "multi_partition_download_enabled")
	if err != nil {
		return false, err
	}
	return text == "on", nil
}

type storedPart struct {
	Candidate
	DownloadPath *string
	DownloadedAt *dbtime.DateTime
}

type storedAudio struct {
	Candidate
	Sid          int64
	DownloadPath *string
	DownloadedAt *dbtime.DateTime
}

func enqueueLostVideos(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, rule rules, wake func(context.Context, string) error) (int, error) {
	var rows []storedPart
	query := db.WithContext(ctx).Table("video_parts").
		Select(`video_parts.id, video_parts.video_id, video_parts.page, video_parts.cid, video_parts.duration,
			videos.title, videos.duration as video_duration, videos.fav_time, videos.pubtime,
			video_parts.video_download_path as download_path, video_parts.video_downloaded_at as downloaded_at`).
		Joins("JOIN videos ON videos.id = video_parts.video_id").
		Where("videos.deleted_at IS NULL AND videos.invalid = 0 AND videos.type <> ?", 12).
		Where("(video_parts.video_download_path IS NOT NULL AND video_parts.video_download_path <> '') OR video_parts.video_downloaded_at IS NOT NULL")
	if videoID > 0 {
		query = query.Where("video_parts.video_id = ?", videoID)
	}
	if err := query.Order("video_parts.id").Scan(&rows).Error; err != nil {
		return 0, err
	}
	return enqueueStored(ctx, db, now, rule, jobs.SaveOptions{}, wake, rows, func(row storedPart) (schedule.Draft, error) {
		if fileReady(dataDir, row.DownloadPath, videoFilePath(dataDir, row.VideoID, row.Cid)) {
			return schedule.Draft{}, nil
		}
		return partDraft(TypeVideo, row.Candidate)
	})
}

func enqueueLostAudios(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, rule rules, wake func(context.Context, string) error) (int, error) {
	var rows []storedAudio
	query := db.WithContext(ctx).Table("audio_parts").
		Select(`audio_parts.id, audio_parts.video_id, audio_parts.sid, audio_parts.duration,
			videos.title, videos.duration as video_duration, videos.fav_time, videos.pubtime,
			audio_parts.audio_download_path as download_path, audio_parts.audio_downloaded_at as downloaded_at`).
		Joins("JOIN videos ON videos.id = audio_parts.video_id").
		Where("videos.deleted_at IS NULL AND videos.invalid = 0").
		Where("(audio_parts.audio_download_path IS NOT NULL AND audio_parts.audio_download_path <> '') OR audio_parts.audio_downloaded_at IS NOT NULL")
	if videoID > 0 {
		query = query.Where("audio_parts.video_id = ?", videoID)
	}
	if err := query.Order("audio_parts.id").Scan(&rows).Error; err != nil {
		return 0, err
	}
	candidates := make([]Candidate, len(rows))
	for i := range rows {
		rows[i].Page = 1
		candidates[i] = rows[i].Candidate
	}
	favs, subs, err := relations(ctx, db, candidates)
	if err != nil {
		return 0, err
	}
	quality, err := loadQualitySettings(db)
	if err != nil {
		return 0, err
	}
	store := jobs.Store{DB: db}
	opened := 0
	for _, row := range rows {
		if audioReady(dataDir, row.DownloadPath, row.Sid) {
			continue
		}
		if !rule.allow(row.Candidate, favs[row.VideoID], subs[row.VideoID]) {
			continue
		}
		if _, skip := quality.choose(favs[row.VideoID], subs[row.VideoID]); skip {
			continue
		}
		body, err := json.Marshal(audioPayload{VideoID: row.VideoID, PartID: row.ID, Sid: row.Sid})
		if err != nil {
			return opened, err
		}
		result, err := store.Enqueue(ctx, jobs.Draft{
			Queue:      taskrun.QueueBiliDefault,
			Type:       TypeAudio,
			UniqueKey:  fmt.Sprintf("download:audio:%d", row.VideoID),
			Payload:    string(body),
			MaxRetries: jobs.DefaultMaxRetries,
		}, now, jobs.SaveOptions{}, wake)
		if err != nil {
			return opened, err
		}
		if result.Opened {
			opened++
		}
	}
	return opened, nil
}

// VideoDownloadGap 查看一个稿件的分 P。不写任务，也不请求 Bilibili。
// missing 表示至少有一个分 P 的文件不在，并且下载规则允许它。
// newTask 表示这次会新开下载任务。下载开关关掉、规则排除，或这个分 P 已经有过任务时为 false。
func VideoDownloadGap(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time) (missing bool, newTask bool, err error) {
	if db == nil || videoID <= 0 {
		return false, false, nil
	}
	on, err := settingOn(db, "video_download_enabled")
	if err != nil || !on {
		return false, false, err
	}
	rule, err := loadRules(db, now)
	if err != nil {
		return false, false, err
	}
	var rows []storedPart
	err = db.WithContext(ctx).Table("video_parts").
		Select(`video_parts.id, video_parts.video_id, video_parts.page, video_parts.cid, video_parts.duration,
			videos.title, videos.duration as video_duration, videos.fav_time, videos.pubtime,
			video_parts.video_download_path as download_path, video_parts.video_downloaded_at as downloaded_at`).
		Joins("JOIN videos ON videos.id = video_parts.video_id").
		Where("video_parts.video_id = ?", videoID).
		Where("videos.deleted_at IS NULL AND videos.invalid = 0 AND videos.type <> ?", 12).
		Order("video_parts.page").
		Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return false, false, err
	}
	candidates := make([]Candidate, len(rows))
	for i := range rows {
		candidates[i] = rows[i].Candidate
	}
	favs, subs, err := relations(ctx, db, candidates)
	if err != nil {
		return false, false, err
	}
	quality, err := loadQualitySettings(db)
	if err != nil {
		return false, false, err
	}
	for _, row := range rows {
		if !rule.allow(row.Candidate, favs[row.VideoID], subs[row.VideoID]) {
			continue
		}
		if _, skip := quality.choose(favs[row.VideoID], subs[row.VideoID]); skip {
			continue
		}
		if fileReady(dataDir, row.DownloadPath, videoFilePath(dataDir, row.VideoID, row.Cid)) {
			continue
		}
		missing = true
		var n int64
		err = db.WithContext(ctx).Model(&model.Job{}).Where("unique_key = ?", fmt.Sprintf("download:video:%d", row.ID)).Count(&n).Error
		if err != nil {
			return false, false, err
		}
		if n == 0 {
			return true, true, nil
		}
	}
	return missing, false, nil
}

// EnqueueParts 给一个稿件里还没有文件的分 P 写下载任务。
// 固定文件名已经存在时，记下下载路径、时间和已下载分 P 数，不新建任务。
// 下载开关关掉、或这个分 P 已经有过任务时，不新开、也不把结束的任务重新打开。
func EnqueueParts(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, wake func(context.Context, string) error) (int, error) {
	return enqueueParts(ctx, db, dataDir, videoID, now, wake, jobs.SaveOptions{ReopenAfter: time.Unix(1, 0)})
}

// RequeueParts 在用户删掉本地文件后重新排队。已经结束的下载任务会再打开。
func RequeueParts(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, wake func(context.Context, string) error) (int, error) {
	return enqueueParts(ctx, db, dataDir, videoID, now, wake, jobs.SaveOptions{})
}

func enqueueParts(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, wake func(context.Context, string) error, opt jobs.SaveOptions) (int, error) {
	if videoID <= 0 {
		return 0, nil
	}
	on, err := settingOn(db, "video_download_enabled")
	if err != nil || !on {
		return 0, err
	}
	rule, err := loadRules(db, now)
	if err != nil {
		return 0, err
	}
	var rows []storedPart
	err = db.WithContext(ctx).Table("video_parts").
		Select(`video_parts.id, video_parts.video_id, video_parts.page, video_parts.cid, video_parts.duration,
			videos.title, videos.duration as video_duration, videos.fav_time, videos.pubtime,
			video_parts.video_download_path as download_path, video_parts.video_downloaded_at as downloaded_at`).
		Joins("JOIN videos ON videos.id = video_parts.video_id").
		Where("video_parts.video_id = ?", videoID).
		Where("videos.deleted_at IS NULL AND videos.invalid = 0 AND videos.type <> ?", 12).
		Order("video_parts.page").
		Scan(&rows).Error
	if err != nil {
		return 0, err
	}
	return enqueueStored(ctx, db, now, rule, opt, wake, rows, func(row storedPart) (schedule.Draft, error) {
		canonical := videoFilePath(dataDir, row.VideoID, row.Cid)
		if storedFileReady(dataDir, row.DownloadPath) {
			return schedule.Draft{}, nil
		}
		if fileExists(canonical) {
			err := rememberExistingVideo(ctx, db, row.ID, row.VideoID, row.Cid, canonical, now)
			return schedule.Draft{}, err
		}
		return partDraft(TypeVideo, row.Candidate)
	})
}

// rememberExistingVideo 把已经在固定文件名上的视频记成已下载，不再排下载任务。
// 下载时间用文件修改时间。稿件上的已下载分 P 数按库里的路径重算。
func rememberExistingVideo(ctx context.Context, db *gorm.DB, partID, videoID, cid int64, path string, now time.Time) error {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() == 0 {
		if err == nil {
			err = fmt.Errorf("video file is empty")
		}
		return err
	}
	downloaded := dbtime.DateTime{Time: info.ModTime()}
	updated := dbtime.DateTime{Time: now}
	err = db.WithContext(ctx).Model(&model.VideoPart{}).Where("id = ?", partID).Updates(map[string]any{
		"video_download_path": storageVideoPath(videoID, cid),
		"video_downloaded_at": downloaded,
		"updated_at":          updated,
	}).Error
	if err != nil {
		return err
	}
	var count int64
	if err := db.WithContext(ctx).Model(&model.VideoPart{}).
		Where("video_id = ? AND video_download_path IS NOT NULL AND video_download_path <> ''", videoID).
		Count(&count).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&model.Video{}).Where("id = ?", videoID).Updates(map[string]any{
		"video_downloaded_num": count,
		"video_downloaded_at":  downloaded,
		"updated_at":           updated,
	}).Error
}

// EnqueueAudio 给还没有下载任务的音频排队。已经结束的任务不重新打开。
func EnqueueAudio(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, wake func(context.Context, string) error) (int, error) {
	return enqueueAudio(ctx, db, dataDir, videoID, now, wake, jobs.SaveOptions{ReopenAfter: time.Unix(1, 0)})
}

// RequeueAudio 在用户删掉音频文件后重新排队。已经结束的任务会再打开。
func RequeueAudio(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, wake func(context.Context, string) error) (int, error) {
	return enqueueAudio(ctx, db, dataDir, videoID, now, wake, jobs.SaveOptions{})
}

func enqueueAudio(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time, wake func(context.Context, string) error, opt jobs.SaveOptions) (int, error) {
	if videoID <= 0 {
		return 0, nil
	}
	on, err := settingOn(db, "video_download_enabled")
	if err != nil || !on {
		return 0, err
	}
	rule, err := loadRules(db, now)
	if err != nil {
		return 0, err
	}
	var row storedAudio
	err = db.WithContext(ctx).Table("audio_parts").
		Select(`audio_parts.id, audio_parts.video_id, audio_parts.sid, audio_parts.duration,
			videos.title, videos.duration as video_duration, videos.fav_time, videos.pubtime,
			audio_parts.audio_download_path as download_path, audio_parts.audio_downloaded_at as downloaded_at`).
		Joins("JOIN videos ON videos.id = audio_parts.video_id").
		Where("audio_parts.video_id = ?", videoID).
		Where("videos.deleted_at IS NULL AND videos.invalid = 0").
		Limit(1).
		Scan(&row).Error
	if err != nil || row.ID == 0 {
		return 0, err
	}
	row.Page = 1
	if audioReady(dataDir, row.DownloadPath, row.Sid) {
		return 0, nil
	}
	favs, subs, err := relations(ctx, db, []Candidate{row.Candidate})
	if err != nil {
		return 0, err
	}
	if !rule.allow(row.Candidate, favs[row.VideoID], subs[row.VideoID]) {
		return 0, nil
	}
	quality, err := loadQualitySettings(db)
	if err != nil {
		return 0, err
	}
	if _, skip := quality.choose(favs[row.VideoID], subs[row.VideoID]); skip {
		return 0, nil
	}
	body, err := json.Marshal(audioPayload{VideoID: row.VideoID, PartID: row.ID, Sid: row.Sid})
	if err != nil {
		return 0, err
	}
	result, err := jobs.Store{DB: db}.Enqueue(ctx, jobs.Draft{
		Queue:      taskrun.QueueBiliDefault,
		Type:       TypeAudio,
		UniqueKey:  fmt.Sprintf("download:audio:%d", row.VideoID),
		Payload:    string(body),
		MaxRetries: jobs.DefaultMaxRetries,
	}, now, opt, wake)
	if err != nil || !result.Opened {
		return 0, err
	}
	return 1, nil
}

func enqueueStored(ctx context.Context, db *gorm.DB, now time.Time, rule rules, opt jobs.SaveOptions, wake func(context.Context, string) error, rows []storedPart, build func(storedPart) (schedule.Draft, error)) (int, error) {
	candidates := make([]Candidate, len(rows))
	for i := range rows {
		candidates[i] = rows[i].Candidate
	}
	favs, subs, err := relations(ctx, db, candidates)
	if err != nil {
		return 0, err
	}
	quality, err := loadQualitySettings(db)
	if err != nil {
		return 0, err
	}
	store := jobs.Store{DB: db}
	opened := 0
	for _, row := range rows {
		if !rule.allow(row.Candidate, favs[row.VideoID], subs[row.VideoID]) {
			continue
		}
		if _, skip := quality.choose(favs[row.VideoID], subs[row.VideoID]); skip {
			continue
		}
		draft, err := build(row)
		if err != nil || draft.UniqueKey == "" {
			if err != nil {
				return opened, err
			}
			continue
		}
		result, err := store.Enqueue(ctx, jobs.Draft{
			Queue:      draft.Queue,
			Type:       draft.Type,
			UniqueKey:  draft.UniqueKey,
			Payload:    draft.Payload,
			MaxRetries: draft.MaxRetries,
		}, now, opt, wake)
		if err != nil {
			return opened, err
		}
		if result.Opened {
			opened++
		}
	}
	return opened, nil
}

func fileReady(dataDir string, stored *string, canonical string) bool {
	if storedFileReady(dataDir, stored) {
		return true
	}
	return fileExists(canonical)
}

func storedFileReady(dataDir string, stored *string) bool {
	return stored != nil && fileExists(resolveStored(dataDir, *stored))
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

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}

func relations(ctx context.Context, db *gorm.DB, rows []Candidate) (map[int64][]int64, map[int64][]int64, error) {
	favs := map[int64][]int64{}
	subs := map[int64][]int64{}
	if len(rows) == 0 {
		return favs, subs, nil
	}
	ids := make([]int64, 0, len(rows))
	seen := map[int64]struct{}{}
	for _, row := range rows {
		if _, ok := seen[row.VideoID]; ok {
			continue
		}
		seen[row.VideoID] = struct{}{}
		ids = append(ids, row.VideoID)
	}
	var favLinks []model.FavoriteListVideo
	if err := db.WithContext(ctx).Where("video_id IN ?", ids).Find(&favLinks).Error; err != nil {
		return nil, nil, err
	}
	for _, link := range favLinks {
		favs[link.VideoID] = append(favs[link.VideoID], link.FavoriteListID)
	}
	var subLinks []model.SubscriptionVideo
	if err := db.WithContext(ctx).Where("video_id IN ?", ids).Find(&subLinks).Error; err != nil {
		return nil, nil, err
	}
	for _, link := range subLinks {
		subs[link.VideoID] = append(subs[link.VideoID], link.SubscriptionID)
	}
	return favs, subs, nil
}

func partDraft(taskType string, row Candidate) (schedule.Draft, error) {
	body, err := json.Marshal(partPayload{
		VideoID: row.VideoID,
		PartID:  row.ID,
		Page:    row.Page,
		Cid:     row.Cid,
	})
	if err != nil {
		return schedule.Draft{}, err
	}
	key := fmt.Sprintf("download:video:%d", row.ID)
	if taskType == TypeDanmaku {
		key = fmt.Sprintf("download:danmaku:%d", row.ID)
	}
	return schedule.Draft{
		Queue:      taskrun.QueueBiliDefault,
		Type:       taskType,
		UniqueKey:  key,
		Payload:    string(body),
		MaxRetries: jobs.DefaultMaxRetries,
	}, nil
}

const danmakuFreshFor = 7 * 24 * time.Hour

// EnqueueDanmaku 给一个稿件里需要弹幕的分 P 写下载任务。
// 弹幕开关关掉时不写。多分 P 下载关掉时只写第 1 P。
// 7 天内下过的跳过。没下过的不把已经结束的任务重新打开。
func EnqueueDanmaku(ctx context.Context, db *gorm.DB, videoID int64, now time.Time, wake func(context.Context, string) error) (int, error) {
	if videoID <= 0 {
		return 0, nil
	}
	on, err := settingOn(db, "danmaku_download_enabled")
	if err != nil || !on {
		return 0, err
	}
	multi, err := multiPartOn(db)
	if err != nil {
		return 0, err
	}
	var video model.Video
	err = db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL AND invalid = 0 AND type <> ?", videoID, 12).Limit(1).Find(&video).Error
	if err != nil || video.ID == 0 {
		return 0, err
	}
	var parts []model.VideoPart
	if err := db.WithContext(ctx).Where("video_id = ?", videoID).Order("page").Find(&parts).Error; err != nil {
		return 0, err
	}
	store := jobs.Store{DB: db}
	opened := 0
	for _, part := range parts {
		if part.Cid == 0 || (!multi && part.Page > 1) {
			continue
		}
		opt := jobs.SaveOptions{ReopenAfter: time.Unix(1, 0)}
		if part.DanmakuDownloadedAt != nil && !part.DanmakuDownloadedAt.Time.IsZero() {
			age := now.Sub(part.DanmakuDownloadedAt.Time)
			if age < danmakuFreshFor {
				continue
			}
			opt = jobs.SaveOptions{}
		}
		draft, err := partDraft(TypeDanmaku, Candidate{
			ID:      part.ID,
			VideoID: part.VideoID,
			Page:    part.Page,
			Cid:     part.Cid,
		})
		if err != nil {
			return opened, err
		}
		result, err := store.Enqueue(ctx, jobs.Draft{
			Queue:      draft.Queue,
			Type:       draft.Type,
			UniqueKey:  draft.UniqueKey,
			Payload:    draft.Payload,
			MaxRetries: draft.MaxRetries,
		}, now, opt, wake)
		if err != nil {
			return opened, err
		}
		if result.Opened {
			opened++
		}
	}
	return opened, nil
}

// RefreshDanmaku 给一个视频的分 P 写弹幕任务。已经结束的行会重新打开。
// 多分 P 下载关掉时只写第 1 P。
func RefreshDanmaku(ctx context.Context, db *gorm.DB, videoID int64, now time.Time, wake func(context.Context, string) error) (int, error) {
	multi, err := multiPartOn(db)
	if err != nil {
		return 0, err
	}
	var parts []model.VideoPart
	if err := db.WithContext(ctx).Where("video_id = ?", videoID).Order("page").Find(&parts).Error; err != nil {
		return 0, err
	}
	store := jobs.Store{DB: db}
	queued := 0
	for _, part := range parts {
		if !multi && part.Page > 1 {
			continue
		}
		draft, err := partDraft(TypeDanmaku, Candidate{
			ID:      part.ID,
			VideoID: part.VideoID,
			Page:    part.Page,
			Cid:     part.Cid,
		})
		if err != nil {
			return queued, err
		}
		result, err := store.Enqueue(ctx, jobs.Draft{
			Queue:      draft.Queue,
			Type:       draft.Type,
			UniqueKey:  draft.UniqueKey,
			Payload:    draft.Payload,
			MaxRetries: draft.MaxRetries,
		}, now, jobs.SaveOptions{}, wake)
		if err != nil {
			return queued, err
		}
		if result.ID > 0 {
			queued++
		}
	}
	return queued, nil
}
