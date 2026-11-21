package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mybili/internal/config"
)

func TestTelegramConnectionRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	app := &App{DB: openQueueDB(t), Config: config.Config{}}
	body := `{"telegram_bot_token":"token","telegram_chat_id":"chat","telegram_bot_api_url":"` + server.URL + `"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/settings/test-telegram", strings.NewReader(body))
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"result":true`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	app.Config.SettingReadOnly = true
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/settings/test-telegram", strings.NewReader(body))
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("readonly status=%d", rec.Code)
	}
}
