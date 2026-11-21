package download

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/notify"
	"mybili/internal/taskrun"
)

// Register 把视频和弹幕下载登记到 worker。
// hooks.PartReady 在文件记库成功后调用，它的失败不会改下载结果。
func Register(tasks *taskrun.Tasks, db *gorm.DB, dataDir string, hooks Hooks) {
	runner := Runner{DB: db, DataDir: dataDir, Logs: tasks, OnCookie: func(ctx context.Context) {
		_ = notify.EnqueueCookie(ctx, tasks)
	}}
	store := jobs.Store{DB: db}
	tasks.Handle(TypeVideo, func(ctx context.Context, msg taskrun.Message) error {
		return runSaved(ctx, store, tasks, msg, func(ctx context.Context) error {
			var payload partPayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil || payload.PartID <= 0 {
				return taskrun.ErrCannotRetry
			}
			ready, err := runner.Video(ctx, msg.ID, payload)
			if err != nil {
				return err
			}
			if hooks.PartReady != nil {
				hooks.PartReady(ctx, ready)
			}
			return nil
		})
	})
	tasks.Handle(TypeAudio, func(ctx context.Context, msg taskrun.Message) error {
		return runSaved(ctx, store, tasks, msg, func(ctx context.Context) error {
			var payload audioPayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil || payload.PartID <= 0 {
				return taskrun.ErrCannotRetry
			}
			return runner.Audio(ctx, msg.ID, payload)
		})
	})
	tasks.Handle(TypeDanmaku, func(ctx context.Context, msg taskrun.Message) error {
		return runSaved(ctx, store, tasks, msg, func(ctx context.Context) error {
			var payload partPayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil || payload.PartID <= 0 {
				return taskrun.ErrCannotRetry
			}
			return runner.Danmaku(ctx, msg.ID, payload)
		})
	})
}

func runSaved(ctx context.Context, store jobs.Store, tasks *taskrun.Tasks, msg taskrun.Message, work func(context.Context) error) error {
	job, err := store.Load(ctx, msg.ID)
	if err != nil {
		return err
	}
	switch job.Status {
	case jobs.StatusCancelled, jobs.StatusDone, jobs.StatusFailed:
		return nil
	}
	ok, err := store.MarkRunning(ctx, msg.ID, time.Now())
	if err != nil || !ok {
		return err
	}
	err = work(ctx)
	if errors.Is(err, errSkipDownload) {
		if store.DB != nil {
			if delErr := store.DB.WithContext(ctx).Delete(&model.Job{}, msg.ID).Error; delErr != nil {
				return delErr
			}
		}
		return nil
	}
	if err == nil {
		_, err = store.MarkDone(ctx, msg.ID, time.Now())
		return err
	}
	if bilibili.IsRiskControl(err) {
		_, _ = store.MarkFailed(ctx, msg.ID, err.Error(), job.RetryCount+1, time.Now())
		if tasks != nil {
			_ = tasks.Pause(ctx, job.Queue, 2*time.Hour)
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
