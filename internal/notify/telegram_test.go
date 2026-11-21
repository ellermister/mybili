package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/migrate"
	"mybili/internal/model"
	"mybili/internal/taskrun"
)

type memQueue struct {
	once bool
	msgs []taskrun.Message
}

func (m *memQueue) Push(_ context.Context, _ string, msg taskrun.Message) error {
	m.msgs = append(m.msgs, msg)
	return nil
}

func (m *memQueue) MarkOnce(context.Context, string, time.Duration) (bool, error) {
	if m.once {
		return false, nil
	}
	m.once = true
	return true, nil
}

func (m *memQueue) Unmark(context.Context, string) error {
	m.once = false
	return nil
}

func TestEnqueueCookieOnce(t *testing.T) {
	q := &memQueue{}
	if err := EnqueueCookie(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if err := EnqueueCookie(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if len(q.msgs) != 1 || !strings.Contains(q.msgs[0].Payload, cookieKind) {
		t.Fatalf("msgs = %+v", q.msgs)
	}
	if err := ClearCookieNotice(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if err := EnqueueCookie(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if len(q.msgs) != 2 {
		t.Fatalf("msgs = %d", len(q.msgs))
	}
}

func TestTelegramConnection(t *testing.T) {
	var send string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/getMe") {
			_, _ = w.Write([]byte(`{"ok":true}`))
			return
		}
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		send = string(body)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	ok, err := Test(context.Background(), "token", "chat", server.URL)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if !strings.Contains(send, "connected successfully") {
		t.Fatalf("send = %s", send)
	}
	ok, err = Test(context.Background(), "", "chat", server.URL)
	if err != nil || ok {
		t.Fatalf("empty token ok=%v err=%v", ok, err)
	}
}

func TestEnqueueDownloaded(t *testing.T) {
	q := &memQueue{}
	if err := EnqueueDownloaded(context.Background(), q, 9); err != nil {
		t.Fatal(err)
	}
	if err := EnqueueDownloaded(context.Background(), q, 9); err != nil {
		t.Fatal(err)
	}
	if len(q.msgs) != 2 || !strings.Contains(q.msgs[0].Payload, downloadKind) {
		t.Fatalf("msgs = %+v", q.msgs)
	}
}

func TestDownloadedMessage(t *testing.T) {
	conn := openNotifyDB(t)
	stamp := dbtime.DateTime{Time: time.Date(2026, 10, 31, 12, 0, 0, 0, dbtime.Location())}
	video := model.Video{
		Link: "l", Title: "标题<A>", Intro: "", Cover: "c", Bvid: "BV1",
		Pubtime: stamp, Type: 2, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: video.ID, Cid: 5, Page: 2, From: "v", Part: "第二P", Duration: 65,
		Vid: "", Weblink: "", FirstFrame: "",
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	text, err := downloadedMessage(context.Background(), conn, t.TempDir(), part.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Video Download Completed", "标题&lt;A&gt;", "第二P", "Part 2", "00:01:05", "BV1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s in %s", want, text)
		}
	}
}

func openNotifyDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "notify.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestFreezeMessage(t *testing.T) {
	text := formatFileSize(0)
	if text != "0 MB" {
		t.Fatal(text)
	}
	if formatDuration(3661) != "01:01:01" {
		t.Fatal(formatDuration(3661))
	}
}
