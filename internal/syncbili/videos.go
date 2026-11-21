package syncbili

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

// VideoSyncResult 是一次收藏夹视频同步写库的数量。
type VideoSyncResult struct {
	Folders int
	Videos  int
}

// FavoriteVideos 同步收藏夹里的视频，只写 videos 和 favorite_list_videos。不在这里拉分 P。
// 这一页里还没有分 P 的普通视频会写入稿件详情任务，由队列再去拉分 P 并下载。
// favID 为 0 表示本地已有的全部收藏夹。page 为 0 表示全部页，大于 0 时只拉那一页，并且不解开收藏关系。
func (s Sync) FavoriteVideos(ctx context.Context, favID int64, page int) (VideoSyncResult, error) {
	if favID < 0 {
		return VideoSyncResult{}, fmt.Errorf("fav must be >= 1")
	}
	if page < 0 {
		return VideoSyncResult{}, fmt.Errorf("page must be >= 1")
	}
	now := s.now()
	if until, active, err := suspendUntil(s.DB, now); err != nil {
		return VideoSyncResult{}, err
	} else if active {
		return VideoSyncResult{}, &SuspendedError{Until: until}
	}
	folders, err := s.folders(favID)
	if err != nil {
		return VideoSyncResult{}, err
	}
	var result VideoSyncResult
	for i, folder := range folders {
		if i > 0 {
			if err := waitInterval(ctx, s.PageInterval); err != nil {
				return result, err
			}
		}
		count, err := s.syncFolderVideos(ctx, folder, page, now)
		if err != nil {
			if bilibili.IsRiskControl(err) {
				return result, s.suspend(now)
			}
			return result, fmt.Errorf("favorite %d: %w", folder.ID, err)
		}
		result.Folders++
		result.Videos += count
	}
	return result, nil
}

// folders 取出要同步的本地收藏夹。favID 大于 0 时只取那一条，本地没有就报错。
func (s Sync) folders(favID int64) ([]model.FavoriteList, error) {
	if favID > 0 {
		var folder model.FavoriteList
		err := s.DB.First(&folder, favID).Error
		if err != nil {
			return nil, fmt.Errorf("favorite %d is not in the local list, run sync-fav first", favID)
		}
		return []model.FavoriteList{folder}, nil
	}
	var folders []model.FavoriteList
	if err := s.DB.Order("id").Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

// syncFolderVideos 拉取一个收藏夹的视频并写库。page 为 0 且条数对得上时，解开多余的收藏关系。
func (s Sync) syncFolderVideos(ctx context.Context, folder model.FavoriteList, page int, now time.Time) (int, error) {
	medias, err := bilibili.ListFavoriteMedias(ctx, s.API, folder.ID, s.PageSize, page, s.PageInterval)
	if err != nil {
		return 0, err
	}
	existing, err := videosByID(s.DB, mediaIDs(medias))
	if err != nil {
		return 0, err
	}
	var kept []model.Video
	keptIDs := map[int64]struct{}{}
	for _, media := range medias {
		current, ok := existing[media.ID]
		var currentPtr *model.Video
		if ok {
			currentPtr = &current
		}
		row, include := prepareVideo(currentPtr, media, now)
		if !include {
			continue
		}
		kept = append(kept, row)
		keptIDs[row.ID] = struct{}{}
	}
	if len(kept) == 0 {
		return 0, nil
	}
	// 读出已经关联的视频 ID
	links, err := favoriteVideoIDs(s.DB, folder.ID)
	if err != nil {
		return 0, err
	}
	var added []int64
	var removed []int64
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		// 将远端过滤后的保留数据写入数据库
		for _, row := range kept {
			if err := tx.Save(&row).Error; err != nil {
				return err
			}
			// 如果视频已存在，则跳过
			if _, linked := links[row.ID]; linked {
				continue
			}

			// 若 kept 中存在的视频没有在本地关联，则将视频与收藏夹关联
			link := model.FavoriteListVideo{
				FavoriteListID: folder.ID,
				VideoID:        row.ID,
				CreatedAt:      row.FavTime,
				UpdatedAt:      row.FavTime,
			}
			if err := tx.Create(&link).Error; err != nil {
				return err
			}
			added = append(added, row.ID)
		}
		if page != 0 || folder.MediaCount != len(kept) {
			return nil
		}

		// 若 page 为 0 且条数对不上，则删除本地多余的关联
		// 遍历本地已关联的视频 ID，若不在 keptIDs 中，则删除关系，视频本身保留
		var remove []int64
		for id := range links {
			if _, ok := keptIDs[id]; !ok {
				remove = append(remove, id)
			}
		}
		for start := 0; start < len(remove); start += 500 {
			end := start + 500
			if end > len(remove) {
				end = len(remove)
			}
			err := tx.Where("favorite_list_id = ? AND video_id IN ?", folder.ID, remove[start:end]).Delete(&model.FavoriteListVideo{}).Error
			if err != nil {
				return err
			}
		}
		removed = remove
		return nil
	})
	if err != nil {
		return 0, err
	}
	if len(added) > 0 && s.AfterLinked != nil {
		s.AfterLinked(ctx, added)
	}
	if len(removed) > 0 && s.AfterUnlinked != nil {
		s.AfterUnlinked(ctx, folder.Title, removed)
	}
	if err := s.enqueueMissingPartDetails(ctx, folder.ID, kept, now); err != nil {
		return len(kept), err
	}

	// 遍历保留的视频，若无效则跳过，否则入队封面检查下载
	for _, row := range kept {
		if row.Invalid != 0 {
			continue
		}
		if _, err := s.EnqueueCover(nil, CoverKindVideo, row.ID, row.Cover); err != nil {
			return len(kept), err
		}
		if row.IsAudio() {
			if err := EnqueueAudioDetail(ctx, s.DB, row.ID, now, s.Wake); err != nil {
				return len(kept), err
			}
		}
	}
	return len(kept), nil
}

