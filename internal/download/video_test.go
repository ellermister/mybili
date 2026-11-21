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

func TestVideoDownloadsPlayURLAndMerges(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	stamp := dbtime.DateTime{Time: dbtime.Now()}
	value := `".bilibili.com\tTRUE\t/\tFALSE\t0\tSESSDATA\tabc"`
	if err := conn.Create(&model.Setting{Name: "cookies_content", Value: &value, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "l", Title: "新视频", Intro: "", Cover: "c", Bvid: "BV1",
		Pubtime: stamp, Type: 2, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: video.ID, Cid: 8, Page: 1, From: "v", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", FirstFrame: "",
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	var playHits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/x/player/playurl":
			playHits++
			if r.URL.Query().Get("cid") != "8" || r.URL.Query().Get("avid") == "" || r.URL.Query().Get("fnval") != "80" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			if r.Header.Get("Cookie") != "SESSDATA=abc" {
				t.Errorf("playurl cookie = %s", r.Header.Get("Cookie"))
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"dash":{"video":[{"id":80,"baseUrl":"` + serverURL(r, "/v.m4s") + `","bandwidth":20,"codecid":12,"codecs":"hvc1"}],"audio":[{"id":30280,"baseUrl":"` + serverURL(r, "/a.m4s") + `","bandwidth":10}]}}}`))
		case "/v.m4s", "/a.m4s":
			if r.Header.Get("Referer") != fetchFrom || r.Header.Get("User-Agent") != fetchUA || r.Header.Get("Cookie") != "SESSDATA=abc" {
				t.Errorf("headers referer=%s ua=%s cookie=%s", r.Header.Get("Referer"), r.Header.Get("User-Agent"), r.Header.Get("Cookie"))
			}
			if r.Method == http.MethodHead {
				w.Header().Set("Content-Length", "4")
				return
			}
			_, _ = w.Write([]byte("m4s!"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	bin := filepath.Join(dir, "fake-ffmpeg")
	script := "#!/bin/sh\nout=\nfor arg in \"$@\"; do out=$arg; done\nprintf merged > \"$out\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	runner := Runner{DB: conn, DataDir: dir, PlayBase: server.URL, FFmpeg: bin}
	ready, err := runner.Video(context.Background(), 1, partPayload{PartID: part.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !ready.Fresh || ready.PartID != part.ID || playHits != 1 {
		t.Fatalf("ready = %+v hits=%d", ready, playHits)
	}
	body, err := os.ReadFile(videoFilePath(dir, video.ID, part.Cid))
	if err != nil || string(body) != "merged" {
		t.Fatalf("file = %q %v", body, err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "public", "tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("tmp left %d dirs", len(entries))
	}
	ready, err = runner.Video(context.Background(), 1, partPayload{PartID: part.ID})
	if err != nil {
		t.Fatal(err)
	}
	if ready.Fresh || playHits != 1 {
		t.Fatalf("existing = %+v hits=%d", ready, playHits)
	}
}

func serverURL(r *http.Request, path string) string {
	return "http://" + r.Host + path
}
