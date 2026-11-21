package jobs

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/migrate"
	"mybili/internal/model"
)

func TestSaveSkipsActiveAndReopensFinished(t *testing.T) {
	conn := openJobDB(t)
	store := Store{DB: conn}
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, dbtime.Location())
	draft := Draft{UniqueKey: "sync:fav-list", Queue: "bili-fast", Type: "fav-list", Payload: "{}"}

	created, err := store.Save(context.Background(), draft, now, SaveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !created.Opened || !created.Wake || created.ID == 0 {
		t.Fatalf("created = %+v", created)
	}
	again, err := store.Save(context.Background(), draft, now, SaveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if again.Opened || !again.Wake || again.ID != created.ID {
		t.Fatalf("pending = %+v", again)
	}

	if err := conn.Model(&model.Job{}).Where("id = ?", created.ID).Update("status", StatusQueued).Error; err != nil {
		t.Fatal(err)
	}
	queued, err := store.Save(context.Background(), draft, now, SaveOptions{})
	if err != nil || queued.Opened || queued.Wake {
		t.Fatalf("queued = %+v err=%v", queued, err)
	}

	if err := conn.Model(&model.Job{}).Where("id = ?", created.ID).Update("status", StatusRunning).Error; err != nil {
		t.Fatal(err)
	}
	running, err := store.Save(context.Background(), draft, now, SaveOptions{})
	if err != nil || running.Opened || running.Wake {
		t.Fatalf("running = %+v err=%v", running, err)
	}

	stamp := dbtime.DateTime{Time: now}
	err = conn.Model(&model.Job{}).Where("id = ?", created.ID).Updates(map[string]any{
		"status": StatusDone, "updated_at": stamp,
	}).Error
	if err != nil {
		t.Fatal(err)
	}
	held, err := store.Save(context.Background(), draft, now.Add(time.Minute), SaveOptions{ReopenAfter: now})
	if err != nil || held.Opened {
		t.Fatalf("inside slot = %+v err=%v", held, err)
	}
	opened, err := store.Save(context.Background(), draft, now.Add(10*time.Minute), SaveOptions{ReopenAfter: now.Add(10 * time.Minute)})
	if err != nil || !opened.Opened || opened.ID != created.ID {
		t.Fatalf("reopen = %+v err=%v", opened, err)
	}
}

func TestTakePendingAndReapStatuses(t *testing.T) {
	conn := openJobDB(t)
	store := Store{DB: conn}
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, dbtime.Location())
	draft := Draft{UniqueKey: "sync:fav-list", Queue: "bili-fast", Type: "fav-list", Payload: "{}"}
	saved, err := store.Save(context.Background(), draft, now, SaveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	job, found, err := store.TakePending(context.Background(), "bili-fast", now)
	if err != nil || !found || job.Status != StatusQueued {
		t.Fatalf("take = %+v found=%v err=%v", job, found, err)
	}
	ok, err := store.MarkRunning(context.Background(), saved.ID, now)
	if err != nil || !ok {
		t.Fatalf("running ok=%v err=%v", ok, err)
	}
	ok, err = store.MarkPending(context.Background(), saved.ID, 1, "again", now)
	if err != nil || !ok {
		t.Fatalf("pending ok=%v err=%v", ok, err)
	}
	loaded, err := store.Load(context.Background(), saved.ID)
	if err != nil || loaded.Status != StatusPending || loaded.RetryCount != 1 || loaded.StartedAt != nil {
		t.Fatalf("loaded = %+v err=%v", loaded, err)
	}
}

func openJobDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "jobs.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}
