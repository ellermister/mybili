package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

const (
	// StatusPending 表示还没有交给 worker。
	StatusPending = "pending"
	// StatusQueued 表示已经推进 Redis，worker 还没开始。
	StatusQueued = "queued"
	// StatusRunning 表示正在执行。
	StatusRunning = "running"
	// StatusDone 表示已经成功结束。
	StatusDone = "done"
	// StatusFailed 表示不再自动重试。
	StatusFailed = "failed"
	// StatusCancelled 表示人工取消，取到后直接跳过。
	StatusCancelled = "cancelled"

	// DefaultMaxRetries 是新建任务的默认重试上限。
	DefaultMaxRetries = 3
)

// Draft 是准备写入 tasks 的一行。UniqueKey 相同的活跃行不会重复插入。
type Draft struct {
	Queue      string
	Type       string
	UniqueKey  string
	Payload    string
	MaxRetries int
}

// SaveOptions 控制已经结束的行要不要在这一档重新打开。
type SaveOptions struct {
	// ReopenAfter 之后才允许把 done 或 failed 重新打开。
	// 同一分钟里任务已经跑完时，不再打开第二次。零值表示只要结束了就打开。
	ReopenAfter time.Time
}

// SaveResult 是一次写入的结果。Opened 表示新建或重新打开。
// Wake 表示库里有 pending，调用方应通知 Feeder。
type SaveResult struct {
	ID     int64
	Opened bool
	Wake   bool
}

// Store 读写 tasks 表。worker 不使用它。
type Store struct {
	DB *gorm.DB
}

// Save 按 unique_key 保证最多一条活跃任务。
// pending、queued、running 都算活跃。结束的行在 ReopenAfter 之前不打开。
func (s Store) Save(ctx context.Context, draft Draft, now time.Time, opt SaveOptions) (SaveResult, error) {
	if s.DB == nil {
		return SaveResult{}, fmt.Errorf("job store is incomplete")
	}
	if draft.UniqueKey == "" || draft.Queue == "" || draft.Type == "" {
		return SaveResult{}, fmt.Errorf("job draft is incomplete")
	}
	if draft.MaxRetries <= 0 {
		draft.MaxRetries = DefaultMaxRetries
	}
	stamp := dbtime.DateTime{Time: now}
	var result SaveResult
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		active, found, err := latestJob(tx, draft.UniqueKey, []string{StatusPending, StatusQueued, StatusRunning})
		if err != nil {
			return err
		}
		if found {
			result.ID = active.ID
			result.Wake = active.Status == StatusPending
			return nil
		}
		row, found, err := latestJob(tx, draft.UniqueKey, nil)
		if err != nil {
			return err
		}
		if !found {
			created := model.Job{
				Queue:      draft.Queue,
				Type:       draft.Type,
				UniqueKey:  draft.UniqueKey,
				Status:     StatusPending,
				Payload:    draft.Payload,
				MaxRetries: draft.MaxRetries,
				CreatedAt:  &stamp,
				UpdatedAt:  &stamp,
			}
			if err := tx.Create(&created).Error; err != nil {
				return err
			}
			result = SaveResult{ID: created.ID, Opened: true, Wake: true}
			return nil
		}
		if row.UpdatedAt != nil && !opt.ReopenAfter.IsZero() && !row.UpdatedAt.Time.Before(opt.ReopenAfter) {
			result.ID = row.ID
			return nil
		}
		err = tx.Model(&row).Updates(map[string]any{
			"queue":       draft.Queue,
			"type":        draft.Type,
			"status":      StatusPending,
			"payload":     draft.Payload,
			"error_msg":   "",
			"retry_count": 0,
			"started_at":  gorm.Expr("NULL"),
			"finished_at": gorm.Expr("NULL"),
			"updated_at":  stamp,
		}).Error
		if err != nil {
			return err
		}
		result = SaveResult{ID: row.ID, Opened: true, Wake: true}
		return nil
	})
	return result, err
}

// Enqueue 写入 tasks，并在还有 pending 时通知 Feeder。wake 为空则只落表。
func (s Store) Enqueue(ctx context.Context, draft Draft, now time.Time, opt SaveOptions, wake func(context.Context, string) error) (SaveResult, error) {
	result, err := s.Save(ctx, draft, now, opt)
	if err != nil || !result.Wake || wake == nil {
		return result, err
	}
	if err := wake(ctx, draft.Queue); err != nil {
		return result, err
	}
	return result, nil
}

