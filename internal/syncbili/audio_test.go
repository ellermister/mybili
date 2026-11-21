package syncbili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mybili/internal/bilibili"
	"mybili/internal/model"
)

func TestSyncAudioWritesPartAndQueuesDownload(t *testing.T) {
	conn := openTestDB(t)
	video := sampleVideo(42, "音频")
	video.Type = 12
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	value := `"on"`
	if err := conn.Create(&model.Setting{Name: "video_download_enabled", Value: &value}).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("sid") != "42" {
			t.Errorf("sid = %s", r.URL.Query().Get("sid"))
		}
		_, _ = w.Write([]byte(`{"code":0,"message":"0","data":{"duration":95}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, time.Now(), 20)
	syncer.DataDir = t.TempDir()
	if err := syncer.SyncAudio(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	var part model.AudioPart
	if err := conn.Where("video_id = ?", 42).First(&part).Error; err != nil {
		t.Fatal(err)
	}
	if part.Sid != 42 || part.Duration != 95 {
		t.Fatalf("part = %+v", part)
	}
	var job model.Job
	if err := conn.Where("unique_key = ?", "download:audio:42").First(&job).Error; err != nil {
		t.Fatal(err)
	}
	if job.Type != "download-audio" || job.Status != "pending" {
		t.Fatalf("job = %+v", job)
	}
	if err := syncer.SyncAudio(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := conn.Model(&model.AudioPart{}).Where("video_id = ?", 42).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("parts = %d", count)
	}
}
