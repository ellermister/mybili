package taskrun

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 任务输出保留 12 小时，到期由 Redis 删掉。
	outputTTL = 12 * time.Hour
	// 单条任务最多留这么多行，避免一次下载把内存撑满。
	outputMaxLines = 4000

	taskLogKeyPrefix  = "mybili:tasklog:"
	taskProgKeyPrefix = "mybili:taskprog:"
)

// Progress 是任务当前进度。Total 为 0 表示还没有百分比。
type Progress struct {
	Done  int    `json:"done"`
	Total int    `json:"total"`
	Text  string `json:"text"`
}

func taskLogKey(id int64) string {
	return taskLogKeyPrefix + strconv.FormatInt(id, 10)
}

func taskProgKey(id int64) string {
	return taskProgKeyPrefix + strconv.FormatInt(id, 10)
}

// AppendLog 追加一行标准输出。键是 mybili:tasklog:{任务 id}。
func (t *Tasks) AppendLog(ctx context.Context, id int64, line string) error {
	if t == nil || t.redis == nil || id <= 0 || line == "" {
		return nil
	}
	key := taskLogKey(id)
	pipe := t.redis.Pipeline()
	pipe.RPush(ctx, key, line)
	pipe.LTrim(ctx, key, -outputMaxLines, -1)
	pipe.Expire(ctx, key, outputTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// SetProgress 写下当前进度。键是 mybili:taskprog:{任务 id}，和日志一起过期。
func (t *Tasks) SetProgress(ctx context.Context, id int64, progress Progress) error {
	if t == nil || t.redis == nil || id <= 0 {
		return nil
	}
	body, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	return t.redis.Set(ctx, taskProgKey(id), body, outputTTL).Err()
}

// TaskOutput 读出这次任务留下的输出和进度。过期后两边都是空的。
func (t *Tasks) TaskOutput(ctx context.Context, id int64) ([]string, Progress, error) {
	if t == nil || t.redis == nil || id <= 0 {
		return []string{}, Progress{}, nil
	}
	lines, err := t.redis.LRange(ctx, taskLogKey(id), 0, -1).Result()
	if err != nil {
		return nil, Progress{}, err
	}
	if lines == nil {
		lines = []string{}
	}
	raw, err := t.redis.Get(ctx, taskProgKey(id)).Result()
	if err == redis.Nil {
		return lines, Progress{}, nil
	}
	if err != nil {
		return nil, Progress{}, err
	}
	var progress Progress
	if err := json.Unmarshal([]byte(raw), &progress); err != nil {
		return lines, Progress{}, nil
	}
	return lines, progress, nil
}
