package syncbili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/notify"
	"mybili/internal/schedule"
	"mybili/internal/taskrun"
)

const (
	// TypeFavList 是收藏夹列表同步任务。
	TypeFavList = "fav-list"
	// TypeFavVideos 是收藏夹视频同步任务。payload 里的 page 为 0 表示全部页。
	TypeFavVideos = "fav-videos"
	// TypeCover 是封面和头像下载。消息不落 tasks。
	TypeCover = "cover"
	// TypeFixInvalid 是一页收藏夹里的无效视频修复。
	TypeFixInvalid = "fix-invalid-fav"

	settingFavoriteSync = "favorite_sync_enabled"
)

type favVideoPayload struct {
	Page int `json:"page"`
}

type fixInvalidPayload struct {
	FavID int64 `json:"fav_id"`
	Page  int   `json:"page"`
}

// FavSpecs 返回收藏夹定时同步的落表规则。server 写完 pending 后通知 Feeder。
func FavSpecs(db *gorm.DB) []schedule.Spec {
	enabled := func(context.Context) (bool, error) {
		return settingOn(db, settingFavoriteSync)
	}
	return []schedule.Spec{
		{
			Name:    "收藏夹列表",
			Due:     everyTenMinutes,
			Enabled: enabled,
			Tasks: func(_ context.Context, now time.Time) ([]schedule.Draft, error) {
				return []schedule.Draft{favListDraft(now)}, nil
			},
		},
		{
			Name:    "收藏夹视频",
			Due:     everyTenMinutes,
			Enabled: enabled,
			Tasks: func(_ context.Context, now time.Time) ([]schedule.Draft, error) {
				page, err := favVideoPage(db, now)
				if err != nil || page < 0 {
					return nil, err
				}
				draft, err := favVideoDraft(now, page, false)
				if err != nil {
					return nil, err
				}
				return []schedule.Draft{draft}, nil
			},
		},
		{
			Name: "收藏夹视频全量",
			Due: func(now time.Time) bool {
				return now.Hour() == 4 && now.Minute() == 0
			},
			Enabled: enabled,
			Tasks: func(_ context.Context, now time.Time) ([]schedule.Draft, error) {
				draft, err := favVideoDraft(now, 0, true)
				if err != nil {
					return nil, err
				}
				return []schedule.Draft{draft}, nil
			},
		},
	}
}

// FixInvalidSpecs 每天 04:00 按收藏夹分页写入修复任务。不看收藏夹同步开关，也不在这里请求 Bilibili。
func FixInvalidSpecs(db *gorm.DB, pageSize int) []schedule.Spec {
	if pageSize < 1 {
		pageSize = 40
	}
	return []schedule.Spec{{
		Name: "修复无效收藏",
		Due: func(now time.Time) bool {
			return now.Hour() == 4 && now.Minute() == 0
		},
		Tasks: func(ctx context.Context, now time.Time) ([]schedule.Draft, error) {
			var folders []model.FavoriteList
			if err := db.WithContext(ctx).Order("id").Find(&folders).Error; err != nil {
				return nil, err
			}
			reopen := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
			drafts := make([]schedule.Draft, 0)
			for _, folder := range folders {
				pages := 0
				if folder.MediaCount > 0 {
					pages = (folder.MediaCount + pageSize - 1) / pageSize
				}
				for page := 1; page <= pages; page++ {
					body, err := json.Marshal(fixInvalidPayload{FavID: folder.ID, Page: page})
					if err != nil {
						return nil, err
					}
					drafts = append(drafts, schedule.Draft{
						Queue:       taskrun.QueueBiliFast,
						Type:        TypeFixInvalid,
						UniqueKey:   fmt.Sprintf("sync:fix-invalid:%d:%d", folder.ID, page),
						Payload:     string(body),
						MaxRetries:  jobs.DefaultMaxRetries,
						ReopenAfter: reopen,
					})
				}
			}
			return drafts, nil
		},
	}}
}

// Register 把同步和封面下载登记到 worker。
func Register(tasks *taskrun.Tasks, db *gorm.DB, pageSize int, dataDir string, hooks Hooks) {
	handlers := taskHandlers(tasks, func() (Sync, error) {
		syncer, err := New(db, pageSize)
		if err != nil {
			return Sync{}, err
		}
		syncer.DataDir = dataDir
		syncer.CoverPush = CoverRedis(tasks)
		syncer.Wake = func(ctx context.Context, queue string) error {
			if tasks == nil {
				return nil
			}
			return tasks.Wake(ctx, queue)
		}
		syncer.UseNotify(tasks)
		syncer.AfterLinked = hooks.AfterLinked
		syncer.AfterUnlinked = hooks.AfterUnlinked
		return syncer, nil
	})
	for taskType, handler := range handlers {
		tasks.Handle(taskType, handler)
	}
}

