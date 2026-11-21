package schedule

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
)

func TestTickSkipsDisabledOrNotDue(t *testing.T) {
	conn := openScheduleDB(t)
	runner := Runner{Specs: []Spec{alwaysSpec(false)}, Store: jobs.Store{DB: conn}}
	if err := runner.Tick(context.Background(), shanghai(2026, 10, 30, 10, 0)); err != nil {
		t.Fatal(err)
	}
	due := alwaysSpec(true)
	due.Due = func(time.Time) bool { return false }
	runner.Specs = []Spec{due}
	if err := runner.Tick(context.Background(), shanghai(2026, 10, 30, 10, 3)); err != nil {
		t.Fatal(err)
	}
	if n := countTasks(t, conn); n != 0 {
		t.Fatalf("tasks=%d", n)
	}
}

func TestTickDoesNotDuplicatePending(t *testing.T) {
	conn := openScheduleDB(t)
	var wakes []string
	runner := Runner{
		Specs: []Spec{alwaysSpec(true)},
		Store: jobs.Store{DB: conn},
		Wake: func(_ context.Context, queue string) error {
			wakes = append(wakes, queue)
			return nil
		},
	}
	now := shanghai(2026, 10, 30, 10, 0)
	if err := runner.Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if err := runner.Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if n := countTasks(t, conn); n != 1 || len(wakes) != 2 {
		t.Fatalf("tasks=%d wakes=%v", n, wakes)
	}
}

func TestTickDoesNotReopenFinishedInsideSlot(t *testing.T) {
	conn := openScheduleDB(t)
	var wakes int
	runner := Runner{
		Specs: []Spec{alwaysSpec(true)},
		Store: jobs.Store{DB: conn},
		Wake: func(context.Context, string) error {
			wakes++
			return nil
		},
	}
	now := shanghai(2026, 10, 30, 10, 0)
	if err := runner.Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	stamp := dbtime.DateTime{Time: now}
	err := conn.Model(&model.Job{}).Where("unique_key = ?", "tool:once").Updates(map[string]any{
		"status": jobs.StatusDone, "updated_at": stamp,
	}).Error
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	var job model.Job
	if err := conn.Where("unique_key = ?", "tool:once").First(&job).Error; err != nil {
		t.Fatal(err)
	}
	if job.Status != jobs.StatusDone || wakes != 1 {
		t.Fatalf("status = %s wakes=%d", job.Status, wakes)
	}
}

func TestTickReopensFinishedBeforeSlot(t *testing.T) {
	conn := openScheduleDB(t)
	runner := Runner{Specs: []Spec{alwaysSpec(true)}, Store: jobs.Store{DB: conn}}
	now := shanghai(2026, 10, 30, 10, 0)
	if err := runner.Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	stamp := dbtime.DateTime{Time: now}
	err := conn.Model(&model.Job{}).Where("unique_key = ?", "tool:once").Updates(map[string]any{
		"status": jobs.StatusDone, "updated_at": stamp,
	}).Error
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Tick(context.Background(), now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var job model.Job
	if err := conn.Where("unique_key = ?", "tool:once").First(&job).Error; err != nil {
		t.Fatal(err)
	}
	if job.Status != jobs.StatusPending {
		t.Fatalf("status = %s", job.Status)
	}
}

func alwaysSpec(on bool) Spec {
	return Spec{
		Name: "tool",
		Due:  func(time.Time) bool { return true },
		Enabled: func(context.Context) (bool, error) {
			return on, nil
		},
		Tasks: func(_ context.Context, now time.Time) ([]Draft, error) {
			return []Draft{{
				Queue:       "bili-fast",
				Type:        "tool",
				UniqueKey:   "tool:once",
				Payload:     "{}",
				ReopenAfter: now.Truncate(time.Minute),
			}}, nil
		},
	}
}

func openScheduleDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "schedule.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func countTasks(t *testing.T, conn *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := conn.Model(&model.Job{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func shanghai(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, dbtime.Location())
}
