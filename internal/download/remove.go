package download

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

// DeleteOptions 控制删的是记录还是只删文件。Permanent 为 true 时软删除并清弹幕。
// Requeue 只在临时删除时生效，会把下载任务重新打开。
type DeleteOptions struct {
	Permanent bool
	Requeue   bool
}

// DeleteVideos 删除一组视频的本地文件。找不到的 ID 会跳过。
// 永久删除保留封面、收藏和订阅关系。临时删除保留视频行和弹幕。
func DeleteVideos(ctx context.Context, db *gorm.DB, dataDir string, ids []int64, opt DeleteOptions, now time.Time, wake func(context.Context, string) error) ([]int64, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}
	seen := map[int64]bool{}
	deleted := make([]int64, 0, len(ids))
	stamp := dbtime.DateTime{Time: now}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		var video model.Video
		err := db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Limit(1).Find(&video).Error
		if err != nil {
			return deleted, err
		}
		if video.ID == 0 {
			continue
		}
		if err := removeVideoFiles(ctx, db, dataDir, video, stamp); err != nil {
			return deleted, err
		}
		if opt.Permanent {
			if err := db.WithContext(ctx).Model(&video).Updates(map[string]any{
				"deleted_at": stamp,
				"updated_at": stamp,
			}).Error; err != nil {
				return deleted, err
			}
			if err := db.WithContext(ctx).Where("video_id = ?", video.ID).Delete(&model.Danmaku{}).Error; err != nil {
				return deleted, err
			}
			deleted = append(deleted, video.ID)
			continue
		}
		if err := db.WithContext(ctx).Model(&video).Updates(map[string]any{
			"video_downloaded_num": 0,
			"audio_downloaded_num": 0,
			"video_downloaded_at":  nil,
			"updated_at":           stamp,
		}).Error; err != nil {
			return deleted, err
		}
		if opt.Requeue {
			if video.IsAudio() {
				_, err = RequeueAudio(ctx, db, dataDir, video.ID, now, wake)
			} else {
				_, err = RequeueParts(ctx, db, dataDir, video.ID, now, wake)
			}
			if err != nil {
				return deleted, err
			}
		}
		deleted = append(deleted, video.ID)
	}
	return deleted, nil
}

// removeVideoFiles 删掉分 P 和音频文件，并清掉下载路径和时间。
func removeVideoFiles(ctx context.Context, db *gorm.DB, dataDir string, video model.Video, stamp dbtime.DateTime) error {
	var parts []model.VideoPart
	if err := db.WithContext(ctx).Where("video_id = ?", video.ID).Find(&parts).Error; err != nil {
		return err
	}
	for _, part := range parts {
		if part.VideoDownloadPath != nil {
			removeUnderPublic(dataDir, *part.VideoDownloadPath)
		}
		removeUnderPublic(dataDir, storageVideoPath(video.ID, part.Cid))
		removeUnderPublic(dataDir, legacyVideoPath(video.ID, int64(part.Page)))
		if int64(part.Page) != part.Cid {
			removeUnderPublic(dataDir, legacyVideoPath(video.ID, part.Cid))
		}
		if err := db.WithContext(ctx).Model(&part).Updates(map[string]any{
			"video_download_path": nil,
			"video_downloaded_at": nil,
			"updated_at":          stamp,
		}).Error; err != nil {
			return err
		}
	}
	var audio model.AudioPart
	err := db.WithContext(ctx).Where("video_id = ?", video.ID).Limit(1).Find(&audio).Error
	if err != nil {
		return err
	}
	if audio.ID == 0 {
		return nil
	}
	if audio.AudioDownloadPath != nil {
		removeUnderPublic(dataDir, *audio.AudioDownloadPath)
	}
	removeUnderPublic(dataDir, storageAudioPath(audio.Sid))
	removeUnderPublic(dataDir, legacyStorageAudioPath(audio.Sid))
	return db.WithContext(ctx).Model(&audio).Updates(map[string]any{
		"audio_download_path": nil,
		"audio_downloaded_at": nil,
		"updated_at":          stamp,
	}).Error
}

// legacyVideoPath 是旧程序按页码或 cid 拼出来的文件。第一段不带 part。
func legacyVideoPath(videoID, part int64) string {
	id := strconv.FormatInt(videoID, 10)
	if part <= 1 {
		return "/storage/videos/" + id + ".mp4"
	}
	return "/storage/videos/" + id + ".part" + strconv.FormatInt(part, 10) + ".mp4"
}

// removeUnderPublic 删除 public 目录里的文件和旁边的 .hash。目录外的路径不动。
func removeUnderPublic(dataDir, stored string) {
	path := resolveStored(dataDir, stored)
	root := filepath.Join(dataDir, "public")
	if !insideDir(root, path) {
		return
	}
	_ = os.Remove(path)
	_ = os.Remove(path + ".hash")
}

func insideDir(root, path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