func taskHandlers(tasks *taskrun.Tasks, open func() (Sync, error)) map[string]taskrun.Handler {
	store := func(syncer Sync) jobs.Store { return jobs.Store{DB: syncer.DB} }
	return map[string]taskrun.Handler{
		TypeFavList: func(ctx context.Context, msg taskrun.Message) error {
			syncer, err := open()
			if err != nil {
				return err
			}
			return runSaved(ctx, store(syncer), tasks, msg, func(ctx context.Context) error {
				_, err := syncer.Folders(ctx)
				return err
			})
		},
		TypeFixFreeze: func(ctx context.Context, msg taskrun.Message) error {
			syncer, err := open()
			if err != nil {
				return err
			}
			return runSaved(ctx, store(syncer), tasks, msg, func(ctx context.Context) error {
				_, err := syncer.FixFreeze(ctx)
				return err
			})
		},
		TypeFixInvalid: func(ctx context.Context, msg taskrun.Message) error {
			syncer, err := open()
			if err != nil {
				return err
			}
			return runSaved(ctx, store(syncer), tasks, msg, func(ctx context.Context) error {
				var payload fixInvalidPayload
				if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil || payload.FavID <= 0 || payload.Page <= 0 {
					return taskrun.ErrCannotRetry
				}
				_, err := syncer.FixInvalidFavorite(ctx, payload.FavID, payload.Page)
				return err
			})
		},
		TypeFavVideos: func(ctx context.Context, msg taskrun.Message) error {
			syncer, err := open()
			if err != nil {
				return err
			}
			return runSaved(ctx, store(syncer), tasks, msg, func(ctx context.Context) error {
				var payload favVideoPayload
				if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
					return taskrun.ErrCannotRetry
				}
				if payload.Page < 0 {
					return taskrun.ErrCannotRetry
				}
				_, err := syncer.FavoriteVideos(ctx, 0, payload.Page)
				return err
			})
		},
		TypeSubscription: func(ctx context.Context, msg taskrun.Message) error {
			syncer, err := open()
			if err != nil {
				return err
			}
			syncer.QueueVideo = func(ctx context.Context, bvid string) error {
				return EnqueueVideoDetail(ctx, syncer.DB, bvid, time.Now(), func(ctx context.Context, queue string) error {
					if tasks == nil {
						return nil
					}
					return tasks.Wake(ctx, queue)
				})
			}
			return runSaved(ctx, store(syncer), tasks, msg, func(ctx context.Context) error {
				body, err := subscriptionTask(msg.Payload)
				if err != nil {
					return taskrun.ErrCannotRetry
				}
				return syncer.UpdateSubscription(ctx, body.ID, body.PullAll)
			})
		},
		TypeSyncAudio: func(ctx context.Context, msg taskrun.Message) error {
			syncer, err := open()
			if err != nil {
				return err
			}
			return runSaved(ctx, store(syncer), tasks, msg, func(ctx context.Context) error {
				var body struct {
					ID int64 `json:"id"`
				}
				if err := json.Unmarshal([]byte(msg.Payload), &body); err != nil || body.ID <= 0 {
					return taskrun.ErrCannotRetry
				}
				return syncer.SyncAudio(ctx, body.ID)
			})
		},
		TypeSyncVideo: func(ctx context.Context, msg taskrun.Message) error {
			syncer, err := open()
			if err != nil {
				return err
			}
			return runSaved(ctx, store(syncer), tasks, msg, func(ctx context.Context) error {
				var body struct {
					Bvid string `json:"bvid"`
				}
				if err := json.Unmarshal([]byte(msg.Payload), &body); err != nil || strings.TrimSpace(body.Bvid) == "" {
					return taskrun.ErrCannotRetry
				}
				return syncer.Video(ctx, 0, body.Bvid)
			})
		},
		TypeCover: func(ctx context.Context, msg taskrun.Message) error {
			var payload coverPayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				return taskrun.ErrCannotRetry
			}
			syncer, err := open()
			if err != nil {
				return err
			}
			err = syncer.Cover(ctx, payload.Kind, payload.ID, payload.URL)
			if err == nil {
				if syncer.CoverPush != nil {
					_ = syncer.CoverPush.Forget(ctx, coverTaskKey(payload.Kind, payload.ID))
				}
				return nil
			}
			if errors.Is(err, ErrThrottled) && tasks != nil {
				_ = tasks.Pause(ctx, taskrun.QueueCover, 2*time.Hour)
			}
			return err
		},
	}
}

