package syncbili

import (
	"context"
	"sort"
	"sync"
	"testing"
)

// memCoverPush 记下封面去重键和正文，测试不连 Redis。
type memCoverPush struct {
	mu    sync.Mutex
	items map[string]string
}

func (m *memCoverPush) PushDedup(_ context.Context, dedupKey, payload string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items == nil {
		m.items = map[string]string{}
	}
	if m.items[dedupKey] == payload {
		return false, nil
	}
	m.items[dedupKey] = payload
	return true, nil
}

func (m *memCoverPush) Forget(_ context.Context, dedupKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, dedupKey)
	return nil
}

func (m *memCoverPush) get(key string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.items[key]
}

func (m *memCoverPush) keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.items))
	for key := range m.items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func coverPushOf(t *testing.T, syncer Sync) *memCoverPush {
	t.Helper()
	push, ok := syncer.CoverPush.(*memCoverPush)
	if !ok || push == nil {
		t.Fatal("cover push missing")
	}
	return push
}