// enqueueMissingPartDetails 给这一页还没有分 P 的视频写入稿件详情任务。
// 已有分 P、音频、无效、没有 bvid，或当前收藏夹被排除时跳过。
// 同一 bvid 已有未完成任务时不会再插一条。任务行被清空后，下次扫到这一页会重新写入。
func (s Sync) enqueueMissingPartDetails(ctx context.Context, folderID int64, kept []model.Video, now time.Time) error {
	if s.DB == nil || len(kept) == 0 {
		return nil
	}
	excluded, err := excludedFavoriteIDs(s.DB)
	if err != nil || excluded[folderID] {
		return err
	}
	ids := make([]int64, 0, len(kept))
	bvidByID := make(map[int64]string, len(kept))
	for _, row := range kept {
		bvid := strings.TrimSpace(row.Bvid)
		if row.Invalid != 0 || row.IsAudio() || bvid == "" {
			continue
		}
		ids = append(ids, row.ID)
		bvidByID[row.ID] = bvid
	}
	if len(ids) == 0 {
		return nil
	}
	var have []int64
	if err := s.DB.WithContext(ctx).Model(&model.VideoPart{}).
		Where("video_id IN ?", ids).
		Pluck("video_id", &have).Error; err != nil {
		return err
	}
	got := make(map[int64]struct{}, len(have))
	for _, id := range have {
		got[id] = struct{}{}
	}
	for _, id := range ids {
		if _, ok := got[id]; ok {
			continue
		}
		if err := EnqueueVideoDetail(ctx, s.DB, bvidByID[id], now, s.Wake); err != nil {
			return err
		}
	}
	return nil
}

// prepareVideo 把远端视频收成一行。已删除、两端都无效、已下载且远端无效的不写入。
func prepareVideo(existing *model.Video, media bilibili.FavoriteMedia, now time.Time) (model.Video, bool) {
	remoteInvalid := media.Attr > 0 || media.Title == "已失效视频"
	// 本地手动删除的不写入
	if existing != nil && existing.DeletedAt != nil && !existing.DeletedAt.Time.IsZero() {
		return model.Video{}, false
	}
	// 本地无效且远端无效的不写入
	if existing != nil && existing.Invalid != 0 && remoteInvalid {
		return model.Video{}, false
	}
	// 本地下载了且远端无效的不写入
	downloaded := existing != nil && (existing.VideoDownloadedNum > 0 || existing.AudioDownloadedNum > 0)
	if downloaded && remoteInvalid {
		return model.Video{}, false
	}

	row := model.Video{ID: media.ID, CacheImage: ""}
	if existing != nil {
		row = *existing
	} else {
		created := dbtime.DateTime{Time: now}
		row.CreatedAt = &created
	}
	row.Link = media.Link
	if row.Link == "" && media.Bvid != "" {
		row.Link = "https://www.bilibili.com/video/" + media.Bvid
	}
	row.Title = media.Title
	row.Intro = media.Intro
	row.Cover = media.Cover
	row.Bvid = media.Bvid
	row.Pubtime = unixDateTime(media.Pubtime)
	row.Duration = media.Duration
	row.Attr = media.Attr
	row.Invalid = 0
	if remoteInvalid {
		row.Invalid = 1
	}
	row.Frozen = 0
	row.Page = media.Page
	favTime := unixDateTime(media.FavTime)
	row.FavTime = &favTime
	if media.Upper != nil {
		if media.Upper.Mid == 0 {
			row.UpperID = nil
		} else {
			mid := media.Upper.Mid
			row.UpperID = &mid
		}
	}
	row.Type = media.Type
	updated := dbtime.DateTime{Time: now}
	row.UpdatedAt = &updated
	return row, true
}

// videosByID 按 ID 分批读出已有视频。
func videosByID(db *gorm.DB, ids []int64) (map[int64]model.Video, error) {
	found := map[int64]model.Video{}
	for start := 0; start < len(ids); start += 500 {
		end := start + 500
		if end > len(ids) {
			end = len(ids)
		}
		var rows []model.Video
		if err := db.Where("id IN ?", ids[start:end]).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			found[row.ID] = row
		}
	}
	return found, nil
}

// favoriteVideoIDs 读出一个收藏夹已经关联的视频 ID。
func favoriteVideoIDs(db *gorm.DB, favID int64) (map[int64]struct{}, error) {
	var links []model.FavoriteListVideo
	if err := db.Where("favorite_list_id = ?", favID).Find(&links).Error; err != nil {
		return nil, err
	}
	ids := make(map[int64]struct{}, len(links))
	for _, link := range links {
		ids[link.VideoID] = struct{}{}
	}
	return ids, nil
}

// mediaIDs 取出这一页远端视频的 ID。
func mediaIDs(medias []bilibili.FavoriteMedia) []int64 {
	ids := make([]int64, 0, len(medias))
	for _, media := range medias {
		ids = append(ids, media.ID)
	}
	return ids
}