// runSaved 执行一条已经落表的任务，并把状态写回 tasks。可重试的错误留给 Feeder 再投。
func runSaved(ctx context.Context, store jobs.Store, tasks *taskrun.Tasks, msg taskrun.Message, work func(context.Context) error) error {
	job, err := store.Load(ctx, msg.ID)
	if err != nil {
		return err
	}
	switch job.Status {
	case jobs.StatusCancelled, jobs.StatusDone, jobs.StatusFailed:
		return nil
	}
	now := time.Now()
	ok, err := store.MarkRunning(ctx, msg.ID, now)
	if err != nil || !ok {
		return err
	}
	err = work(ctx)
	if err == nil {
		_, err = store.MarkDone(ctx, msg.ID, time.Now())
		return err
	}
	var suspended *SuspendedError
	if errors.As(err, &suspended) {
		_, _ = store.MarkFailed(ctx, msg.ID, err.Error(), job.RetryCount+1, time.Now())
		if tasks != nil {
			_ = tasks.Pause(ctx, job.Queue, time.Until(suspended.Until))
		}
		return nil
	}
	if errors.Is(err, taskrun.ErrCannotRetry) || job.RetryCount+1 >= maxRetries(job) {
		_, failErr := store.MarkFailed(ctx, msg.ID, err.Error(), job.RetryCount+1, time.Now())
		if failErr != nil {
			return failErr
		}
		return taskrun.ErrCannotRetry
	}
	if _, pendingErr := store.MarkPending(ctx, msg.ID, job.RetryCount+1, err.Error(), time.Now()); pendingErr != nil {
		return pendingErr
	}
	if tasks != nil {
		if wakeErr := tasks.Wake(ctx, job.Queue); wakeErr != nil {
			log.Printf("通知投喂失败 %s: %v", job.Queue, wakeErr)
		}
	}
	return err
}

func maxRetries(job model.Job) int {
	if job.MaxRetries > 0 {
		return job.MaxRetries
	}
	return jobs.DefaultMaxRetries
}

func everyTenMinutes(now time.Time) bool {
	return now.Minute()%10 == 0
}

func favVideoPage(db *gorm.DB, now time.Time) (int, error) {
	var count int64
	err := db.Model(&model.Video{}).Where("deleted_at IS NULL").Count(&count).Error
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}
	if now.Hour() == 4 {
		return -1, nil
	}
	return 1, nil
}

func favListDraft(now time.Time) schedule.Draft {
	return schedule.Draft{
		Queue:       taskrun.QueueBiliFast,
		Type:        TypeFavList,
		UniqueKey:   "sync:fav-list",
		Payload:     "{}",
		MaxRetries:  jobs.DefaultMaxRetries,
		ReopenAfter: now.Truncate(time.Minute),
	}
}

func favVideoDraft(now time.Time, page int, daily bool) (schedule.Draft, error) {
	encoded, err := json.Marshal(favVideoPayload{Page: page})
	if err != nil {
		return schedule.Draft{}, err
	}
	reopen := now.Truncate(time.Minute)
	if daily {
		reopen = time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	}
	return schedule.Draft{
		Queue:       taskrun.QueueBiliFast,
		Type:        TypeFavVideos,
		UniqueKey:   fmt.Sprintf("sync:fav-videos:%d", page),
		Payload:     string(encoded),
		MaxRetries:  jobs.DefaultMaxRetries,
		ReopenAfter: reopen,
	}, nil
}

// UseNotify 把 Cookie 过期和视频冻结接到通知队列。tasks 为空时不发送。
func (s *Sync) UseNotify(tasks *taskrun.Tasks) {
	if s == nil || tasks == nil {
		return
	}
	s.NotifyFrozen = func(ctx context.Context, videoID int64) {
		_ = notify.EnqueueFrozen(ctx, tasks, videoID)
	}
	if s.API == nil {
		return
	}
	s.API.OnError = func(err error) {
		if bilibili.IsCookieExpired(err) {
			_ = notify.EnqueueCookie(context.Background(), tasks)
		}
	}
}

func settingOn(db *gorm.DB, name string) (bool, error) {
	text, err := readString(db, name)
	if err != nil || text == "" {
		return false, err
	}
	return text == "on", nil
}
