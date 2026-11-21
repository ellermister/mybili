package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func TestAudioDownloadsM4A(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	stamp := dbtime.DateTime{Time: dbtime.Now()}
	value := `".bilibili.com\tTRUE\t/\tFALSE\t0\tSESSDATA\tabc"`
	if err := conn.Create(&model.Setting{Name: "cookies_content", Value: &value, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "l", Title: "音频", Intro: "", Cover: "c", Bvid: "",
		Pubtime: stamp, Type: 12, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.AudioPart{VideoID: video.ID, Sid: video.ID, Duration: 30, CreatedAt: &stamp}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "SESSDATA=abc" {
			t.Errorf("cookie = %s", r.Header.Get("Cookie"))
		}
		switch r.URL.Path {
		case "/audio/music-service-c/web/url":
			hits++
			if r.URL.Query().Get("sid") == "" || r.URL.Query().Get("privilege") != "2" || r.URL.Query().Get("quality") != "2" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"sid":1,"size":4,"cdns":["` + serverURL(r, "/a.m4a") + `"]}}`))
		case "/a.m4a":
			if r.Header.Get("Referer") != fetchFrom || r.Header.Get("User-Agent") != fetchUA {
				t.Errorf("headers = %s %s", r.Header.Get("Referer"), r.Header.Get("User-Agent"))
			}
			_, _ = w.Write([]byte("m4a!"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	runner := Runner{DB: conn, DataDir: dir, PlayBase: server.URL}
	if err := runner.Audio(context.Background(), 1, audioPayload{PartID: part.ID, Sid: part.Sid}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(audioFilePath(dir, part.Sid))
	if err != nil || string(body) != "m4a!" {
		t.Fatalf("file = %q %v", body, err)
	}
	var saved model.AudioPart
	if err := conn.First(&saved, part.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.AudioDownloadPath == nil || *saved.AudioDownloadPath != storageAudioPath(part.Sid) {
		t.Fatalf("path = %v", saved.AudioDownloadPath)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "public", "tmp"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("tmp = %v %v", entries, err)
	}
	if err := runner.Audio(context.Background(), 1, audioPayload{PartID: part.ID}); err != nil {
		t.Fatal(err)
	}
	if hits != 1 {
		t.Fatalf("hits = %d", hits)
	}
}

func TestAudioKeepsExistingMP3(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	stamp := dbtime.DateTime{Time: dbtime.Now()}
	value := `".bilibili.com\tTRUE\t/\tFALSE\t0\tSESSDATA\tabc"`
	if err := conn.Create(&model.Setting{Name: "cookies_content", Value: &value, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "l", Title: "音频", Intro: "", Cover: "c", Bvid: "",
		Pubtime: stamp, Type: 12, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.AudioPart{VideoID: video.ID, Sid: 9, Duration: 3, CreatedAt: &stamp}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	old := legacyAudioFilePath(dir, part.Sid)
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("mp3"), 0o644); err != nil {
		t.Fatal(err)
	}
	runner := Runner{DB: conn, DataDir: dir, PlayBase: "http://127.0.0.1:1"}
	if err := runner.Audio(context.Background(), 1, audioPayload{PartID: part.ID, Sid: part.Sid}); err != nil {
		t.Fatal(err)
	}
	var saved model.AudioPart
	if err := conn.First(&saved, part.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.AudioDownloadPath == nil || *saved.AudioDownloadPath != legacyStorageAudioPath(part.Sid) {
		t.Fatalf("path = %v", saved.AudioDownloadPath)
	}
}
