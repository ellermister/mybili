package syncbili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mybili/internal/bilibili"
	"mybili/internal/model"
)

func TestParseSubscriptionURL(t *testing.T) {
	mid, listID, err := parseSubscriptionURL("seasons", "https://space.bilibili.com/123/lists/456?type=season")
	if err != nil || mid != "123" || listID != "456" {
		t.Fatalf("seasons = %s %s %v", mid, listID, err)
	}
	mid, listID, err = parseSubscriptionURL("series", "https://space.bilibili.com/9/lists/8")
	if err != nil || mid != "9" || listID != "8" {
		t.Fatalf("series = %s %s %v", mid, listID, err)
	}
	mid, listID, err = parseSubscriptionURL("up", "https://space.bilibili.com/42/upload/video")
	if err != nil || mid != "42" || listID != "" {
		t.Fatalf("upload = %s %s %v", mid, listID, err)
	}
	mid, listID, err = parseSubscriptionURL("up", "https://space.bilibili.com/77")
	if err != nil || mid != "77" || listID != "" {
		t.Fatalf("space = %s %s %v", mid, listID, err)
	}
	if _, _, err := parseSubscriptionURL("up", "https://www.bilibili.com/video/BV1"); err == nil {
		t.Fatal("expected invalid up url")
	}
	if _, _, err := parseSubscriptionURL("audio", "https://space.bilibili.com/1/lists/2"); err == nil {
		t.Fatal("expected invalid type")
	}
}

func TestCreateSubscriptionKeepsExistingSeasons(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, time.Local)
	first, err := CreateSubscription(context.Background(), conn, "seasons", "https://space.bilibili.com/123/lists/456", now)
	if err != nil {
		t.Fatal(err)
	}
	if first.Mid != "123" || first.ListID == nil || *first.ListID != "456" || first.Status != 1 {
		t.Fatalf("first = %+v", first)
	}
	second, err := CreateSubscription(context.Background(), conn, "seasons", "https://space.bilibili.com/123/lists/456?type=season", now)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("duplicate id = %d, want %d", second.ID, first.ID)
	}
	var count int64
	if err := conn.Model(&model.Subscription{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rows = %d", count)
	}
}

func TestDeleteSubscriptionKeepsFavoritedVideo(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, time.Local)
	stamp := stamp(now.Unix())
	videos := []model.Video{
		{ID: 11, Link: "a", Title: "收藏里", Intro: "", Cover: "", Bvid: "BV11", Pubtime: stamp, CacheImage: ""},
		{ID: 12, Link: "b", Title: "只在订阅", Intro: "", Cover: "", Bvid: "BV12", Pubtime: stamp, CacheImage: ""},
		{ID: 13, Link: "c", Title: "别的订阅", Intro: "", Cover: "", Bvid: "BV13", Pubtime: stamp, CacheImage: ""},
	}
	if err := conn.Create(&videos).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: 1, VideoID: 11}).Error; err != nil {
		t.Fatal(err)
	}
	listA, listB := "1", "2"
	subs := []model.Subscription{
		{ID: 1, Type: "seasons", Mid: "9", ListID: &listA, Name: "甲", URL: "u", Total: "2", Status: 1},
		{ID: 2, Type: "seasons", Mid: "9", ListID: &listB, Name: "乙", URL: "u", Total: "1", Status: 1},
	}
	if err := conn.Create(&subs).Error; err != nil {
		t.Fatal(err)
	}
	links := []model.SubscriptionVideo{
		{SubscriptionID: 1, VideoID: 11, Bvid: "BV11"},
		{SubscriptionID: 1, VideoID: 12, Bvid: "BV12"},
		{SubscriptionID: 1, VideoID: 13, Bvid: "BV13"},
		{SubscriptionID: 2, VideoID: 13, Bvid: "BV13"},
	}
	if err := conn.Create(&links).Error; err != nil {
		t.Fatal(err)
	}
	ids, err := DeleteSubscription(context.Background(), conn, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 {
		t.Fatalf("ids = %v", ids)
	}
	var gone model.Subscription
	if err := conn.First(&gone, 1).Error; err == nil {
		t.Fatal("subscription still exists")
	}
	var kept, removed, shared model.Video
	if err := conn.First(&kept, 11).Error; err != nil || kept.DeletedAt != nil {
		t.Fatalf("favorited = %+v %v", kept, err)
	}
	if err := conn.First(&removed, 12).Error; err != nil || removed.DeletedAt == nil {
		t.Fatalf("only-sub = %+v %v", removed, err)
	}
	if err := conn.First(&shared, 13).Error; err != nil || shared.DeletedAt != nil {
		t.Fatalf("shared = %+v %v", shared, err)
	}
	var left int64
	if err := conn.Model(&model.SubscriptionVideo{}).Where("subscription_id = ?", 1).Count(&left).Error; err != nil {
		t.Fatal(err)
	}
	if left != 0 {
		t.Fatalf("links left = %d", left)
	}
}

