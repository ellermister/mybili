package schedule

import (
	"context"
	"errors"
	"log"
	"time"

	"mybili/internal/dbtime"
	"mybili/internal/jobs"
)

const tickEvery = 20 * time.Second

// Draft 是定时器准备写入 tasks 的一行。写入后再通知 Feeder。
type Draft struct {
	Queue       string
	Type        string
	UniqueKey   string
	Payload     string
	MaxRetries  int
	ReopenAfter time.Time
}

// Spec 描述一类定时任务何时产生、开关是否打开、要写入哪些行。
type Spec struct {
	Name    string
	Due     func(time.Time) bool
	Enabled func(context.Context) (bool, error)
	Tasks   func(context.Context, time.Time) ([]Draft, error)
}

// Runner 按 Spec 把到点的任务写成 pending，并通知 Feeder。
type Runner struct {
	Specs []Spec
	Store jobs.Store
	Now   func() time.Time
	// Wake 通知 Feeder 某个队列有 pending。为空时只落表。
	Wake func(context.Context, string) error
}

// Loop 大约每 20 秒看一次钟。
func (r Runner) Loop(ctx context.Context) {
	r.runOnce(ctx)
	ticker := time.NewTicker(tickEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runOnce(ctx)
		}
	}
}

func (r Runner) runOnce(ctx context.Context) {
	if err := r.Tick(ctx, r.now()); err != nil {
		log.Printf("定时任务派发失败: %v", err)
	}
}

// Tick 把到点的任务写入 tasks，不请求 Bilibili。有 pending 时通知 Feeder。
func (r Runner) Tick(ctx context.Context, now time.Time) error {
	if r.Store.DB == nil {
		return errors.New("schedule runner is incomplete")
	}
	now = now.In(dbtime.Location())
	for _, spec := range r.Specs {
		if err := r.dispatch(ctx, spec, now); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) dispatch(ctx context.Context, spec Spec, now time.Time) error {
	if spec.Due != nil && !spec.Due(now) {
		return nil
	}
	if spec.Enabled != nil {
		on, err := spec.Enabled(ctx)
		if err != nil || !on {
			return err
		}
	}
	if spec.Tasks == nil {
		return nil
	}
	drafts, err := spec.Tasks(ctx, now)
	if err != nil {
		return err
	}
	for _, draft := range drafts {
		result, err := r.Store.Enqueue(ctx, jobs.Draft{
			Queue:      draft.Queue,
			Type:       draft.Type,
			UniqueKey:  draft.UniqueKey,
			Payload:    draft.Payload,
			MaxRetries: draft.MaxRetries,
		}, now, jobs.SaveOptions{ReopenAfter: draft.ReopenAfter}, r.Wake)
		if err != nil {
			return err
		}
		if result.Opened {
			log.Printf("定时任务已写入 %s id=%d", draft.UniqueKey, result.ID)
		}
	}
	return nil
}

func (r Runner) now() time.Time {
	if r.Now != nil {
		return r.Now().In(dbtime.Location())
	}
	return dbtime.Now()
}