// Load 读取一行任务。
func (s Store) Load(ctx context.Context, id int64) (model.Job, error) {
	if s.DB == nil {
		return model.Job{}, fmt.Errorf("job store is incomplete")
	}
	var job model.Job
	err := s.DB.WithContext(ctx).First(&job, id).Error
	return job, err
}

// MarkRunning 把 queued 或 pending 改成 running。改不到时返回 false。
func (s Store) MarkRunning(ctx context.Context, id int64, now time.Time) (bool, error) {
	return s.update(ctx, id, []string{StatusQueued, StatusPending}, map[string]any{
		"status":     StatusRunning,
		"started_at": dbtime.DateTime{Time: now},
		"updated_at": dbtime.DateTime{Time: now},
	})
}

// MarkDone 在仍是 running 时记成功。
func (s Store) MarkDone(ctx context.Context, id int64, now time.Time) (bool, error) {
	return s.update(ctx, id, []string{StatusRunning}, map[string]any{
		"status":      StatusDone,
		"finished_at": dbtime.DateTime{Time: now},
		"updated_at":  dbtime.DateTime{Time: now},
		"error_msg":   "",
	})
}

// MarkFailed 在仍是 running 时记失败，不再自动重试。
func (s Store) MarkFailed(ctx context.Context, id int64, msg string, retry int, now time.Time) (bool, error) {
	return s.update(ctx, id, []string{StatusRunning}, map[string]any{
		"status":      StatusFailed,
		"error_msg":   msg,
		"retry_count": retry,
		"finished_at": dbtime.DateTime{Time: now},
		"updated_at":  dbtime.DateTime{Time: now},
	})
}

// MarkPending 把 running 或 queued 改回 pending，交给 Feeder 再投。
func (s Store) MarkPending(ctx context.Context, id int64, retry int, msg string, now time.Time) (bool, error) {
	return s.update(ctx, id, []string{StatusRunning, StatusQueued}, map[string]any{
		"status":      StatusPending,
		"retry_count": retry,
		"error_msg":   msg,
		"started_at":  gorm.Expr("NULL"),
		"finished_at": gorm.Expr("NULL"),
		"updated_at":  dbtime.DateTime{Time: now},
	})
}

// TakePending 取这个队列里最早的 pending，并改成 queued。
func (s Store) TakePending(ctx context.Context, queue string, now time.Time) (model.Job, bool, error) {
	if s.DB == nil {
		return model.Job{}, false, fmt.Errorf("job store is incomplete")
	}
	var rows []model.Job
	err := s.DB.WithContext(ctx).Where("queue = ? AND status = ?", queue, StatusPending).Order("id").Limit(1).Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return model.Job{}, false, err
	}
	job := rows[0]
	ok, err := s.update(ctx, job.ID, []string{StatusPending}, map[string]any{
		"status":     StatusQueued,
		"updated_at": dbtime.DateTime{Time: now},
	})
	if err != nil || !ok {
		return model.Job{}, false, err
	}
	job.Status = StatusQueued
	return job, true, nil
}

// CountPending 返回这个队列里还没投喂的行数。启动时用来补一次投喂。
func (s Store) CountPending(ctx context.Context, queue string) (int64, error) {
	if s.DB == nil {
		return 0, fmt.Errorf("job store is incomplete")
	}
	var n int64
	err := s.DB.WithContext(ctx).Model(&model.Job{}).Where("queue = ? AND status = ?", queue, StatusPending).Count(&n).Error
	return n, err
}

// ResetStaleQueued 把长时间停在 queued 的行改回 pending，并返回涉及的队列。
func (s Store) ResetStaleQueued(ctx context.Context, queues []string, before time.Time) ([]string, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("job store is incomplete")
	}
	if len(queues) == 0 {
		return nil, nil
	}
	var rows []model.Job
	err := s.DB.WithContext(ctx).Where("queue IN ? AND status = ? AND updated_at < ?", queues, StatusQueued, before).Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	seen := map[string]struct{}{}
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
		if _, ok := seen[row.Queue]; ok {
			continue
		}
		seen[row.Queue] = struct{}{}
		names = append(names, row.Queue)
	}
	err = s.DB.WithContext(ctx).Model(&model.Job{}).Where("id IN ? AND status = ?", ids, StatusQueued).Updates(map[string]any{
		"status":     StatusPending,
		"updated_at": dbtime.DateTime{Time: dbtime.Now()},
	}).Error
	return names, err
}

