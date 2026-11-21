package syncbili

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

// Folders 同步「我创建的收藏夹」列表。
// 只写 favorite_lists 的标题、封面、时间、数量。不拉视频。封面有变化时入队。
func (s Sync) Folders(ctx context.Context) (int, error) {
	now := s.now()
	if until, active, err := suspendUntil(s.DB, now); err != nil {
		return 0, err
	} else if active {
		return 0, &SuspendedError{Until: until}
	}
	folders, err := bilibili.ListCreatedFolders(ctx, s.API, s.Cookies, s.PageInterval)
	if err != nil {
		if bilibili.IsRiskControl(err) {
			return 0, s.suspend(now)
		}
		return 0, err
	}
	if err := upsertFolders(s.DB, folders, now); err != nil {
		return 0, err
	}
	for _, folder := range folders {
		if _, err := s.EnqueueCover(nil, CoverKindFavorite, folder.ID, folder.Cover); err != nil {
			return len(folders), err
		}
	}
	return len(folders), nil
}

// upsertFolders 按收藏夹 ID 写入或更新 favorite_lists。接口没返回的行不删除。
func upsertFolders(db *gorm.DB, folders []bilibili.Folder, now time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, folder := range folders {
			var row model.FavoriteList
			err := tx.First(&row, folder.ID).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				row = model.FavoriteList{ID: folder.ID, CacheImage: ""}
				created := dbtime.DateTime{Time: now}
				row.CreatedAt = &created
			} else if err != nil {
				return err
			}
			row.Title = folder.Title
			row.Cover = folder.Cover
			row.Ctime = dbtime.DateTime{Time: time.Unix(folder.Ctime, 0).In(dbtime.Location())}
			row.Mtime = dbtime.DateTime{Time: time.Unix(folder.Mtime, 0).In(dbtime.Location())}
			row.MediaCount = folder.MediaCount
			updated := dbtime.DateTime{Time: now}
			row.UpdatedAt = &updated
			if err := tx.Save(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
