package feeder

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/migrate"
	"mybili/internal/model"
	"mybili/internal/taskrun"
)

func TestTickPushesPendingThenStopsWithoutToken(t *testing.T) {
	conn := openFeederDB(t)
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, dbtime.Location())
	row := model.Job{
		Queue: "bili-fast", Type: "fav-list", UniqueKey: "sync:fav-list",
		Status: jobs.StatusPending, Payload: "{}", MaxRetries: 3,
	}
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	broker := &fakeBroker{tokens: 1}
	feeder := Feeder{
		DB:     conn,
		Tasks:  broker,
		Queues: []taskrun.Queue{{Name: "bili-fast", Waiting: 4, Every: 5 * time.Second, Limit: 2, Feed: true}},
		Now:    func() time.Time { return now },
	}
	if err := feeder.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(broker.pushed) != 1 || broker.pushed[0].ID != row.ID || broker.pushed[0].Payload != "{}" {
		t.Fatalf("pushed = %+v", broker.pushed)
	}
	saved := reload(t, conn, row.ID)
	if saved.Status != jobs.StatusQueued {
		t.Fatalf("status = %s", saved.Status)
	}
	if err := feeder.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(broker.pushed) != 1 {
		t.Fatalf("pushed again = %+v", broker.pushed)
	}
}

func TestTickReturnsRowWhenRateMisses(t *testing.T) {
	conn := openFeederDB(t)
	row := model.Job{
		Queue: "bili-fast", Type: "fav-list", UniqueKey: "sync:fav-list",
		Status: jobs.StatusPending, Payload: "{}", MaxRetries: 3,
	}
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	broker := &fakeBroker{}
	feeder := Feeder{
		DB:     conn,
		Tasks:  broker,
		Queues: []taskrun.Queue{{Name: "bili-fast", Waiting: 4, Every: time.Second, Limit: 1, Feed: true}},
	}
	if err := feeder.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(broker.pushed) != 0 {
		t.Fatalf("pushed = %+v", broker.pushed)
	}
	saved := reload(t, conn, row.ID)
	if saved.Status != jobs.StatusPending {
		t.Fatalf("status = %s", saved.Status)
	}
}

func TestReapResetsRunningWithoutLease(t *testing.T) {
	conn := openFeederDB(t)
	running := model.Job{
		Queue: "bili-fast", Type: "fav-list", UniqueKey: "sync:fav-list",
		Status: jobs.StatusRunning, Payload: "{}", MaxRetries: 3,
	}
	held := model.Job{
		Queue: "bili-fast", Type: "fav-videos", UniqueKey: "sync:fav-videos:1",
		Status: jobs.StatusRunning, Payload: `{"page":1}`, MaxRetries: 3,
	}
	if err := conn.Create(&running).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&held).Error; err != nil {
		t.Fatal(err)
	}
	broker := &fakeBroker{held: map[int64]bool{held.ID: true}}
	feeder := Feeder{
		DB:     conn,
		Tasks:  broker,
		Queues: taskrun.FeedQueues(),
	}
	if err := feeder.Reap(context.Background()); err != nil {
		t.Fatal(err)
	}
	if reload(t, conn, running.ID).Status != jobs.StatusPending {
		t.Fatal("running was not reset")
	}
	if reload(t, conn, held.ID).Status != jobs.StatusRunning {
		t.Fatal("held task was reset")
	}
	if len(broker.wakes) != 1 || broker.wakes[0] != "bili-fast" {
		t.Fatalf("wakes = %v", broker.wakes)
	}
}

type fakeBroker struct {
	paused  bool
	waiting int64
	tokens  int
	pushed  []taskrun.Message
	held    map[int64]bool
	wakes   []string
}

func (f *fakeBroker) Paused(context.Context, string) (bool, error) { return f.paused, nil }

func (f *fakeBroker) Len(context.Context, string) (int64, error) { return f.waiting, nil }

func (f *fakeBroker) Take(context.Context, string, time.Duration, int) (bool, error) {
	if f.tokens <= 0 {
		return false, nil
	}
	f.tokens--
	return true, nil
}

func (f *fakeBroker) Push(_ context.Context, _ string, msg taskrun.Message) error {
	f.pushed = append(f.pushed, msg)
	return nil
}

func (f *fakeBroker) Held(_ context.Context, id int64) (bool, error) {
	return f.held[id], nil
}

func (f *fakeBroker) Wake(_ context.Context, queue string) error {
	f.wakes = append(f.wakes, queue)
	return nil
}

func (f *fakeBroker) NextWake(context.Context, time.Duration) (string, bool, error) {
	return "", false, nil
}

func openFeederDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "feeder.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func reload(t *testing.T, conn *gorm.DB, id int64) model.Job {
	t.Helper()
	var job model.Job
	if err := conn.First(&job, id).Error; err != nil {
		t.Fatal(err)
	}
	return job
}