// Running 返回这些队列里正在执行的行，供收回使用。
func (s Store) Running(ctx context.Context, queues []string) ([]model.Job, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("job store is incomplete")
	}
	if len(queues) == 0 {
		return nil, nil
	}
	var rows []model.Job
	err := s.DB.WithContext(ctx).Where("queue IN ? AND status = ?", queues, StatusRunning).Find(&rows).Error
	return rows, err
}

func (s Store) update(ctx context.Context, id int64, from []string, updates map[string]any) (bool, error) {
	if s.DB == nil {
		return false, fmt.Errorf("job store is incomplete")
	}
	res := s.DB.WithContext(ctx).Model(&model.Job{}).Where("id = ? AND status IN ?", id, from).Updates(updates)
	return res.RowsAffected > 0, res.Error
}

// List 按状态分页返回任务，新的在前。statuses 为空时不过滤。
func (s Store) List(ctx context.Context, statuses []string, page, perPage int) ([]model.Job, int64, error) {
	if s.DB == nil {
		return nil, 0, fmt.Errorf("job store is incomplete")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 30
	}
	query := s.DB.WithContext(ctx).Model(&model.Job{})
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Job
	err := query.Order("id desc").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error
	return rows, total, err
}

// CountByStatus 统计每种状态的行数。没有的状态不会出现在结果里。
func (s Store) CountByStatus(ctx context.Context) (map[string]int64, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("job store is incomplete")
	}
	var rows []struct {
		Status string
		Count  int64
	}
	err := s.DB.WithContext(ctx).Model(&model.Job{}).Select("status, COUNT(*) AS count").Group("status").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := map[string]int64{}
	for _, row := range rows {
		counts[row.Status] = row.Count
	}
	return counts, nil
}

// Cancel 把还没成功结束的任务改成取消。已经结束的返回 false。
func (s Store) Cancel(ctx context.Context, id int64, now time.Time) (bool, error) {
	return s.update(ctx, id, []string{StatusPending, StatusQueued, StatusRunning}, map[string]any{
		"status":      StatusCancelled,
		"error_msg":   "已取消",
		"finished_at": dbtime.DateTime{Time: now},
		"updated_at":  dbtime.DateTime{Time: now},
	})
}

// Retry 把失败或取消的任务改回 pending。同键已经有活跃行时不打开。
func (s Store) Retry(ctx context.Context, id int64, now time.Time) (model.Job, bool, error) {
	if s.DB == nil {
		return model.Job{}, false, fmt.Errorf("job store is incomplete")
	}
	job, err := s.Load(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Job{}, false, nil
	}
	if err != nil {
		return model.Job{}, false, err
	}
	if job.Status != StatusFailed && job.Status != StatusCancelled {
		return job, false, nil
	}
	active, found, err := latestJob(s.DB.WithContext(ctx), job.UniqueKey, []string{StatusPending, StatusQueued, StatusRunning})
	if err != nil {
		return job, false, err
	}
	if found && active.ID != job.ID {
		return job, false, nil
	}
	ok, err := s.update(ctx, id, []string{StatusFailed, StatusCancelled}, map[string]any{
		"status":      StatusPending,
		"error_msg":   "",
		"retry_count": 0,
		"started_at":  gorm.Expr("NULL"),
		"finished_at": gorm.Expr("NULL"),
		"updated_at":  dbtime.DateTime{Time: now},
	})
	if err != nil || !ok {
		return job, false, err
	}
	job.Status = StatusPending
	job.RetryCount = 0
	return job, true, nil
}

func latestJob(tx *gorm.DB, key string, statuses []string) (model.Job, bool, error) {
	var rows []model.Job
	query := tx.Where("unique_key = ?", key)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Order("id desc").Limit(1).Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return model.Job{}, false, err
	}
	return rows[0], true, nil
}