func TestUpdateSeasonSubscriptionOnePage(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, time.Local)
	var pages int
	var queued []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages++
		_, _ = w.Write([]byte(`{"code":0,"data":{"meta":{"name":"合集甲","description":"简介","cover":"https://example.com/c.jpg","total":40},"archives":[{"aid":101,"bvid":"BV101","title":"第一期","pic":"https://example.com/v.jpg","duration":12,"pubdate":1700000000}]}}`))
	}))
	defer server.Close()
	row, err := CreateSubscription(context.Background(), conn, "seasons", "https://space.bilibili.com/123/lists/456", now)
	if err != nil {
		t.Fatal(err)
	}
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, now, 0)
	syncer.QueueVideo = func(_ context.Context, bvid string) error {
		queued = append(queued, bvid)
		return nil
	}
	if err := syncer.UpdateSubscription(context.Background(), row.ID, false); err != nil {
		t.Fatal(err)
	}
	if pages != 1 {
		t.Fatalf("pages = %d", pages)
	}
	if len(queued) != 1 || queued[0] != "BV101" {
		t.Fatalf("queued = %v", queued)
	}
	var saved model.Subscription
	if err := conn.First(&saved, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Name != "合集甲" || saved.Total != "40" || saved.Cover != "https://example.com/c.jpg" {
		t.Fatalf("saved = %+v", saved)
	}
	var link model.SubscriptionVideo
	if err := conn.Where("subscription_id = ? AND video_id = ?", row.ID, 101).First(&link).Error; err != nil {
		t.Fatal(err)
	}
	if link.Bvid != "BV101" {
		t.Fatalf("link = %+v", link)
	}
	var videos int64
	if err := conn.Model(&model.Video{}).Count(&videos).Error; err != nil {
		t.Fatal(err)
	}
	if videos != 0 {
		t.Fatalf("season sync should leave the video row to the detail task, got %d", videos)
	}
	if err := syncer.UpdateSubscription(context.Background(), row.ID, false); err != nil {
		t.Fatal(err)
	}
	if len(queued) != 2 {
		t.Fatalf("missing video should be queued again, got %v", queued)
	}
}

func TestSubscriptionEnqueuesDownloadForExistingParts(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, time.Local)
	stamp := stamp(now.Unix())
	on := `"on"`
	if err := conn.Create(&model.Setting{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		ID: 101, Link: "https://www.bilibili.com/video/BV101", Title: "第一期", Intro: "", Cover: "",
		Bvid: "BV101", Pubtime: stamp, CacheImage: "", Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: 101, Cid: 900, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	views := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/x/web-interface/view") {
			views++
			_, _ = w.Write([]byte(`{"code":0,"data":{"aid":101,"bvid":"BV101","title":"第一期","desc":"","pic":"https://example.com/v.jpg","pubdate":1700000000,"state":0,"owner":{"mid":9,"name":"up","face":""},"pages":[{"cid":900,"page":1,"from":"vupload","part":"P1","duration":12,"vid":"","weblink":"","dimension":{"width":0,"height":0,"rotate":0},"first_frame":""}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"meta":{"name":"合集甲","description":"简介","cover":"https://example.com/c.jpg","total":1},"archives":[{"aid":101,"bvid":"BV101","title":"第一期","pic":"https://example.com/v.jpg","duration":12,"pubdate":1700000000}]}}`))
	}))
	defer server.Close()
	row, err := CreateSubscription(context.Background(), conn, "seasons", "https://space.bilibili.com/123/lists/456", now)
	if err != nil {
		t.Fatal(err)
	}
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, now, 0)
	syncer.DataDir = t.TempDir()
	if err := syncer.UpdateSubscription(context.Background(), row.ID, false); err != nil {
		t.Fatal(err)
	}
	var tasks int64
	if err := conn.Model(&model.Job{}).Where("type = ? AND status = ?", "download-video", "pending").Count(&tasks).Error; err != nil {
		t.Fatal(err)
	}
	if tasks != 1 {
		t.Fatalf("download tasks = %d", tasks)
	}
	if views != 1 {
		t.Fatalf("views = %d", views)
	}
	if err := syncer.UpdateSubscription(context.Background(), row.ID, false); err != nil {
		t.Fatal(err)
	}
	if err := conn.Model(&model.Job{}).Where("type = ?", "download-video").Count(&tasks).Error; err != nil {
		t.Fatal(err)
	}
	if tasks != 1 {
		t.Fatalf("download tasks after second sync = %d", tasks)
	}
	if views != 1 {
		t.Fatalf("views after second sync = %d", views)
	}
}

