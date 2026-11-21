package feeder

import (
	"context"
	"log"
	"math"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/taskrun"
)

const (
	wakeWait  = 5 * time.Second
	pauseWait = 5 * time.Second
	retryWait = time.Second
	reapEvery = 30 * time.Second
)

// Broker 是 Feeder 用到的 Redis 操作。worker 的消费循环不在这里。
type Broker interface {
	Paused(ctx context.Context, queue string) (bool, error)
	Len(ctx context.Context, queue string) (int64, error)
	Take(ctx context.Context, queue string, every time.Duration, limit int) (bool, error)
	Push(ctx context.Context, queue string, msg taskrun.Message) error
	Held(ctx context.Context, id int64) (bool, error)
	Wake(ctx context.Context, queue string) error
	NextWake(ctx context.Context, wait time.Duration) (string, bool, error)
}

// Feeder 从 tasks 表把 pending 推进限速队列。它阻塞在收件列表上，没有任务时不查库。
type Feeder struct {
	DB     *gorm.DB
	Tasks  Broker
	Queues []taskrun.Queue
	Now    func() time.Time
}

// Loop 阻塞等待收件列表。业务写完 pending 后调用 Wake。大约每 30 秒收回中断的行。
func (f Feeder) Loop(ctx context.Context) {
	if f.DB == nil || f.Tasks == nil {
		log.Printf("投喂未启动: 配置不完整")
		return
	}
	delayed := map[string]time.Time{}
	f.boot(ctx, delayed)
	nextReap := time.Now().Add(reapEvery)
	for {
		if ctx.Err() != nil {
			return
		}
		if !time.Now().Before(nextReap) {
			if err := f.Reap(ctx); err != nil {
				log.Printf("收回任务失败: %v", err)
			}
			nextReap = time.Now().Add(reapEvery)
		}
		if name, ok := dueQueue(delayed, f.now()); ok {
			f.drainNamed(ctx, name, delayed)
			continue
		}
		wait := wakeWait
		if d := time.Until(nextReap); d > 0 && d < wait {
			wait = d
		}
		if d, ok := soonerDelay(delayed, f.now()); ok && d < wait {
			wait = d
		}
		// 将 wait 向上取整
		wait = time.Duration(math.Ceil(wait.Seconds())) * time.Second
		name, ok, err := f.Tasks.NextWake(ctx, wait)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("等待投喂失败: %v", err)
			if sleep(ctx, time.Second) != nil {
				return
			}
			continue
		}
		if !ok {
			continue
		}
		f.drainNamed(ctx, name, delayed)
	}
}

// Tick 把当前 pending 各投喂一轮。测试和启动补投使用，平时由收件列表唤醒。
func (f Feeder) Tick(ctx context.Context) error {
	if f.DB == nil || f.Tasks == nil {
		return errIncomplete
	}
	for _, queue := range f.Queues {
		if !queue.Feed {
			continue
		}
		if err := f.drain(ctx, queue, nil); err != nil {
			return err
		}
	}
	return nil
}

