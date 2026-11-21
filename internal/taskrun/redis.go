package taskrun

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"mybili/internal/config"
)

// Tasks 是 Redis 上的队列。它不查询 tasks 表。
type Tasks struct {
	redis    *redis.Client
	handlers map[string]Handler
}

// New 给 server 和命令使用。命令很快返回，读超时设短。
func New(cfg config.Config) *Tasks {
	return newTasks(cfg, 2*time.Second)
}

// NewBlocking 给 worker 使用，避免 BRPOP 被客户端读超时打断。
func NewBlocking(cfg config.Config) *Tasks {
	return newTasks(cfg, 0)
}

func newTasks(cfg config.Config, readTimeout time.Duration) *Tasks {
	return &Tasks{redis: redis.NewClient(&redis.Options{
		Addr:         net.JoinHostPort(cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  readTimeout,
		WriteTimeout: 2 * time.Second,
	})}
}

// Wake 通知 Feeder 这个队列有 pending。任务正文仍在 tasks 表。
func (t *Tasks) Wake(ctx context.Context, queue string) error {
	if queue == "" {
		return fmt.Errorf("queue is empty")
	}
	return t.redis.LPush(ctx, feedWakeKey, queue).Err()
}

// NextWake 阻塞直到收到一个队列名。超时没有消息时 ok 为 false。
func (t *Tasks) NextWake(ctx context.Context, wait time.Duration) (string, bool, error) {
	if wait <= 0 {
		wait = time.Second
	}
	values, err := t.redis.BRPop(ctx, wait, feedWakeKey).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if len(values) != 2 {
		return "", false, fmt.Errorf("feed wake returned %d values", len(values))
	}
	return values[1], true, nil
}

// Push 把消息直接放到执行队列左侧。这条路径不写 tasks，也不经过 Feeder。
func (t *Tasks) Push(ctx context.Context, queue string, msg Message) error {
	raw, err := encode(msg)
	if err != nil {
		return err
	}
	return t.redis.LPush(ctx, listKey(queue), raw).Err()
}

// PushDedup 在去重键上的正文没变时不推送。变了才把 msg 放进 queue。
// 队列名、消息类型和去重时间由调用方决定。
func (t *Tasks) PushDedup(ctx context.Context, queue string, msg Message, dedupKey string, ttl time.Duration) (bool, error) {
	if t == nil || t.redis == nil || queue == "" || dedupKey == "" || ttl <= 0 {
		return false, fmt.Errorf("dedup push is incomplete")
	}
	key := dedupRedisKey(dedupKey)
	prev, err := t.redis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if err == nil && prev == msg.Payload {
		return false, nil
	}
	if err := t.redis.Set(ctx, key, msg.Payload, ttl).Err(); err != nil {
		return false, err
	}
	err = t.Push(ctx, queue, msg)
	if err != nil {
		_ = t.redis.Del(ctx, key).Err()
		return false, err
	}
	return true, nil
}

// Forget 删掉去重键，之后同一条消息可以再推。
func (t *Tasks) Forget(ctx context.Context, dedupKey string) error {
	if t == nil || t.redis == nil || dedupKey == "" {
		return nil
	}
	return t.redis.Del(ctx, dedupRedisKey(dedupKey)).Err()
}

// MarkOnce 在时限内只成功一次。已经有标记时 ok 为 false。
func (t *Tasks) MarkOnce(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if t == nil || t.redis == nil || key == "" {
		return false, nil
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return t.redis.SetNX(ctx, key, "1", ttl).Result()
}

// OnceValue 在键不存在时写入 value 并返回它。键已经存在时返回里面的值。
func (t *Tasks) OnceValue(ctx context.Context, key, value string, ttl time.Duration) (string, error) {
	if t == nil || t.redis == nil || key == "" {
		return "", fmt.Errorf("redis is missing")
	}
	if ttl <= 0 {
		ttl = 365 * 24 * time.Hour
	}
	ok, err := t.redis.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return "", err
	}
	if ok {
		return value, nil
	}
	got, err := t.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("value disappeared")
	}
	return got, err
}

// Unmark 去掉 MarkOnce 写下的标记。
func (t *Tasks) Unmark(ctx context.Context, key string) error {
	if t == nil || t.redis == nil || key == "" {
		return nil
	}
	return t.redis.Del(ctx, key).Err()
}

// Len 返回队列里还没取走的条数。
func (t *Tasks) Len(ctx context.Context, queue string) (int64, error) {
	return t.redis.LLen(ctx, listKey(queue)).Result()
}

// Pop 从队列右侧取一条。超时没有消息时 ok 为 false。
func (t *Tasks) Pop(ctx context.Context, queue Queue, wait time.Duration) (Message, string, bool, error) {
	values, err := t.redis.BRPop(ctx, wait, listKey(queue.Name)).Result()
	if err == redis.Nil {
		return Message{}, "", false, nil
	}
	if err != nil {
		return Message{}, "", false, err
	}
	if len(values) != 2 {
		return Message{}, "", false, fmt.Errorf("task pop returned %d values", len(values))
	}
	raw := values[1]
	msg, err := decode(raw)
	if err != nil {
		return Message{}, "", false, err
	}
	if !queue.Feed {
		deadline := time.Now().Add(activeTTL(queue)).UnixMilli()
		if err := t.redis.ZAdd(ctx, activeKey(queue.Name), redis.Z{Score: float64(deadline), Member: raw}).Err(); err != nil {
			return Message{}, "", false, err
		}
	}
	return msg, raw, true, nil
}

// Ack 确认一条直接投递的消息，收回逻辑不会再把它推回队列。
func (t *Tasks) Ack(ctx context.Context, queue, raw string) error {
	if raw == "" {
		return nil
	}
	return t.redis.ZRem(ctx, activeKey(queue), raw).Err()
}

