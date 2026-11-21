package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"mybili/internal/config"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func TestVideoDestroy(t *testing.T) {
	conn := openQueueDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	video := model.Video{
		Link: "l", Title: "要删", Intro: "", Cover: "c", Bvid: "BV9",
		Pubtime: stamp, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	app := &App{DB: conn, Config: config.Config{DataDir: t.TempDir()}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/videos/"+strconv.FormatInt(video.ID, 10), strings.NewReader(`{"permanent":true}`))
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var saved model.Video
	if err := conn.First(&saved, video.ID).Error; err != nil || saved.DeletedAt == nil {
		t.Fatal(err, saved.DeletedAt)
	}

	app.Config.SettingReadOnly = true
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/videos/"+strconv.FormatInt(video.ID, 10), strings.NewReader(`{}`))
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("readonly status=%d", rec.Code)
	}
}