func TestSubscriptionMarksMissingVideoInvalid(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, time.Local)
	stamp := stamp(now.Unix())
	on := `"on"`
	if err := conn.Create(&model.Setting{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		ID: 101, Link: "https://www.bilibili.com/video/BV101", Title: "第一期", Intro: "", Cover: "",
		Bvid: "BV101", Pubtime: stamp, CacheImage: "", Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: 101, Cid: 900, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	views := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/x/web-interface/view") {
			views++
			_, _ = w.Write([]byte(`{"code":-404,"message":"啥都木有"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"meta":{"name":"合集甲","description":"简介","cover":"https://example.com/c.jpg","total":1},"archives":[{"aid":101,"bvid":"BV101","title":"第一期","pic":"https://example.com/v.jpg","duration":12,"pubdate":1700000000}]}}`))
	}))
	defer server.Close()
	row, err := CreateSubscription(context.Background(), conn, "seasons", "https://space.bilibili.com/123/lists/456", now)
	if err != nil {
		t.Fatal(err)
	}
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, now, 0)
	syncer.DataDir = t.TempDir()
	if err := syncer.UpdateSubscription(context.Background(), row.ID, false); err != nil {
		t.Fatal(err)
	}
	var saved model.Video
	if err := conn.First(&saved, 101).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Invalid != 1 {
		t.Fatalf("invalid = %d", saved.Invalid)
	}
	var tasks int64
	if err := conn.Model(&model.Job{}).Where("type = ?", "download-video").Count(&tasks).Error; err != nil {
		t.Fatal(err)
	}
	if tasks != 0 || views != 1 {
		t.Fatalf("tasks = %d views = %d", tasks, views)
	}
	if err := syncer.UpdateSubscription(context.Background(), row.ID, false); err != nil {
		t.Fatal(err)
	}
	if views != 1 {
		t.Fatalf("views after second sync = %d", views)
	}
}

func TestSubscriptionSkipsViewWhenFileExists(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, time.Local)
	stamp := stamp(now.Unix())
	on := `"on"`
	if err := conn.Create(&model.Setting{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		ID: 101, Link: "https://www.bilibili.com/video/BV101", Title: "第一期", Intro: "", Cover: "",
		Bvid: "BV101", Pubtime: stamp, CacheImage: "", Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: 101, Cid: 900, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "public", "videos", "101.900.mp4")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	views := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/x/web-interface/view") {
			views++
			t.Errorf("file exists, view was called")
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"meta":{"name":"合集甲","description":"简介","cover":"https://example.com/c.jpg","total":1},"archives":[{"aid":101,"bvid":"BV101","title":"第一期","pic":"https://example.com/v.jpg","duration":12,"pubdate":1700000000}]}}`))
	}))
	defer server.Close()
	row, err := CreateSubscription(context.Background(), conn, "seasons", "https://space.bilibili.com/123/lists/456", now)
	if err != nil {
		t.Fatal(err)
	}
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, now, 0)
	syncer.DataDir = dir
	if err := syncer.UpdateSubscription(context.Background(), row.ID, false); err != nil {
		t.Fatal(err)
	}
	var tasks int64
	if err := conn.Model(&model.Job{}).Where("type = ?", "download-video").Count(&tasks).Error; err != nil {
		t.Fatal(err)
	}
	if tasks != 0 || views != 0 {
		t.Fatalf("tasks = %d views = %d", tasks, views)
	}
}
