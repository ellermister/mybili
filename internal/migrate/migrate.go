package migrate

import (
	"fmt"

	"gorm.io/gorm"

	"mybili/internal/model"
)

// Auto 用 GORM AutoMigrate 建表、补列、改已有列。
// 任务表叫 tasks，避开 Laravel 的 jobs 队列表。
func Auto(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Video{},
		&model.VideoPart{},
		&model.AudioPart{},
		&model.Danmaku{},
		&model.FavoriteList{},
		&model.FavoriteListVideo{},
		&model.Subscription{},
		&model.SubscriptionVideo{},
		&model.Upper{},
		&model.Cover{},
		&model.Coverable{},
		&model.Setting{},
		&model.Job{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	// 建索引相关额外写为SQL语句
	// 避免SQLite 的 AutoMigrate 一对不上列定义就会整表重建, 速度远不如直接写SQL语句。
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS danmaku_video_id_index ON danmaku(video_id)`).Error; err != nil {
		return fmt.Errorf("danmaku video index: %w", err)
	}
	return nil
}
