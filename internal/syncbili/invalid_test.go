package syncbili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mybili/internal/bilibili"
	"mybili/internal/model"
)

func TestFixInvalidFavoriteUpdatesRealMetadata(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 4, 0, 0, 0, time.Local)
	stamp := stamp(now.Unix())
	videos := []model.Video{
		{ID: 11, Link: "a", Title: "已失效视频", Intro: "", Cover: "old", Bvid: "BV11", Pubtime: stamp, CacheImage: ""},
		{ID: 12, Link: "b", Title: "占位", Intro: "", Cover: "old", Bvid: "BV12", Pubtime: stamp, CacheImage: ""},
		{ID: 13, Link: "c", Title: "有效", Intro: "", Cover: "keep", Bvid: "BV13", Pubtime: stamp, CacheImage: "", Invalid: 0},
	}
	if err := conn.Create(&videos).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/v3/fav/folder/resources" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"list":[
			{"oid":11,"title":"真实标题","cover":"https://example.com/real.jpg","is_invalid":true},
			{"oid":12,"title":"已失效视频","cover":"https://i0.hdslb.com/bfs/archive/be27fd62c99036dce67efface486fb0a88ffed06.jpg","is_invalid":true},
			{"oid":13,"title":"有效","cover":"https://example.com/ok.jpg","is_invalid":false},
			{"oid":99,"title":"本地没有","cover":"https://example.com/missing.jpg","is_invalid":true}
		]}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, now, 40)
	fixed, err := syncer.FixInvalidFavorite(context.Background(), 7, 1)
	if err != nil {
		t.Fatal(err)
	}
	if fixed != 1 {
		t.Fatalf("fixed = %d", fixed)
	}
	var updated, polluted, valid, missing model.Video
	if err := conn.First(&updated, 11).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Invalid != 1 || updated.Title != "真实标题" || updated.Cover != "https://example.com/real.jpg" {
		t.Fatalf("updated = %+v", updated)
	}
	if err := conn.First(&polluted, 12).Error; err != nil {
		t.Fatal(err)
	}
	if polluted.Title != "占位" || polluted.Cover != "old" || polluted.Invalid != 0 {
		t.Fatalf("polluted = %+v", polluted)
	}
	if err := conn.First(&valid, 13).Error; err != nil {
		t.Fatal(err)
	}
	if valid.Title != "有效" || valid.Cover != "keep" || valid.Invalid != 0 {
		t.Fatalf("valid = %+v", valid)
	}
	if err := conn.First(&missing, 99).Error; err == nil {
		t.Fatal("missing video was created")
	}
	push := coverPushOf(t, syncer)
	if len(push.keys()) != 1 {
		t.Fatalf("cover jobs = %v", push.keys())
	}
}

func TestFixInvalidFavoriteRedownloadsMissingCacheImage(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 4, 0, 0, 0, time.Local)
	stamp := stamp(now.Unix())
	cover := "https://example.com/real.jpg"
	video := model.Video{
		ID: 11, Link: "a", Title: "旧标题", Intro: "", Cover: cover, Bvid: "BV11",
		Pubtime: stamp, CacheImage: "images/real.jpg",
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"list":[{"oid":11,"title":"真实标题","cover":"` + cover + `","is_invalid":true}]}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	dir := t.TempDir()
	syncer := newTestSync(conn, client, nil, now, 40)
	syncer.DataDir = dir
	if _, err := syncer.FixInvalidFavorite(context.Background(), 7, 1); err != nil {
		t.Fatal(err)
	}
	if keys := coverPushOf(t, syncer).keys(); len(keys) != 1 {
		t.Fatalf("missing file should enqueue, got %v", keys)
	}

	if err := os.MkdirAll(filepath.Join(dir, "public", "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "public", "images", "real.jpg"), []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	syncer.CoverPush = &memCoverPush{}
	if _, err := syncer.FixInvalidFavorite(context.Background(), 7, 1); err != nil {
		t.Fatal(err)
	}
	if keys := coverPushOf(t, syncer).keys(); len(keys) != 0 {
		t.Fatalf("existing file should stay, got %v", keys)
	}
}

func TestFixInvalidFavoriteRedownloadsWhenAPIOmitsVideo(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 4, 0, 0, 0, time.Local)
	stamp := stamp(now.Unix())
	videos := []model.Video{
		{
			ID: 11, Link: "a", Title: "真实标题", Intro: "", Cover: "https://example.com/real.jpg", Bvid: "BV11",
			Pubtime: stamp, CacheImage: "images/real.jpg", Invalid: 1,
		},
		{
			ID: 12, Link: "b", Title: "已失效视频", Intro: "", Cover: "https://i0.hdslb.com/bfs/archive/be27fd62c99036dce67efface486fb0a88ffed06.jpg", Bvid: "BV12",
			Pubtime: stamp, CacheImage: "images/bad.jpg", Invalid: 1,
		},
		{
			ID: 13, Link: "c", Title: "别的收藏夹", Intro: "", Cover: "https://example.com/other.jpg", Bvid: "BV13",
			Pubtime: stamp, CacheImage: "images/other.jpg", Invalid: 1,
		},
	}
	if err := conn.Create(&videos).Error; err != nil {
		t.Fatal(err)
	}
	links := []model.FavoriteListVideo{
		{FavoriteListID: 7, VideoID: 11},
		{FavoriteListID: 7, VideoID: 12},
		{FavoriteListID: 8, VideoID: 13},
	}
	if err := conn.Create(&links).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"list":[]}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	dir := t.TempDir()
	syncer := newTestSync(conn, client, nil, now, 40)
	syncer.DataDir = dir
	fixed, err := syncer.FixInvalidFavorite(context.Background(), 7, 1)
	if err != nil {
		t.Fatal(err)
	}
	if fixed != 0 {
		t.Fatalf("fixed = %d", fixed)
	}
	if keys := coverPushOf(t, syncer).keys(); len(keys) != 1 || keys[0] != coverTaskKey(CoverKindVideo, 11) {
		t.Fatalf("cover jobs = %v", keys)
	}
}

func TestFixInvalidSpecsPages(t *testing.T) {
	conn := openTestDB(t)
	rows := []model.FavoriteList{
		{ID: 1, Title: "甲", Cover: "", MediaCount: 45, Ctime: stamp(1), Mtime: stamp(1)},
		{ID: 2, Title: "空", Cover: "", MediaCount: 0, Ctime: stamp(1), Mtime: stamp(1)},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 31, 4, 0, 0, 0, time.Local)
	specs := FixInvalidSpecs(conn, 40)
	if !specs[0].Due(now) || specs[0].Due(now.Add(time.Minute)) {
		t.Fatal("due window")
	}
	drafts, err := specs[0].Tasks(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(drafts) != 2 {
		t.Fatalf("drafts = %d", len(drafts))
	}
	if drafts[0].UniqueKey != "sync:fix-invalid:1:1" || drafts[1].UniqueKey != "sync:fix-invalid:1:2" {
		t.Fatalf("keys = %s %s", drafts[0].UniqueKey, drafts[1].UniqueKey)
	}
}
