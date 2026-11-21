package model

import "mybili/internal/dbtime"

type FavoriteList struct {
	ID         int64            `gorm:"primaryKey;autoIncrement"`
	Title      string           `gorm:"type:varchar;not null"`
	Cover      string           `gorm:"type:varchar;not null"`
	Ctime      dbtime.DateTime  `gorm:"type:datetime;not null"`
	Mtime      dbtime.DateTime  `gorm:"type:datetime;not null"`
	MediaCount int              `gorm:"not null;default:0"`
	CacheImage string           `gorm:"type:varchar;not null;default:''"`
	CreatedAt  *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt  *dbtime.DateTime `gorm:"type:datetime"`
}

func (FavoriteList) TableName() string { return "favorite_lists" }

type FavoriteListVideo struct {
	FavoriteListID int64            `gorm:"primaryKey"`
	VideoID        int64            `gorm:"primaryKey"`
	CreatedAt      *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt      *dbtime.DateTime `gorm:"type:datetime"`
}

func (FavoriteListVideo) TableName() string { return "favorite_list_videos" }

type Subscription struct {
	ID          int64            `gorm:"primaryKey;autoIncrement"`
	Type        string           `gorm:"type:varchar;not null"`
	Mid         string           `gorm:"type:varchar;not null"`
	ListID      *string          `gorm:"column:list_id;type:varchar"`
	Name        string           `gorm:"type:varchar;not null;default:''"`
	URL         string           `gorm:"type:varchar;not null;default:''"`
	Cover       string           `gorm:"type:varchar;not null;default:''"`
	CacheImage  string           `gorm:"type:varchar;not null;default:''"`
	Description string           `gorm:"type:varchar;not null;default:''"`
	Total       string           `gorm:"type:varchar;not null;default:0"`
	Status      int              `gorm:"not null;default:1"`
	LastCheckAt *dbtime.DateTime `gorm:"type:datetime"`
	CreatedAt   *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt   *dbtime.DateTime `gorm:"type:datetime"`
}

func (Subscription) TableName() string { return "subscriptions" }

type SubscriptionVideo struct {
	SubscriptionID int64            `gorm:"primaryKey"`
	VideoID        int64            `gorm:"primaryKey"`
	Bvid           string           `gorm:"type:varchar;not null"`
	CreatedAt      *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt      *dbtime.DateTime `gorm:"type:datetime"`
}

func (SubscriptionVideo) TableName() string { return "subscription_videos" }

type Upper struct {
	ID        int64            `gorm:"primaryKey;autoIncrement"`
	Mid       int64            `gorm:"not null;uniqueIndex"`
	Name      string           `gorm:"type:varchar;not null"`
	Face      string           `gorm:"type:varchar;not null"`
	CreatedAt *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt *dbtime.DateTime `gorm:"type:datetime"`
}

func (Upper) TableName() string { return "upper" }

type Setting struct {
	ID        int64            `gorm:"primaryKey;autoIncrement"`
	Name      string           `gorm:"type:varchar;not null;uniqueIndex"`
	Value     *string          `gorm:"type:varchar"`
	CreatedAt *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt *dbtime.DateTime `gorm:"type:datetime"`
}

func (Setting) TableName() string { return "settings" }

type Job struct {
	ID         int64            `gorm:"primaryKey;autoIncrement"`
	Queue      string           `gorm:"type:varchar;not null;default:fast;index"`
	Type       string           `gorm:"type:varchar;not null;default:''"`
	UniqueKey  string           `gorm:"type:varchar;not null;default:''"`
	Status     string           `gorm:"type:varchar;not null;default:pending;index"`
	Priority   int              `gorm:"not null;default:0"`
	Payload    string           `gorm:"type:text;not null;default:''"`
	ErrorMsg   *string          `gorm:"type:text"`
	RetryCount int              `gorm:"not null;default:0"`
	MaxRetries int              `gorm:"not null;default:3"`
	RunAt      *dbtime.DateTime `gorm:"type:datetime"`
	StartedAt  *dbtime.DateTime `gorm:"type:datetime"`
	FinishedAt *dbtime.DateTime `gorm:"type:datetime"`
	CreatedAt  *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt  *dbtime.DateTime `gorm:"type:datetime"`
}

func (Job) TableName() string { return "tasks" }
