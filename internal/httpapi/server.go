package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) Handler(webDist string) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Get("/up", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	router.Route("/api", func(api chi.Router) {
		api.Get("/fav", a.favIndex)
		api.Get("/fav/{id}", a.favShow)
		api.Get("/fav/{id}/videos", a.favVideos)
		api.Get("/videos", a.videoIndex)
		api.Get("/videos/{id}", a.videoShow)
		api.Delete("/videos/{id}", a.videoDestroy)
		api.Post("/videos/{id}/danmaku/refresh", a.refreshDanmaku)
		api.Get("/danmaku", a.danmaku)
		api.Get("/progress", a.progress)
		api.Get("/cookie/exist", a.cookieExist)
		api.Get("/cookie/status", a.cookieStatus)
		api.Post("/cookie/upload", a.cookieUpload)
		api.Get("/settings", a.getSettings)
		api.Post("/settings", a.saveSettings)
		api.Post("/settings/test-telegram", a.testTelegram)
		api.Get("/subscription", a.subscriptionIndex)
		api.Get("/subscription/{id}", a.subscriptionShow)
		api.Post("/subscription", a.subscriptionStore)
		api.Put("/subscription/{id}", a.subscriptionUpdate)
		api.Delete("/subscription/{id}", a.subscriptionDestroy)
		api.Get("/system/info", a.systemInfo)
		api.Get("/system/media-usage", a.mediaUsage)
		api.Post("/system/migrate", a.migrateDB)
		api.Get("/download-queue", a.taskQueue)
		api.Get("/download-queue/stat", a.taskQueueStat)
		api.Post("/download-queue/clear", a.taskClear)
		api.Get("/download-queue/{id}/output", a.taskOutput)
		api.Post("/download-queue/{id}/cancel", a.taskCancel)
		api.Post("/download-queue/{id}/retry", a.taskRetry)
	})
	publicDir := filepath.Join(a.Config.DataDir, "public")
	router.Handle("/storage/*", http.StripPrefix("/storage/", http.FileServer(http.Dir(publicDir))))
	router.NotFound(spaHandler(webDist).ServeHTTP)
	return router
}

func spaHandler(dist string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dist == "" {
			http.NotFound(w, r)
			return
		}
		clean := filepath.Clean(r.URL.Path)
		if strings.Contains(clean, "..") {
			http.NotFound(w, r)
			return
		}
		target := filepath.Join(dist, clean)
		info, err := os.Stat(target)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
		index := filepath.Join(dist, "index.html")
		if _, err := os.Stat(index); err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}
