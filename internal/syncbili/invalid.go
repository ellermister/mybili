package syncbili

import (
	"context"
	"fmt"
	"os"
	"strings"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

// 收藏夹接口在稿件失效后常用的占位封面。这种封面不能拿来替换本地图片。
const invalidCoverMark = "be27fd62c99036dce67efface486fb0a88ffed06"

// FixInvalidFavorite 用移动端收藏夹接口修正一页失效稿件的标题、封面和无效标记。
// 只改本地已经有的视频。标题或封面已经是失效占位时跳过。
func (s Sync) FixInvalidFavorite(ctx context.Context, favID int64, page int) (int, error) {
	if s.DB == nil || s.API == nil {
		return 0, fmt.Errorf("syncer is incomplete")
	}
	if favID <= 0 {
		return 0, fmt.Errorf("fav must be >= 1")
	}
	if page <= 0 {
		return 0, fmt.Errorf("page must be >= 1")
	}
	now := s.now()
	if until, active, err := suspendUntil(s.DB, now); err != nil {
		return 0, err
	} else if active {
		return 0, &SuspendedError{Until: until}
	}
	pageSize := s.PageSize
	if pageSize < 1 {
		pageSize = 40
	}
	items, err := bilibili.ListFolderResources(ctx, s.API, favID, page, pageSize)
	if err != nil {
		if bilibili.IsRiskControl(err) {
			return 0, s.suspend(now)
		}
		return 0, err
	}
	stamp := dbtime.DateTime{Time: now}
	fixed := 0
	for _, item := range items {
		if !item.IsInvalid || item.OID <= 0 {
			continue
		}
		if !usableCover(item.Title, item.Cover) {
			continue
		}
		var video model.Video
		if err := s.DB.WithContext(ctx).Where("id = ?", item.OID).Limit(1).Find(&video).Error; err != nil {
			return fixed, err
		}
		if video.ID == 0 {
			continue
		}
		oldCover := video.Cover
		if err := s.DB.WithContext(ctx).Model(&video).Updates(map[string]any{
			"invalid":    1,
			"title":      item.Title,
			"cover":      item.Cover,
			"updated_at": stamp,
		}).Error; err != nil {
			return fixed, err
		}
		fixed++
		// 地址变了，或者 cache_image 指着的文件不在，都要重新下载。
		if item.Cover != "" && (item.Cover != oldCover || !s.localImageReady(video.CacheImage)) {
			if _, err := s.EnqueueCover(nil, CoverKindVideo, video.ID, item.Cover); err != nil {
				return fixed, err
			}
		}
	}
	// 失效稿件有时不再出现在接口列表里，已经标过无效的也要补封面。
	if err := s.repairMissingInvalidCovers(ctx, favID); err != nil {
		return fixed, err
	}
	return fixed, nil
}

// repairMissingInvalidCovers 给这个收藏夹里已标记无效、封面文件却不在的视频补下载。
func (s Sync) repairMissingInvalidCovers(ctx context.Context, favID int64) error {
	var videos []model.Video
	err := s.DB.WithContext(ctx).
		Joins("JOIN favorite_list_videos ON favorite_list_videos.video_id = videos.id AND favorite_list_videos.favorite_list_id = ?", favID).
		Where("videos.invalid = ? AND videos.deleted_at IS NULL", 1).
		Find(&videos).Error
	if err != nil {
		return err
	}
	for _, video := range videos {
		if !usableCover(video.Title, video.Cover) || s.localImageReady(video.CacheImage) {
			continue
		}
		if _, err := s.EnqueueCover(nil, CoverKindVideo, video.ID, video.Cover); err != nil {
			return err
		}
	}
	return nil
}

// usableCover 拒绝空地址和失效占位图。
func usableCover(title, cover string) bool {
	if strings.TrimSpace(cover) == "" || title == "已失效视频" || strings.Contains(cover, invalidCoverMark) {
		return false
	}
	return true
}

// localImageReady 判断 cache_image 记下的文件是否还在，并且不是空文件。字段为空算没有。
func (s Sync) localImageReady(rel string) bool {
	rel = strings.TrimSpace(rel)
	rel = strings.TrimPrefix(rel, "/storage/")
	rel = strings.TrimPrefix(rel, "storage/")
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || strings.Contains(rel, "..") {
		return false
	}
	info, err := os.Stat(s.dataDir() + "/public/" + rel)
	return err == nil && !info.IsDir() && info.Size() > 0
}
