package syncbili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/download"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/taskrun"
)

// TypeSyncAudio 是收藏夹音频稿件的详情和下载入队。
const TypeSyncAudio = "sync-audio"

// EnqueueAudioDetail 把一条音频详情写成 pending。已经跑完的不重新打开。
func EnqueueAudioDetail(ctx context.Context, db *gorm.DB, videoID int64, now time.Time, wake func(context.Context, string) error) error {
	if db == nil || videoID <= 0 {
		return nil
	}
	body, err := json.Marshal(map[string]int64{"id": videoID})
	if err != nil {
		return err
	}
	_, err = jobs.Store{DB: db}.Enqueue(ctx, jobs.Draft{
		Queue:      taskrun.QueueBiliFast,
		Type:       TypeSyncAudio,
		UniqueKey:  fmt.Sprintf("sync:audio:%d", videoID),
		Payload:    string(body),
		MaxRetries: jobs.DefaultMaxRetries,
	}, now, jobs.SaveOptions{ReopenAfter: time.Unix(1, 0)}, wake)
	return err
}

// SyncAudio 拉取音频时长，写入 audio_parts，并在下载开关打开时排队。
func (s Sync) SyncAudio(ctx context.Context, videoID int64) error {
	if videoID <= 0 {
		return fmt.Errorf("audio id is empty")
	}
	now := s.now()
	if until, active, err := suspendUntil(s.DB, now); err != nil {
		return err
	} else if active {
		return &SuspendedError{Until: until}
	}
	return s.syncAudio(ctx, videoID, now)
}

func (s Sync) syncAudio(ctx context.Context, videoID int64, now time.Time) error {
	var video model.Video
	err := s.DB.WithContext(ctx).First(&video, videoID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return taskrun.ErrCannotRetry
	}
	if err != nil {
		return err
	}
	if video.DeletedAt != nil && !video.DeletedAt.Time.IsZero() {
		return nil
	}
	if !video.IsAudio() || video.Invalid != 0 {
		return nil
	}
	song, err := bilibili.AudioInfo(ctx, s.API, videoID)
	if err != nil {
		if bilibili.IsRiskControl(err) {
			return s.suspend(now)
		}
		return err
	}
	if err := s.saveAudioPart(ctx, videoID, song.Duration, now); err != nil {
		return err
	}
	if s.DataDir == "" {
		return nil
	}
	_, err = download.EnqueueAudio(ctx, s.DB, s.DataDir, videoID, now, s.Wake)
	return err
}

func (s Sync) saveAudioPart(ctx context.Context, videoID int64, duration int, now time.Time) error {
	stamp := dbtime.DateTime{Time: now}
	var part model.AudioPart
	err := s.DB.WithContext(ctx).Where("video_id = ?", videoID).First(&part).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		part = model.AudioPart{VideoID: videoID, Sid: videoID, Duration: duration, CreatedAt: &stamp, UpdatedAt: &stamp}
		return s.DB.WithContext(ctx).Create(&part).Error
	}
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Model(&part).Updates(map[string]any{
		"sid":        videoID,
		"duration":   duration,
		"updated_at": stamp,
	}).Error
}
