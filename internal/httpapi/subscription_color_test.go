package httpapi

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mybili/internal/config"
	"mybili/internal/model"
)

func TestSubscriptionDominantColorOnlyForUp(t *testing.T) {
	conn := openQueueDB(t)
	dir := t.TempDir()
	rel := "images/face.png"
	writeSolidPNG(t, filepath.Join(dir, "public", rel), color.RGBA{R: 220, G: 30, B: 40, A: 255})
	cover := model.Cover{Type: "avatar", URL: "u", Filename: "face.png", Path: rel, MimeType: "image/png"}
	if err := conn.Create(&cover).Error; err != nil {
		t.Fatal(err)
	}
	up := model.Subscription{Type: "up", Mid: "1", Name: "UP"}
	series := model.Subscription{Type: "seasons", Mid: "2", Name: "合集"}
	if err := conn.Create(&up).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&series).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.Coverable{CoverID: cover.ID, CoverableID: up.ID, CoverableType: model.CoverableSubscription}).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.Coverable{CoverID: cover.ID, CoverableID: series.ID, CoverableType: model.CoverableSubscription}).Error; err != nil {
		t.Fatal(err)
	}
	app := &App{DB: conn, Config: config.Config{DataDir: dir}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/subscription", nil)
	app.Handler("").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"dominant_color":"#dc1e28"`) || strings.Count(body, `"dominant_color":"#dc1e28"`) != 1 {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, `"dominant_color":null`) {
		t.Fatalf("body=%s", body)
	}
}

func writeSolidPNG(t *testing.T, path string, c color.Color) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, c)
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
}
