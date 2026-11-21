package model

import "mybili/internal/dbtime"

const (
	CoverableVideo        = `App\Models\Video`
	CoverableFavoriteList = `App\Models\FavoriteList`
	CoverableSubscription = `App\Models\Subscription`
	CoverableUpper        = `App\Models\Upper`
)

type Cover struct {
	ID                   int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	Type                 string           `gorm:"column:type;type:varchar;not null" json:"type"`
	URL                  string           `gorm:"column:url;type:varchar;not null;index" json:"url"`
	Filename             string           `gorm:"column:filename;type:varchar;not null;uniqueIndex" json:"filename"`
	Path                 string           `gorm:"column:path;type:varchar;not null" json:"path"`
	MimeType             string           `gorm:"column:mime_type;type:varchar;not null;default:''" json:"mime_type"`
	Size                 int              `gorm:"column:size;not null;default:0" json:"size"`
	Width                int              `gorm:"column:width;not null;default:0" json:"width"`
	Height               int              `gorm:"column:height;not null;default:0" json:"height"`
	ThumbnailGeneratedAt *dbtime.DateTime `gorm:"column:thumbnail_generated_at;type:datetime" json:"thumbnail_generated_at"`
	CreatedAt            *dbtime.DateTime `gorm:"column:created_at;type:datetime" json:"created_at"`
}

func (Cover) TableName() string { return "covers" }

type Coverable struct {
	ID            int64            `gorm:"primaryKey;autoIncrement"`
	CoverID       int64            `gorm:"column:cover_id;not null;index"`
	CoverableID   int64            `gorm:"column:coverable_id;not null;uniqueIndex:coverable_owner"`
	CoverableType string           `gorm:"column:coverable_type;type:varchar;not null;uniqueIndex:coverable_owner"`
	CreatedAt     *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt     *dbtime.DateTime `gorm:"type:datetime"`
}

func (Coverable) TableName() string { return "coverables" }
