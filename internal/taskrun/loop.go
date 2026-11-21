package taskrun

import (
	"context"
	"errors"
	"log"
	"time"
)

const directAttempts = 3

// Loop 按队列各开一条循环，直到 ctx 结束。
func (t *Tasks) Loop(ctx context.Context, queues []Queue) error {
	if t == nil || t.redis == nil {
		return errors.New("task runner is incomplete")
	}
	for _, queue := range queues {
		if !queue.Feed {
			if err := t.RequeueExpired(ctx, queue.Name); err != nil {
				log.Printf("收回直接任务失败 %s: %v", queue.Name, err)
			}
		}
		go t.loop(ctx, queue)
	}
	<-ctx.Done()
	return ctx.Err()
}

// Handle 登记一种消息的处理函数。
func (t *Tasks) Handle(taskType string, handler Handler) {
	if t.handlers == nil {
		t.handlers = map[string]Handler{}
	}
	t.handlers[taskType] = handler
}

func (t *Tasks) loop(ctx context.Context, queue Queue) {
	slots := queue.Concurrent
	if slots < 1 {
		slots = 1
	}
	free := make(chan struct{}, slots)
	sweep := time.NewTicker(30 * time.Second)
	defer sweep.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		if !queue.Feed {
			select {
			case <-sweep.C:
				if err := t.RequeueExpired(ctx, queue.Name); err != nil {
					log.Printf("收回直接任务失败 %s: %v", queue.Name, err)
				}
			default:
			}
		}
		paused, err := t.Paused(ctx, queue.Name)
		if err != nil {
			log.Printf("读取队列暂停失败 %s: %v", queue.Name, err)
			if sleep(ctx, time.Second) != nil {
				return
			}
			continue
		}
		if paused {
			if sleep(ctx, time.Second) != nil {
				return
			}
			continue
		}
		free <- struct{}{}
		msg, raw, ok, err := t.Pop(ctx, queue, 5*time.Second)
		if err != nil || !ok {
			<-free
			if err != nil && ctx.Err() == nil {
				log.Printf("取任务失败 %s: %v", queue.Name, err)
				if sleep(ctx, time.Second) != nil {
					return
				}
			}
			continue
		}
		go func(msg Message, raw string) {
			defer func() { <-free }()
			t.run(ctx, queue, msg, raw)
		}(msg, raw)
	}
}

func (t *Tasks) run(ctx context.Context, queue Queue, msg Message, raw string) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	token, ok, err := t.Lease(runCtx, msg.ID, activeTTL(queue))
	if err != nil || !ok {
		if !queue.Feed {
			_ = t.Ack(ctx, queue.Name, raw)
			_ = t.Push(ctx, queue.Name, msg)
		}
		return
	}
	defer t.Release(context.Background(), msg.ID, token)
	stop := make(chan struct{})
	defer close(stop)
	go t.touch(runCtx, stop, queue, msg, raw, token)

	if queue.LimitAt == LimitAtWorker {
		if err := t.WaitRate(runCtx, queue); err != nil {
			_ = t.Ack(context.Background(), queue.Name, raw)
			_ = t.Push(context.Background(), queue.Name, msg)
			return
		}
	}
	handler := t.handlers[msg.Type]
	if handler == nil {
		log.Printf("没有处理函数 type=%s", msg.Type)
		_ = t.Ack(context.Background(), queue.Name, raw)
		return
	}
	timeout := queue.Timeout
	if timeout <= 0 {
		timeout = time.Minute
	}
	callCtx, cancelCall := context.WithTimeout(runCtx, timeout)
	err = handler(callCtx, msg)
	cancelCall()
	_ = t.Ack(context.Background(), queue.Name, raw)
	if err == nil || errors.Is(err, ErrCannotRetry) {
		return
	}
	if queue.Feed {
		return
	}
	if msg.Attempt+1 >= directAttempts {
		log.Printf("直接任务放弃 type=%s attempt=%d: %v", msg.Type, msg.Attempt, err)
		return
	}
	msg.Attempt++
	if err := t.Push(context.Background(), queue.Name, msg); err != nil {
		log.Printf("直接任务推回失败 type=%s: %v", msg.Type, err)
	}
}

func (t *Tasks) touch(ctx context.Context, stop <-chan struct{}, queue Queue, msg Message, raw, token string) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if msg.ID > 0 {
				if err := t.Keep(context.Background(), msg.ID, token, activeTTL(queue)); err != nil {
					log.Printf("延长任务租约 %d 失败: %v", msg.ID, err)
				}
			}
			if !queue.Feed {
				if err := t.TouchActive(context.Background(), queue.Name, raw, queue.Timeout); err != nil {
					log.Printf("延长直接任务期限失败 %s: %v", queue.Name, err)
				}
			}
		}
	}
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
