package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mybili/internal/config"
	"mybili/internal/model"
)

func TestSettingsQualityReplacesSizeExclude(t *testing.T) {
	conn := openQueueDB(t)
	app := &App{DB: conn, Config: config.Config{}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "size_exclude") || !strings.Contains(body, `"video_quality":"1"`) || !strings.Contains(body, `"source_quality":{}`) {
		t.Fatalf("body=%s", body)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader(`{"video_quality":"3","size_exclude":{"type":"1GB","custom_size":1}}`))
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status=%d body=%s", rec.Code, rec.Body.String())
	}
	var row model.Setting
	if err := conn.Where("name = ?", "video_quality").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Value == nil || *row.Value != `"3"` {
		t.Fatalf("stored %v", row.Value)
	}
	var count int64
	if err := conn.Model(&model.Setting{}).Where("name = ?", "size_exclude").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("size_exclude saved")
	}
}
