package model

import "mybili/internal/dbtime"

type Video struct {
	ID                 int64            `gorm:"primaryKey;autoIncrement"`
	Link               string           `gorm:"type:varchar;not null"`
	Title              string           `gorm:"type:varchar;not null"`
	Intro              string           `gorm:"type:text;not null"`
	Cover              string           `gorm:"type:varchar;not null"`
	Bvid               string           `gorm:"type:varchar;not null"`
	Pubtime            dbtime.DateTime  `gorm:"type:datetime;not null"`
	Duration           int              `gorm:"not null;default:0"`
	Attr               int              `gorm:"not null;default:0"`
	Invalid            int              `gorm:"not null;default:0;index:idx_videos_invalid"` // 视频是否无效，同步线上状态
	Frozen             int              `gorm:"not null;default:0;index:idx_videos_frozen"`  // 视频是否冻结，只有本地下载了视频或者音频时，并且线上位无效状态时才标记冻结
	CacheImage         string           `gorm:"type:varchar;not null"`
	Page               int              `gorm:"not null;default:1"`
	FavTime            *dbtime.DateTime `gorm:"type:datetime;index:idx_videos_fav_time"`
	CreatedAt          *dbtime.DateTime `gorm:"type:datetime;index:idx_videos_created_at"`
	UpdatedAt          *dbtime.DateTime `gorm:"type:datetime"`
	VideoDownloadedAt  *dbtime.DateTime `gorm:"type:datetime"`
	VideoDownloadedNum int              `gorm:"not null;default:0;index:idx_videos_downloaded_num"`
	AudioDownloadedNum int              `gorm:"not null;default:0"`
	DeletedAt          *dbtime.DateTime `gorm:"type:datetime;index"`
	UpperID            *int64           `gorm:"column:upper_id"`
	Type               int              `gorm:"not null;default:2"`
}

func (Video) TableName() string { return "videos" }

func (v Video) IsAudio() bool { return v.Type == 12 }

type VideoPart struct {
	ID                  int64            `gorm:"primaryKey;autoIncrement"`
	VideoID             int64            `gorm:"not null;index"`
	Cid                 int64            `gorm:"not null;index"`
	Page                int              `gorm:"not null;default:1"`
	From                string           `gorm:"type:varchar;not null"`
	Part                string           `gorm:"type:varchar;not null"`
	Duration            int              `gorm:"not null"`
	Vid                 string           `gorm:"type:varchar;not null"`
	Weblink             string           `gorm:"type:varchar;not null"`
	Width               int              `gorm:"not null;default:0"`
	Height              int              `gorm:"not null;default:0"`
	Rotate              int              `gorm:"not null;default:0"`
	FirstFrame          string           `gorm:"type:varchar;not null;default:''"`
	VideoDownloadPath   *string          `gorm:"type:varchar"`
	VideoDownloadedAt   *dbtime.DateTime `gorm:"type:datetime"`
	DanmakuDownloadedAt *dbtime.DateTime `gorm:"type:datetime"`
	CreatedAt           *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt           *dbtime.DateTime `gorm:"type:datetime"`
}

func (VideoPart) TableName() string { return "video_parts" }

type AudioPart struct {
	ID                int64            `gorm:"primaryKey;autoIncrement"`
	VideoID           int64            `gorm:"not null;uniqueIndex"`
	Sid               int64            `gorm:"not null;index"`
	Duration          int              `gorm:"not null;default:0"`
	AudioDownloadedAt *dbtime.DateTime `gorm:"type:datetime"`
	AudioDownloadPath *string          `gorm:"type:varchar"`
	CreatedAt         *dbtime.DateTime `gorm:"type:datetime"`
	UpdatedAt         *dbtime.DateTime `gorm:"type:datetime"`
}

func (AudioPart) TableName() string { return "audio_parts" }

type Danmaku struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	VideoID   int64 `gorm:"not null"`
	Cid       int64 `gorm:"not null;index:danmaku_cid_index"`
	Progress  int   `gorm:"not null;default:0"`
	Mode      int   `gorm:"not null;default:1"`
	Color     *int
	Content   string          `gorm:"type:varchar;not null"`
	CreatedAt dbtime.DateTime `gorm:"type:datetime;not null"`
}

func (Danmaku) TableName() string { return "danmaku" }
