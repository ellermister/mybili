package syncbili

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func TestCoverFilenameKeepsExtension(t *testing.T) {
	cases := map[string]string{
		"http://i0.hdslb.com/bfs/archive/abc.jpg":                "abc.jpg",
		"http://i0.hdslb.com/bfs/archive/abc.jpg@100w_100h.webp": "abc.jpg100w100h.webp",
		"https://img.example/hello world.png":                    "helloworld.png",
		"":                                                       "",
	}
	for raw, want := range cases {
		if got := coverFilename(raw); got != want {
			t.Fatalf("coverFilename(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestEnqueueCoverUpdatesPendingURL(t *testing.T) {
	conn := openTestDB(t)
	syncer := newTestSync(conn, nil, nil, time.Now(), 0)
	first := "https://img.example/a.png"
	ok, err := syncer.EnqueueCover(nil, CoverKindVideo, 11, first)
	if err != nil || !ok {
		t.Fatalf("first ok=%v err=%v", ok, err)
	}
	ok, err = syncer.EnqueueCover(nil, CoverKindVideo, 11, first)
	if err != nil || ok {
		t.Fatalf("same ok=%v err=%v", ok, err)
	}
	next := "https://img.example/b.png"
	ok, err = syncer.EnqueueCover(nil, CoverKindVideo, 11, next)
	if err != nil || !ok {
		t.Fatalf("changed ok=%v err=%v", ok, err)
	}
	push := coverPushOf(t, syncer)
	if payloadURL(push.get("cover:video:11")) != next || len(push.keys()) != 1 {
		t.Fatalf("push = %+v", push.keys())
	}
	if _, err := syncer.EnqueueCover(nil, CoverKindVideo, 11, "not-a-url"); err != nil {
		t.Fatal(err)
	}
	if payloadURL(push.get("cover:video:11")) != next {
		t.Fatalf("non-http changed payload = %s", push.get("cover:video:11"))
	}
}

func TestEnqueueDoesNotDownloadAndCoverDoes(t *testing.T) {
	conn := openTestDB(t)
	body := tinyPNG(t)
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	syncer := newTestSync(conn, nil, nil, time.Now(), 0)
	syncer.DataDir = t.TempDir()
	rawURL := server.URL + "/bfs/abc.png"
	ok, err := syncer.EnqueueCover(nil, CoverKindVideo, 11, rawURL)
	if err != nil || !ok {
		t.Fatalf("enqueue ok=%v err=%v", ok, err)
	}
	if hits.Load() != 0 || syncer.CoverJobs() != 1 {
		t.Fatalf("hits=%d jobs=%d", hits.Load(), syncer.CoverJobs())
	}
	push := coverPushOf(t, syncer)
	if payloadURL(push.get("cover:video:11")) != rawURL {
		t.Fatalf("payload = %s", push.get("cover:video:11"))
	}

	if err := syncer.Cover(context.Background(), CoverKindVideo, 11, rawURL); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d", hits.Load())
	}

	var cover model.Cover
	if err := conn.Where("filename = ?", "abc.png").First(&cover).Error; err != nil {
		t.Fatal(err)
	}
	if cover.Path != "images/abc.png" || cover.Width != 1 || cover.Height != 1 || cover.MimeType != "image/png" || cover.Type != CoverKindVideo {
		t.Fatalf("cover = %+v", cover)
	}
	var link model.Coverable
	if err := conn.Where("coverable_id = ? AND coverable_type = ?", 11, model.CoverableVideo).First(&link).Error; err != nil {
		t.Fatal(err)
	}
	if link.CoverID != cover.ID {
		t.Fatalf("link = %+v", link)
	}
	file := filepath.Join(syncer.DataDir, "public", "images", "abc.png")
	hash, err := os.ReadFile(file + ".hash")
	if err != nil || len(hash) != 64 {
		t.Fatalf("hash err=%v len=%d", err, len(hash))
	}

	if err := syncer.Cover(context.Background(), CoverKindVideo, 11, rawURL); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits after row exists = %d", hits.Load())
	}
	if err := conn.Delete(&model.Cover{}, cover.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Delete(&model.Coverable{}, link.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := syncer.Cover(context.Background(), CoverKindVideo, 11, rawURL); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits after hash reuse = %d", hits.Load())
	}
}

func TestCoversSkipsLinkedAndInvalid(t *testing.T) {
	conn := openTestDB(t)
	linked := sampleVideo(1, "有封面")
	linked.Cover = "https://img.example/a.png"
	missing := sampleVideo(2, "缺封面")
	missing.Cover = "https://img.example/b.png"
	invalid := sampleVideo(3, "无效")
	invalid.Cover = "https://img.example/c.png"
	invalid.Invalid = 1
	plain := sampleVideo(4, "不是地址")
	plain.Cover = "c"
	for _, row := range []model.Video{linked, missing, invalid, plain} {
		if err := conn.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	created := dbtime.DateTime{Time: time.Now()}
	cover := model.Cover{
		Type: CoverKindVideo, URL: linked.Cover, Filename: "a.png", Path: "images/a.png",
		MimeType: "image/png", CreatedAt: &created,
	}
	if err := conn.Create(&cover).Error; err != nil {
		t.Fatal(err)
	}
	link := model.Coverable{
		CoverID: cover.ID, CoverableID: linked.ID, CoverableType: model.CoverableVideo,
		CreatedAt: &created, UpdatedAt: &created,
	}
	if err := conn.Create(&link).Error; err != nil {
		t.Fatal(err)
	}
	syncer := newTestSync(conn, nil, nil, time.Now(), 0)
	count, err := syncer.Covers(context.Background(), CoverKindVideo, 0)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d", count)
	}
	push := coverPushOf(t, syncer)
	if keys := push.keys(); len(keys) != 1 || keys[0] != "cover:video:2" {
		t.Fatalf("keys = %v", keys)
	}

	extra := sampleVideo(5, "多一条")
	extra.Cover = "https://img.example/d.png"
	if err := conn.Create(&extra).Error; err != nil {
		t.Fatal(err)
	}
	count, err = syncer.Covers(context.Background(), CoverKindVideo, 1)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("limited count = %d", count)
	}
}

func TestEnqueueCoverWhenImageFileMissing(t *testing.T) {
	conn := openTestDB(t)
	rawURL := "https://img.example/a.png"
	created := dbtime.DateTime{Time: time.Now()}
	cover := model.Cover{
		Type: CoverKindVideo, URL: rawURL, Filename: "a.png", Path: "images/a.png",
		MimeType: "image/png", CreatedAt: &created,
	}
	if err := conn.Create(&cover).Error; err != nil {
		t.Fatal(err)
	}
	link := model.Coverable{
		CoverID: cover.ID, CoverableID: 11, CoverableType: model.CoverableVideo,
		CreatedAt: &created, UpdatedAt: &created,
	}
	if err := conn.Create(&link).Error; err != nil {
		t.Fatal(err)
	}
	syncer := newTestSync(conn, nil, nil, time.Now(), 0)
	syncer.DataDir = t.TempDir()
	ok, err := syncer.EnqueueCover(nil, CoverKindVideo, 11, rawURL)
	if err != nil || !ok {
		t.Fatalf("missing file ok=%v err=%v", ok, err)
	}
	if err := os.MkdirAll(syncer.imagesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(syncer.imagesDir()+"/a.png", tinyPNG(t), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err = syncer.EnqueueCover(nil, CoverKindVideo, 11, rawURL)
	if err != nil || ok {
		t.Fatalf("file present ok=%v err=%v", ok, err)
	}
}

func TestCoversScansUppers(t *testing.T) {
	conn := openTestDB(t)
	row := model.Upper{Mid: 9, Name: "UP", Face: "https://img.example/face.png"}
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	syncer := newTestSync(conn, nil, nil, time.Now(), 0)
	count, err := syncer.Covers(context.Background(), "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d", count)
	}
	push := coverPushOf(t, syncer)
	if keys := push.keys(); len(keys) != 1 || keys[0] != coverTaskKey(CoverKindAvatar, row.ID) {
		t.Fatalf("keys = %v", keys)
	}
}

func TestVideoSyncEnqueuesCoverAndAvatar(t *testing.T) {
	conn := openTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"aid":11,"bvid":"BV11","title":"标题","desc":"","pic":"https://img.example/v.png","pubdate":1,"state":0,"owner":{"mid":9,"name":"UP","face":"https://img.example/face.png"},"pages":[{"cid":100,"page":1,"from":"vupload","part":"P","duration":1}]}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, time.Now(), 20)
	if err := syncer.Video(context.Background(), 11, ""); err != nil {
		t.Fatal(err)
	}
	var upper model.Upper
	if err := conn.Where("mid = ?", 9).First(&upper).Error; err != nil {
		t.Fatal(err)
	}
	if upper.ID == 9 {
		t.Fatal("upper id collided with mid")
	}
	keys := coverPushOf(t, syncer).keys()
	if len(keys) != 2 || keys[0] != coverTaskKey(CoverKindAvatar, upper.ID) || keys[1] != "cover:video:11" {
		t.Fatalf("keys = %v upper=%d", keys, upper.ID)
	}
}

func TestInvalidVideoDoesNotEnqueueCover(t *testing.T) {
	conn := openTestDB(t)
	existing := sampleVideo(11, "旧标题")
	if err := conn.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"aid":11,"bvid":"BV11","title":"失效了","desc":"","pic":"https://img.example/v.png","pubdate":1,"state":-1,"owner":{"mid":3,"name":"已注销","face":"https://img.example/face.png"},"pages":[{"cid":5,"page":1,"part":"不该写入","duration":1}]}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, time.Now(), 20)
	if err := syncer.Video(context.Background(), 11, ""); err != nil {
		t.Fatal(err)
	}
	if keys := coverPushOf(t, syncer).keys(); len(keys) != 0 {
		t.Fatalf("keys = %v", keys)
	}
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
