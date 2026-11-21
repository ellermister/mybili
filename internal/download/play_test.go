package download

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mybili/internal/taskrun"
)

type memLogs struct {
	lines []string
}

func (m *memLogs) AppendLog(_ context.Context, _ int64, line string) error {
	m.lines = append(m.lines, line)
	return nil
}

func (m *memLogs) SetProgress(context.Context, int64, taskrun.Progress) error { return nil }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestConnectFailureUsesBackupURL(t *testing.T) {
	var hosts []string
	logs := &memLogs{}
	dir := t.TempDir()
	dest := filepath.Join(dir, "v.m4s")
	runner := Runner{
		Logs: logs,
		CDN: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			hosts = append(hosts, req.URL.Host)
			if strings.Contains(req.URL.Host, "mcdn") {
				return nil, &netTimeout{msg: "dial tcp 1.2.3.4:8082: i/o timeout"}
			}
			return &http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: 4,
				Body:          io.NopCloser(strings.NewReader("m4s!")),
				Header:        make(http.Header),
			}, nil
		})},
	}
	err := runner.fetchFirst(context.Background(), 1, "视频", []string{
		"https://a.mcdn.bilivideo.cn:8082/v1/resource/a.m4s?e=1",
		"https://upos-sz-mirrorhw.bilivideo.com/a.m4s?e=1",
		"https://upos-sz-mirrorali.bilivideo.com/should-not-use.m4s",
	}, dest, 0, "SESSDATA=abc", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 2 || hosts[1] != "upos-sz-mirrorhw.bilivideo.com" {
		t.Fatalf("hosts = %#v", hosts)
	}
	body, err := os.ReadFile(dest)
	if err != nil || string(body) != "m4s!" {
		t.Fatalf("file = %q %v", body, err)
	}
	text := strings.Join(logs.lines, "\n")
	if !strings.Contains(text, "i/o timeout") || !strings.Contains(text, "更换备用地址 https://upos-sz-mirrorhw.bilivideo.com/a.m4s") {
		t.Fatalf("logs = %s", text)
	}
}

func TestHTTPFailureDoesNotUseBackup(t *testing.T) {
	var hits int
	runner := Runner{CDN: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		hits++
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("blocked")),
			Header:     make(http.Header),
		}, nil
	})}}
	err := runner.fetchFirst(context.Background(), 1, "视频", []string{
		"https://upos-sz-mirrorhw.bilivideo.com/a.m4s",
		"https://upos-sz-mirrorali.bilivideo.com/a.m4s",
	}, filepath.Join(t.TempDir(), "v.m4s"), 0, "", nil)
	if err == nil || hits != 1 || !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("err = %v hits=%d", err, hits)
	}
}

func TestBackupConnectFailureFailsTask(t *testing.T) {
	var hits int
	runner := Runner{CDN: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		hits++
		return nil, &netTimeout{msg: "dial tcp: i/o timeout"}
	})}}
	err := runner.fetchFirst(context.Background(), 1, "视频", []string{
		"https://a.mcdn.bilivideo.cn:8082/a.m4s",
		"https://upos-sz-mirrorhw.bilivideo.com/a.m4s",
	}, filepath.Join(t.TempDir(), "v.m4s"), 0, "", nil)
	if err == nil || hits != 2 || !connectFailed(err) {
		t.Fatalf("err = %v hits=%d", err, hits)
	}
}

func TestConnectFailedOnlyForDial(t *testing.T) {
	dial := &url.Error{Op: "Get", URL: "https://a", Err: &netTimeout{msg: "dial tcp: i/o timeout"}}
	if !connectFailed(dial) {
		t.Fatal("dial should switch address")
	}
	if connectFailed(errText("下载 HTTP 403: blocked")) {
		t.Fatal("http status should fail the task")
	}
}

type errText string

func (e errText) Error() string { return string(e) }

type netTimeout struct{ msg string }

func (e *netTimeout) Error() string   { return e.msg }
func (e *netTimeout) Timeout() bool   { return true }
func (e *netTimeout) Temporary() bool { return true }
