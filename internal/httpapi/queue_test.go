package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/migrate"
	"mybili/internal/model"
)

func TestTaskQueueListsAndRetries(t *testing.T) {
	conn := openQueueDB(t)
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	rows := []model.Job{
		{Queue: "bili-fast", Type: "fav-list", UniqueKey: "sync:fav-list", Status: jobs.StatusPending, Payload: "{}", MaxRetries: 3, CreatedAt: &stamp},
		{Queue: "bili-fast", Type: "fav-videos", UniqueKey: "sync:fav-videos:1", Status: jobs.StatusQueued, Payload: `{"page":1}`, MaxRetries: 3, CreatedAt: &stamp},
		{Queue: "bili-fast", Type: "fav-videos", UniqueKey: "sync:fav-videos:0", Status: jobs.StatusFailed, Payload: `{"page":0}`, MaxRetries: 3, ErrorMsg: strPtr("timeout"), CreatedAt: &stamp},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	waker := &recordWake{}
	app := &App{DB: conn, Tasks: waker}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/download-queue?status=pending,queued&page=1&per_page=30", nil)
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		List []struct {
			ID      int64  `json:"id"`
			Type    string `json:"type"`
			Status  string `json:"status"`
			Subject *struct {
				Page int `json:"page"`
			} `json:"subject"`
		} `json:"list"`
		Total int64            `json:"total"`
		Stat  map[string]int64 `json:"stat"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 2 || len(body.List) != 2 || body.Stat["pending"] != 1 || body.Stat["queued"] != 1 || body.Stat["failed"] != 1 {
		t.Fatalf("body = %+v", body)
	}
	if body.List[0].Status != jobs.StatusQueued || body.List[0].Type != "fav-videos" || body.List[0].Subject == nil || body.List[0].Subject.Page != 1 {
		t.Fatalf("first = %+v", body.List[0])
	}
	if strings.Contains(rec.Body.String(), "同步收藏夹") {
		t.Fatal("task text should stay out of the api")
	}

	failedID := rows[2].ID
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/download-queue/"+itoa64(failedID)+"/retry", nil)
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !decodeSuccess(t, rec) {
		t.Fatalf("retry = %d %s", rec.Code, rec.Body.String())
	}
	if len(waker.queues) != 1 || waker.queues[0] != "bili-fast" {
		t.Fatalf("wakes = %v", waker.queues)
	}
	saved := reloadTask(t, conn, failedID)
	if saved.Status != jobs.StatusPending || saved.RetryCount != 0 {
		t.Fatalf("retried = %+v", saved)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/download-queue/"+itoa64(rows[0].ID)+"/cancel", nil)
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !decodeSuccess(t, rec) {
		t.Fatalf("cancel = %d %s", rec.Code, rec.Body.String())
	}
	if reloadTask(t, conn, rows[0].ID).Status != jobs.StatusCancelled {
		t.Fatal("pending was not cancelled")
	}
}

func TestTaskQueueReadableLinkUsesVideoCover(t *testing.T) {
	conn := openQueueDB(t)
	stamp := dbtime.DateTime{Time: time.Date(2026, 10, 30, 10, 0, 0, 0, dbtime.Location())}
	video := model.Video{
		ID: 101, Link: "https://www.bilibili.com/video/BV101", Title: "第一期", Intro: "",
		Cover: "https://example.com/remote.jpg", Bvid: "BV101", Pubtime: stamp, CacheImage: "", Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: 101, Cid: 900, Page: 2, From: "vupload", Part: "正片", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	cover := model.Cover{Type: "video", URL: "u", Filename: "cover.jpg", Path: "images/cover.jpg", MimeType: "image/jpeg"}
	if err := conn.Create(&cover).Error; err != nil {
		t.Fatal(err)
	}
	link := model.Coverable{CoverID: cover.ID, CoverableID: video.ID, CoverableType: model.CoverableVideo}
	if err := conn.Create(&link).Error; err != nil {
		t.Fatal(err)
	}
	job := model.Job{
		Queue: "local", Type: "readable-link", UniqueKey: fmt.Sprintf("local:readable-link:%d", part.ID),
		Status: jobs.StatusPending, Payload: fmt.Sprintf(`{"part_id":%d}`, part.ID), MaxRetries: 3, CreatedAt: &stamp,
	}
	if err := conn.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	app := &App{DB: conn}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/download-queue?status=pending", nil)
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		List []struct {
			Type     string `json:"type"`
			ImageURL string `json:"image_url"`
			Href     string `json:"href"`
			Subject  struct {
				Title string `json:"title"`
				Bvid  string `json:"bvid"`
				Page  int    `json:"page"`
				Part  string `json:"part"`
			} `json:"subject"`
		} `json:"list"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.List) != 1 {
		t.Fatalf("list = %+v", body.List)
	}
	item := body.List[0]
	if item.Type != "readable-link" || item.ImageURL != "/storage/images/cover.jpg" || item.Href != "/video/101" {
		t.Fatalf("item = %+v", item)
	}
	if item.Subject.Title != "第一期" || item.Subject.Bvid != "BV101" || item.Subject.Page != 2 || item.Subject.Part != "正片" {
		t.Fatalf("subject = %+v", item.Subject)
	}
}

func TestTaskRetrySkipsWhenActiveExists(t *testing.T) {
	conn := openQueueDB(t)
	rows := []model.Job{
		{Queue: "bili-fast", Type: "fav-list", UniqueKey: "sync:fav-list", Status: jobs.StatusFailed, Payload: "{}", MaxRetries: 3},
		{Queue: "bili-fast", Type: "fav-list", UniqueKey: "sync:fav-list", Status: jobs.StatusRunning, Payload: "{}", MaxRetries: 3},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	app := &App{DB: conn, Tasks: &recordWake{}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/download-queue/"+itoa64(rows[0].ID)+"/retry", nil)
	app.Handler("").ServeHTTP(rec, req)
	if decodeSuccess(t, rec) {
		t.Fatal("retry succeeded while another row is running")
	}
	if reloadTask(t, conn, rows[0].ID).Status != jobs.StatusFailed {
		t.Fatal("failed row was reopened")
	}
}

type recordWake struct {
	queues []string
}

func (r *recordWake) Wake(_ context.Context, queue string) error {
	r.queues = append(r.queues, queue)
	return nil
}

func openQueueDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "queue.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func reloadTask(t *testing.T, conn *gorm.DB, id int64) model.Job {
	t.Helper()
	var job model.Job
	if err := conn.First(&job, id).Error; err != nil {
		t.Fatal(err)
	}
	return job
}

func decodeSuccess(t *testing.T, rec *httptest.ResponseRecorder) bool {
	t.Helper()
	var body struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Success
}

func itoa64(n int64) string {
	return strconvItoa(n)
}

func strconvItoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[i:])
}

func strPtr(s string) *string { return &s }
