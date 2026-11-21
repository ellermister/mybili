package usage

import (
	"context"
	"crypto/hmac"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/db"
	"mybili/internal/migrate"
	"mybili/internal/model"
)

func TestPostSessionSignsBody(t *testing.T) {
	const secret = "test-secret-value"
	body := []byte(`{"install_id":"11111111-1111-4111-8111-111111111111","session_id":"22222222-2222-4222-8222-222222222222","platform":"linux","version":"1.0.0","duration_sec":0,"event":"start"}`)
	var gotPath, gotSig, gotKey, gotNonce, gotTS, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("X-MJJ-Key-Id")
		gotNonce = r.Header.Get("X-MJJ-Nonce")
		gotTS = r.Header.Get("X-MJJ-Timestamp")
		gotSig = r.Header.Get("X-MJJ-Signature")
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	if err := postSession(context.Background(), server.Client(), server.URL+"/api/public/software/mybili/session", secret, body); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/public/software/mybili/session" || gotKey != "v1" || gotBody != string(body) {
		t.Fatalf("path=%s key=%s body=%s", gotPath, gotKey, gotBody)
	}
	if len(gotNonce) < 8 || len(gotNonce) > 64 || gotTS == "" {
		t.Fatalf("nonce=%s ts=%s", gotNonce, gotTS)
	}
	want := signHex(secret, signPayload(http.MethodPost, gotPath, gotTS, gotNonce, gotBody))
	if !hmac.Equal([]byte(want), []byte(gotSig)) {
		t.Fatalf("sig=%s want=%s", gotSig, want)
	}
}

func TestInstallIDStaysAcrossSessions(t *testing.T) {
	conn := openUsageDB(t)
	first, err := loadInstallID(conn)
	if err != nil {
		t.Fatal(err)
	}
	second, err := loadInstallID(conn)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || !clientIDPattern.MatchString(first) {
		t.Fatalf("install id %q %q", first, second)
	}
	var row model.Setting
	if err := conn.Where("name = ?", installSetting).First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Value == nil || !strings.Contains(*row.Value, first) {
		t.Fatalf("stored=%v", row.Value)
	}
}

func TestReporterSendsStartThenEnd(t *testing.T) {
	var events []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		events = append(events, string(raw))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	reporter := &Reporter{
		installID: "11111111-1111-4111-8111-111111111111",
		sessionID: "22222222-2222-4222-8222-222222222222",
		started:   time.Now().Add(-2 * time.Minute),
		client:    server.Client(),
		url:       server.URL + "/api/public/software/mybili/session",
		secret:    "test-secret-value",
		version:   "1.0.0",
		platform:  "linux",
	}
	reporter.send(context.Background(), "start")
	reporter.send(context.Background(), "end")
	if len(events) != 2 || !strings.Contains(events[0], `"event":"start"`) || !strings.Contains(events[0], `"duration_sec":0`) {
		t.Fatalf("start=%v", events)
	}
	if !strings.Contains(events[1], `"event":"end"`) || strings.Contains(events[1], `"duration_sec":0`) {
		t.Fatalf("end=%s", events[1])
	}
}

func openUsageDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "usage.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}
