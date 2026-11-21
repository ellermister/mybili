package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mybili/internal/config"
)

func TestSystemInfoOmitsWebsiteID(t *testing.T) {
	conn := openQueueDB(t)
	app := &App{DB: conn, Config: config.Config{Version: "1.0.0"}}
	body := getInfo(t, app)
	if strings.Contains(body, "website_id") || !strings.Contains(body, `"app_version":"1.0.0"`) {
		t.Fatalf("body=%s", body)
	}
}

func getInfo(t *testing.T, app *App) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/info", nil)
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}
