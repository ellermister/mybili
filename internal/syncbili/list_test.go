package syncbili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/migrate"
	"mybili/internal/model"
)

func TestSyncUpsertsFoldersWithoutTouchingOthers(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 29, 3, 0, 0, 0, time.Local)
	existing := model.FavoriteList{
		ID: 11, Title: "旧标题", Cover: "old", CacheImage: "keep-me",
		Ctime: stamp(1499900654), Mtime: stamp(1523352284), MediaCount: 1,
	}
	other := model.FavoriteList{
		ID: 99, Title: "别动", Cover: "x", CacheImage: "other",
		Ctime: stamp(1), Mtime: stamp(2), MediaCount: 4,
	}
	if err := conn.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&other).Error; err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"list":[{"id":11,"title":"新标题","cover":"new","ctime":1499900654,"mtime":1523352284,"media_count":8},{"id":12,"title":"新建","cover":"c","ctime":1743737512,"mtime":1743737512,"media_count":1}],"has_more":false}}`))
	}))
	defer server.Close()

	cookies := []bilibili.Cookie{{Name: "DedeUserID", Value: "7"}}
	client := &bilibili.Client{BaseURL: server.URL, CookieHeader: "DedeUserID=7", HTTP: server.Client()}
	count, err := newTestSync(conn, client, cookies, now, 0).Folders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d", count)
	}

	var updated model.FavoriteList
	if err := conn.First(&updated, 11).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Title != "新标题" || updated.Cover != "new" || updated.MediaCount != 8 || updated.CacheImage != "keep-me" {
		t.Fatalf("updated = %+v", updated)
	}
	if updated.Ctime.Unix() != 1499900654 {
		t.Fatalf("ctime = %d", updated.Ctime.Unix())
	}

	var created model.FavoriteList
	if err := conn.First(&created, 12).Error; err != nil {
		t.Fatal(err)
	}
	if created.Title != "新建" || created.CacheImage != "" {
		t.Fatalf("created = %+v", created)
	}

	var untouched model.FavoriteList
	if err := conn.First(&untouched, 99).Error; err != nil {
		t.Fatal(err)
	}
	if untouched.Title != "别动" || untouched.CacheImage != "other" {
		t.Fatalf("untouched = %+v", untouched)
	}
}

func TestSyncStopsOnRiskControl(t *testing.T) {
	conn := openTestDB(t)
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusPreconditionFailed)
		_, _ = w.Write([]byte(`{"code":-412,"message":"banned"}`))
	}))
	defer server.Close()
	now := time.Date(2026, 10, 29, 3, 0, 0, 0, time.UTC)
	cookies := []bilibili.Cookie{{Name: "DedeUserID", Value: "7"}}
	client := &bilibili.Client{BaseURL: server.URL, CookieHeader: "DedeUserID=7", HTTP: server.Client()}

	_, err := newTestSync(conn, client, cookies, now, 0).Folders(context.Background())
	suspended, ok := err.(*SuspendedError)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if !suspended.Until.Equal(now.Add(2 * time.Hour)) {
		t.Fatalf("until = %s", suspended.Until)
	}
	var rows int64
	if err := conn.Model(&model.FavoriteList{}).Count(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("wrote %d rows", rows)
	}

	_, err = newTestSync(conn, client, cookies, now.Add(time.Minute), 0).Folders(context.Background())
	if _, ok := err.(*SuspendedError); !ok {
		t.Fatalf("second err = %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestRunRequiresCookie(t *testing.T) {
	conn := openTestDB(t)
	_, err := New(conn, 40)
	if err == nil || err.Error() != "no cookie uploaded" {
		t.Fatal(err)
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "fav.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func stamp(unix int64) dbtime.DateTime {
	return dbtime.DateTime{Time: time.Unix(unix, 0)}
}

func newTestSync(conn *gorm.DB, client *bilibili.Client, cookies []bilibili.Cookie, now time.Time, pageSize int) Sync {
	return Sync{
		DB:           conn,
		API:          client,
		Cookies:      cookies,
		PageSize:     pageSize,
		PageInterval: 0,
		Now:          func() time.Time { return now },
		CoverPush:    &memCoverPush{},
		coverJobs:    new(int),
	}
}
