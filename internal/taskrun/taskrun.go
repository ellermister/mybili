package taskrun

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

const (
	// QueueBiliFast 放短请求：收藏夹列表和收藏夹视频。
	QueueBiliFast = "bili-fast"
	// QueueBiliDefault 放耗时下载：视频和弹幕。
	QueueBiliDefault = "bili-default"
	// QueueCover 放封面和头像，不经过 Feeder。
	QueueCover = "cover"
	// QueueNotify 放通知，不限流。
	QueueNotify = "notify"
	// QueueLocal 放本机维护，不请求 Bilibili，也不占用接口限速。
	QueueLocal = "local"

	// LimitAtFeeder 表示名额在投喂前扣除。
	LimitAtFeeder = "feeder"
	// LimitAtWorker 表示名额在执行前扣除。
	LimitAtWorker = "worker"
)

// ErrCannotRetry 表示这条消息不要再推回队列。
var ErrCannotRetry = errors.New("cannot retry")

// ErrLockLost 表示执行租约已经不是本次持有者。
var ErrLockLost = errors.New("task lock lost")

// Queue 是一个 Redis 列表的消费配置。数字是部署时定的，业务只选名字。
type Queue struct {
	Name         string
	Concurrent   int
	Waiting      int
	Every        time.Duration
	Limit        int
	Timeout      time.Duration
	Feed         bool
	LimitAt      string
	PauseWithAPI bool
}

// AllQueues 返回 worker 要监听的全部队列。
func AllQueues() []Queue {
	return []Queue{
		{Name: QueueBiliFast, Concurrent: 4, Waiting: 4, Every: 5 * time.Second, Limit: 2, Timeout: 30 * time.Minute, Feed: true, LimitAt: LimitAtFeeder, PauseWithAPI: true},
		{Name: QueueBiliDefault, Concurrent: 4, Waiting: 4, Every: 12 * time.Second, Limit: 2, Timeout: 2 * time.Hour, Feed: true, LimitAt: LimitAtFeeder, PauseWithAPI: true},
		{Name: QueueCover, Concurrent: 1, Every: time.Minute, Limit: 100, Timeout: time.Minute, Feed: false, LimitAt: LimitAtWorker, PauseWithAPI: true},
		{Name: QueueNotify, Concurrent: 1, Timeout: 15 * time.Second, Feed: false},
		{Name: QueueLocal, Concurrent: 1, Waiting: 1, Timeout: 2 * time.Hour, Feed: true},
	}
}

// FeedQueues 返回要从 tasks 表投喂的队列。
func FeedQueues() []Queue {
	var queues []Queue
	for _, queue := range AllQueues() {
		if queue.Feed {
			queues = append(queues, queue)
		}
	}
	return queues
}

// Message 是 Redis 里的一条任务。Payload 由业务自己解析。
type Message struct {
	ID      int64  `json:"id,omitempty"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Attempt int    `json:"attempt,omitempty"`
}

// Handler 执行一条消息。worker 不解析 Payload。
type Handler func(ctx context.Context, msg Message) error

func encode(msg Message) (string, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func decode(raw string) (Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(raw), &msg)
	return msg, err
}
