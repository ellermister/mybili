package download

import (
	"context"
	"os"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

// ReconcileResult 是一次本地对账的结果。
// Noted 是新记下或改到现存文件上的分 P、音频条数。
// Recounted 是下载数量被改过的稿件数。
type ReconcileResult struct {
	Noted     int
	Recounted int
}

// ReconcileDownloads 把磁盘上已经有的视频和音频记到分 P，并按路径重算稿件上的下载数量。
// 视频沿用播放时的查找：已保存的路径还在就保留，否则旧文件名，再否则 视频ID.cid.mp4。
// 音频先看 au{sid}.m4a，再看旧的 mp3。
// 文件不在时不改已有路径，也不写下载任务。已删除的稿件不处理。videoID 小于等于 0 时扫全部。
func ReconcileDownloads(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time) (ReconcileResult, error) {
	var result ReconcileResult
	noted, err := noteExistingVideos(ctx, db, dataDir, videoID, now)
	if err != nil {
		return result, err
	}
	result.Noted += noted
	noted, err = noteExistingAudios(ctx, db, dataDir, videoID, now)
	if err != nil {
		return result, err
	}
	result.Noted += noted
	result.Recounted, err = recountDownloads(ctx, db, videoID, now)
	return result, err
}

func noteExistingVideos(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time) (int, error) {
	var parts []model.VideoPart
	query := db.WithContext(ctx).Model(&model.VideoPart{}).
		Select("video_parts.*").
		Joins("JOIN videos ON videos.id = video_parts.video_id AND videos.deleted_at IS NULL").
		Order("video_parts.id")
	if videoID > 0 {
		query = query.Where("video_parts.video_id = ?", videoID)
	}
	if err := query.Find(&parts).Error; err != nil {
		return 0, err
	}
	noted := 0
	updated := dbtime.DateTime{Time: now}
	for _, part := range parts {
		stored, mod, ok := existingVideoFile(dataDir, part)
		if !ok {
			continue
		}
		downloaded := dbtime.DateTime{Time: mod}
		err := db.WithContext(ctx).Model(&model.VideoPart{}).Where("id = ?", part.ID).Updates(map[string]any{
			"video_download_path": stored,
			"video_downloaded_at": downloaded,
			"updated_at":          updated,
		}).Error
		if err != nil {
			return noted, err
		}
		noted++
	}
	return noted, nil
}

func noteExistingAudios(ctx context.Context, db *gorm.DB, dataDir string, videoID int64, now time.Time) (int, error) {
	var parts []model.AudioPart
	query := db.WithContext(ctx).Model(&model.AudioPart{}).
		Select("audio_parts.*").
		Joins("JOIN videos ON videos.id = audio_parts.video_id AND videos.deleted_at IS NULL").
		Order("audio_parts.id")
	if videoID > 0 {
		query = query.Where("audio_parts.video_id = ?", videoID)
	}
	if err := query.Find(&parts).Error; err != nil {
		return 0, err
	}
	noted := 0
	updated := dbtime.DateTime{Time: now}
	for _, part := range parts {
		stored, mod, ok := existingAudioFile(dataDir, part)
		if !ok {
			continue
		}
		downloaded := dbtime.DateTime{Time: mod}
		err := db.WithContext(ctx).Model(&model.AudioPart{}).Where("id = ?", part.ID).Updates(map[string]any{
			"audio_download_path": stored,
			"audio_downloaded_at": downloaded,
			"updated_at":          updated,
		}).Error
		if err != nil {
			return noted, err
		}
		noted++
	}
	return noted, nil
}

// existingVideoFile 在已存路径的文件不在时，找出还能播放的那个文件。
// 顺序与详情页找文件一致：旧的按页码命名，再按 cid 命名，最后是 视频ID.cid.mp4。
// 已存路径的文件还在时返回 false，调用方不改路径。
func existingVideoFile(dataDir string, part model.VideoPart) (stored string, mod time.Time, ok bool) {
	if storedFileReady(dataDir, part.VideoDownloadPath) {
		return "", time.Time{}, false
	}
	type candidate struct {
		disk   string
		stored string
	}
	page := int64(part.Page)
	list := []candidate{{
		disk:   resolveStored(dataDir, legacyVideoPath(part.VideoID, page)),
		stored: legacyVideoPath(part.VideoID, page),
	}}
	if part.Cid > 0 && part.Cid != page {
		legacy := legacyVideoPath(part.VideoID, part.Cid)
		list = append(list, candidate{
			disk:   resolveStored(dataDir, legacy),
			stored: legacy,
		})
	}
	if part.Cid > 0 {
		list = append(list, candidate{
			disk:   videoFilePath(dataDir, part.VideoID, part.Cid),
			stored: storageVideoPath(part.VideoID, part.Cid),
		})
	}
	for _, item := range list {
		info, exists := nonemptyFile(item.disk)
		if !exists {
			continue
		}
		return item.stored, info.ModTime(), true
	}
	return "", time.Time{}, false
}

// existingAudioFile 在已存路径的文件不在时，先找 m4a，再找旧的 mp3。
func existingAudioFile(dataDir string, part model.AudioPart) (stored string, mod time.Time, ok bool) {
	if part.Sid <= 0 || storedFileReady(dataDir, part.AudioDownloadPath) {
		return "", time.Time{}, false
	}
	list := []struct {
		disk   string
		stored string
	}{
		{disk: audioFilePath(dataDir, part.Sid), stored: storageAudioPath(part.Sid)},
		{disk: legacyAudioFilePath(dataDir, part.Sid), stored: legacyStorageAudioPath(part.Sid)},
	}
	for _, item := range list {
		info, exists := nonemptyFile(item.disk)
		if !exists {
			continue
		}
		return item.stored, info.ModTime(), true
	}
	return "", time.Time{}, false
}

func nonemptyFile(path string) (os.FileInfo, bool) {
	if path == "" {
		return nil, false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() == 0 {
		return nil, false
	}
	return info, true
}

func recountDownloads(ctx context.Context, db *gorm.DB, videoID int64, now time.Time) (int, error) {
	var videos []model.Video
	query := db.WithContext(ctx).Model(&model.Video{}).
		Select("id, video_downloaded_num, audio_downloaded_num, video_downloaded_at").
		Where("deleted_at IS NULL")
	if videoID > 0 {
		query = query.Where("id = ?", videoID)
	}
	if err := query.Find(&videos).Error; err != nil {
		return 0, err
	}
	videoCounts, err := pathCounts(ctx, db, videoID, "video_parts", "video_download_path")
	if err != nil {
		return 0, err
	}
	audioCounts, err := pathCounts(ctx, db, videoID, "audio_parts", "audio_download_path")
	if err != nil {
		return 0, err
	}
	stamps, err := latestVideoDownloads(ctx, db, videoID)
	if err != nil {
		return 0, err
	}
	updated := dbtime.DateTime{Time: now}
	changed := 0
	for _, video := range videos {
		videoNum := videoCounts[video.ID]
		audioNum := audioCounts[video.ID]
		if video.VideoDownloadedNum == videoNum && video.AudioDownloadedNum == audioNum {
			continue
		}
		fields := map[string]any{
			"video_downloaded_num": videoNum,
			"audio_downloaded_num": audioNum,
			"updated_at":           updated,
		}
		if videoNum > 0 && video.VideoDownloadedAt == nil {
			if stamp, ok := stamps[video.ID]; ok {
				fields["video_downloaded_at"] = stamp
			}
		}
		if err := db.WithContext(ctx).Model(&model.Video{}).Where("id = ?", video.ID).Updates(fields).Error; err != nil {
			return changed, err
		}
		changed++
	}
	return changed, nil
}

func pathCounts(ctx context.Context, db *gorm.DB, videoID int64, table, column string) (map[int64]int, error) {
	var rows []struct {
		VideoID int64 `gorm:"column:video_id"`
		N       int64 `gorm:"column:n"`
	}
	query := db.WithContext(ctx).Table(table).
		Select("video_id, COUNT(*) AS n").
		Where(column + " IS NOT NULL AND " + column + " <> ''").
		Group("video_id")
	if videoID > 0 {
		query = query.Where("video_id = ?", videoID)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[int64]int, len(rows))
	for _, row := range rows {
		counts[row.VideoID] = int(row.N)
	}
	return counts, nil
}

func latestVideoDownloads(ctx context.Context, db *gorm.DB, videoID int64) (map[int64]dbtime.DateTime, error) {
	var rows []struct {
		VideoID int64            `gorm:"column:video_id"`
		At      *dbtime.DateTime `gorm:"column:downloaded_at"`
	}
	query := db.WithContext(ctx).Model(&model.VideoPart{}).
		Select("video_id, MAX(video_downloaded_at) AS downloaded_at").
		Where("video_download_path IS NOT NULL AND video_download_path <> ''").
		Group("video_id")
	if videoID > 0 {
		query = query.Where("video_id = ?", videoID)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	stamps := make(map[int64]dbtime.DateTime, len(rows))
	for _, row := range rows {
		if row.At == nil || row.At.IsZero() {
			continue
		}
		stamps[row.VideoID] = *row.At
	}
	return stamps, nil
}