// RequeueExpired 把超过期限还没确认的直接消息推回原队列。
func (t *Tasks) RequeueExpired(ctx context.Context, queue string) error {
	n, err := requeueExpiredScript.Run(ctx, t.redis, []string{activeKey(queue), listKey(queue)}, time.Now().UnixMilli()).Int()
	if err != nil && err != redis.Nil {
		return err
	}
	_ = n
	return nil
}

// Paused 查看队列是否停止取新消息。
func (t *Tasks) Paused(ctx context.Context, queue string) (bool, error) {
	n, err := t.redis.Exists(ctx, pauseKey(queue)).Result()
	return n > 0, err
}

// Pause 在一段时间内停止这个队列取新消息。已经在执行的不受影响。
func (t *Tasks) Pause(ctx context.Context, queue string, d time.Duration) error {
	if d <= 0 {
		d = 2 * time.Hour
	}
	return t.redis.Set(ctx, pauseKey(queue), "1", d).Err()
}

// Resume 马上恢复一个队列。
func (t *Tasks) Resume(ctx context.Context, queue string) error {
	return t.redis.Del(ctx, pauseKey(queue)).Err()
}

// Take 尝试占用滚动窗口里的一个名额。窗口未配置时直接成功。
func (t *Tasks) Take(ctx context.Context, queue string, every time.Duration, limit int) (bool, error) {
	if every <= 0 || limit <= 0 {
		return true, nil
	}
	member, err := newToken()
	if err != nil {
		return false, err
	}
	n, err := takeRateScript.Run(ctx, t.redis, []string{rateKey(queue)}, every.Milliseconds(), limit, member).Int()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// WaitRate 一直等到这个队列的窗口有名额。
func (t *Tasks) WaitRate(ctx context.Context, queue Queue) error {
	for {
		ok, err := t.Take(ctx, queue.Name, queue.Every, queue.Limit)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// Lease 占用一条落表任务的执行租约。已经被占住时 ok 为 false。
func (t *Tasks) Lease(ctx context.Context, id int64, ttl time.Duration) (string, bool, error) {
	if id <= 0 {
		return "", true, nil
	}
	token, err := newToken()
	if err != nil {
		return "", false, err
	}
	ok, err := t.redis.SetNX(ctx, leaseKey(id), token, ttl).Result()
	if err != nil || !ok {
		return "", false, err
	}
	return token, true, nil
}

// Keep 延长自己的租约。租约已经换人时返回 ErrLockLost。
func (t *Tasks) Keep(ctx context.Context, id int64, token string, ttl time.Duration) error {
	if id <= 0 {
		return nil
	}
	n, err := keepLeaseScript.Run(ctx, t.redis, []string{leaseKey(id)}, token, ttl.Milliseconds()).Int64()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrLockLost
	}
	return nil
}

// Release 放开自己的租约。令牌对不上时不删除。
func (t *Tasks) Release(ctx context.Context, id int64, token string) error {
	if id <= 0 || token == "" {
		return nil
	}
	return unlockScript.Run(ctx, t.redis, []string{leaseKey(id)}, token).Err()
}

// Held 查看落表任务的租约是否还在。
func (t *Tasks) Held(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	n, err := t.redis.Exists(ctx, leaseKey(id)).Result()
	return n > 0, err
}

// TouchActive 把直接消息的收回期限往后推。
func (t *Tasks) TouchActive(ctx context.Context, queue, raw string, ttl time.Duration) error {
	if raw == "" {
		return nil
	}
	score := float64(time.Now().Add(activeTTL(Queue{Timeout: ttl})).UnixMilli())
	return t.redis.ZAdd(ctx, activeKey(queue), redis.Z{Score: score, Member: raw}).Err()
}

const feedWakeKey = "mybili:feed"

func listKey(queue string) string  { return "mybili:tasks:" + queue }
func pauseKey(queue string) string { return "mybili:queue:" + queue + ":pause" }
func rateKey(queue string) string  { return "mybili:rate:" + queue }
func activeKey(queue string) string {
	return "mybili:active:" + queue
}
func leaseKey(id int64) string { return fmt.Sprintf("mybili:task:%d", id) }
func dedupRedisKey(key string) string {
	return "mybili:dedup:" + key
}

func activeTTL(queue Queue) time.Duration {
	if queue.Timeout > 0 {
		return queue.Timeout
	}
	return 2 * time.Minute
}

func newToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

var takeRateScript = redis.NewScript(`
local key = KEYS[1]
local window_ms = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local member = ARGV[3]
local now = redis.call('TIME')
local now_ms = tonumber(now[1]) * 1000 + math.floor(tonumber(now[2]) / 1000)
redis.call('ZREMRANGEBYSCORE', key, '-inf', now_ms - window_ms)
local count = redis.call('ZCARD', key)
if count >= limit then
  return 0
end
redis.call('ZADD', key, now_ms, member)
redis.call('PEXPIRE', key, window_ms)
return 1
`)

var keepLeaseScript = redis.NewScript(`
if redis.call('get', KEYS[1]) == ARGV[1] then
  return redis.call('pexpire', KEYS[1], ARGV[2])
end
return 0
`)

var unlockScript = redis.NewScript(`
if redis.call('get', KEYS[1]) == ARGV[1] then
  return redis.call('del', KEYS[1])
end
return 0
`)

var requeueExpiredScript = redis.NewScript(`
local items = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
for _, item in ipairs(items) do
  redis.call('ZREM', KEYS[1], item)
  redis.call('LPUSH', KEYS[2], item)
end
return #items
`)
