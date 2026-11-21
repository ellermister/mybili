package syncbili

import (
	"context"
	"fmt"
	"time"

	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/schedule"
	"mybili/internal/taskrun"
)

// TypeFixFreeze 是每天修正冻结且无效视频的任务。不请求 Bilibili。
const TypeFixFreeze = "fix-freeze"

// FixFreezeSpecs 每天 00:00 写入一条修正任务。不看同步开关。
func FixFreezeSpecs() []schedule.Spec {
	return []schedule.Spec{{
		Name: "修正冻结状态",
		Due: func(now time.Time) bool {
			return now.Hour() == 0 && now.Minute() == 0
		},
		Tasks: func(_ context.Context, now time.Time) ([]schedule.Draft, error) {
			return []schedule.Draft{fixFreezeDraft(now)}, nil
		},
	}}
}

// fixFreezeDraft 用调用方传入的时间生成当天 00:00 才允许重开的任务。
func fixFreezeDraft(now time.Time) schedule.Draft {
	reopen := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return schedule.Draft{
		Queue:       taskrun.QueueBiliFast,
		Type:        TypeFixFreeze,
		UniqueKey:   "sync:fix-freeze",
		Payload:     "{}",
		MaxRetries:  jobs.DefaultMaxRetries,
		ReopenAfter: reopen,
	}
}

// FixFreeze 修正冻结且无效、但下载次数记成 0 的视频。
// 按分 P 和音频上的下载时间重算次数。重算后仍是 0 就取消冻结，有下载则保持冻结。
// 次数已经不是 0 的不改。不改无效标记，也不下载封面。
func (s Sync) FixFreeze(ctx context.Context) (int, error) {
	if s.DB == nil {
		return 0, fmt.Errorf("syncer is incomplete")
	}
	var videos []model.Video
	err := s.DB.WithContext(ctx).
		Where("frozen = 1 AND invalid = 1 AND deleted_at IS NULL").
		Where("video_downloaded_num = 0 AND audio_downloaded_num = 0").
		Order("id").
		Find(&videos).Error
	if err != nil {
		return 0, err
	}
	stamp := dbtime.DateTime{Time: s.now()}
	fixed := 0
	for _, video := range videos {
		videoNum, audioNum, err := s.downloadedCounts(ctx, video.ID)
		if err != nil {
			return fixed, err
		}
		frozen := 0
		if videoNum > 0 || audioNum > 0 {
			frozen = 1
		}
		err = s.DB.WithContext(ctx).Model(&video).Updates(map[string]any{
			"video_downloaded_num": videoNum,
			"audio_downloaded_num": audioNum,
			"frozen":               frozen,
			"updated_at":           stamp,
		}).Error
		if err != nil {
			return fixed, err
		}
		fixed++
	}
	return fixed, nil
}

// downloadedCounts 统计这个视频已经记下下载时间的分 P 和音频条数。
func (s Sync) downloadedCounts(ctx context.Context, videoID int64) (int64, int64, error) {
	var videoNum int64
	err := s.DB.WithContext(ctx).Model(&model.VideoPart{}).
		Where("video_id = ? AND video_downloaded_at IS NOT NULL", videoID).
		Count(&videoNum).Error
	if err != nil {
		return 0, 0, err
	}
	var audioNum int64
	err = s.DB.WithContext(ctx).Model(&model.AudioPart{}).
		Where("video_id = ? AND audio_downloaded_at IS NOT NULL", videoID).
		Count(&audioNum).Error
	return videoNum, audioNum, err
}