// Reap 把租约消失的 running，以及过久的 queued，改回 pending，并通知 Feeder。
func (f Feeder) Reap(ctx context.Context) error {
	if f.DB == nil || f.Tasks == nil {
		return errIncomplete
	}
	store := jobs.Store{DB: f.DB}
	names := feedNames(f.Queues)
	now := f.now()
	running, err := store.Running(ctx, names)
	if err != nil {
		return err
	}
	woken := map[string]struct{}{}
	for _, row := range running {
		held, err := f.Tasks.Held(ctx, row.ID)
		if err != nil {
			return err
		}
		if held {
			continue
		}
		if _, err := store.MarkPending(ctx, row.ID, row.RetryCount, "worker interrupted", now); err != nil {
			return err
		}
		woken[row.Queue] = struct{}{}
	}
	stale, err := store.ResetStaleQueued(ctx, names, now.Add(-2*time.Minute))
	if err != nil {
		return err
	}
	for _, name := range stale {
		woken[name] = struct{}{}
	}
	for name := range woken {
		if err := f.Tasks.Wake(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (f Feeder) boot(ctx context.Context, delayed map[string]time.Time) {
	if err := f.Reap(ctx); err != nil {
		log.Printf("启动时收回任务失败: %v", err)
	}
	store := jobs.Store{DB: f.DB}
	for _, queue := range f.Queues {
		if !queue.Feed {
			continue
		}
		n, err := store.CountPending(ctx, queue.Name)
		if err != nil {
			log.Printf("启动时统计待投喂失败 %s: %v", queue.Name, err)
			continue
		}
		if n == 0 {
			continue
		}
		f.drainNamed(ctx, queue.Name, delayed)
	}
}

func (f Feeder) drainNamed(ctx context.Context, name string, delayed map[string]time.Time) {
	queue, ok := f.queueByName(name)
	if !ok {
		return
	}
	delete(delayed, name)
	if err := f.drain(ctx, queue, delayed); err != nil {
		log.Printf("投喂失败 %s: %v", name, err)
	}
}

func (f Feeder) drain(ctx context.Context, queue taskrun.Queue, delayed map[string]time.Time) error {
	for {
		again, wait, err := f.feedOne(ctx, queue)
		if err != nil {
			f.deferQueue(delayed, queue.Name, retryWait)
			return err
		}
		if again {
			continue
		}
		if wait > 0 {
			f.deferQueue(delayed, queue.Name, wait)
		}
		return nil
	}
}

// feedOne 最多推进一条。again 为真表示还能继续。wait 大于 0 表示先等一会儿再试。
func (f Feeder) feedOne(ctx context.Context, queue taskrun.Queue) (bool, time.Duration, error) {
	paused, err := f.Tasks.Paused(ctx, queue.Name)
	if err != nil {
		return false, 0, err
	}
	if paused {
		return false, pauseWait, nil
	}
	waiting, err := f.Tasks.Len(ctx, queue.Name)
	if err != nil {
		return false, 0, err
	}
	if queue.Waiting > 0 && waiting >= int64(queue.Waiting) {
		return false, retryWait, nil
	}
	store := jobs.Store{DB: f.DB}
	now := f.now()
	job, found, err := store.TakePending(ctx, queue.Name, now)
	if err != nil || !found {
		return false, 0, err
	}
	ok, err := f.Tasks.Take(ctx, queue.Name, queue.Every, queue.Limit)
	if err != nil {
		_, _ = store.MarkPending(ctx, job.ID, job.RetryCount, "", now)
		return false, 0, err
	}
	if !ok {
		_, _ = store.MarkPending(ctx, job.ID, job.RetryCount, "", now)
		return false, retryWait, nil
	}
	err = f.Tasks.Push(ctx, queue.Name, taskrun.Message{
		ID:      job.ID,
		Type:    job.Type,
		Payload: job.Payload,
	})
	if err != nil {
		_, _ = store.MarkPending(ctx, job.ID, job.RetryCount, "", now)
		return false, 0, err
	}
	return true, 0, nil
}

func (f Feeder) deferQueue(delayed map[string]time.Time, name string, wait time.Duration) {
	if delayed == nil || wait <= 0 {
		return
	}
	delayed[name] = f.now().Add(wait)
}

func (f Feeder) queueByName(name string) (taskrun.Queue, bool) {
	for _, queue := range f.Queues {
		if queue.Feed && queue.Name == name {
			return queue, true
		}
	}
	return taskrun.Queue{}, false
}

func (f Feeder) now() time.Time {
	if f.Now != nil {
		return f.Now()
	}
	return dbtime.Now()
}

func feedNames(queues []taskrun.Queue) []string {
	names := make([]string, 0, len(queues))
	for _, queue := range queues {
		if queue.Feed {
			names = append(names, queue.Name)
		}
	}
	return names
}

func dueQueue(delayed map[string]time.Time, now time.Time) (string, bool) {
	var name string
	var at time.Time
	for queue, when := range delayed {
		if when.After(now) {
			continue
		}
		if name == "" || when.Before(at) {
			name = queue
			at = when
		}
	}
	return name, name != ""
}

func soonerDelay(delayed map[string]time.Time, now time.Time) (time.Duration, bool) {
	var wait time.Duration
	found := false
	for _, when := range delayed {
		d := when.Sub(now)
		if d < 0 {
			d = 0
		}
		if !found || d < wait {
			wait = d
			found = true
		}
	}
	return wait, found
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

var errIncomplete = errString("feeder is incomplete")

type errString string

func (e errString) Error() string { return string(e) }
