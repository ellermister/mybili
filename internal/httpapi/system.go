package httpapi

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"mybili/internal/dbtime"
	"mybili/internal/migrate"
	"mybili/internal/model"
)

func (a *App) systemInfo(w http.ResponseWriter, r *http.Request) {
	var favs, videos, parts int64
	_ = a.DB.Model(&model.FavoriteList{}).Count(&favs).Error
	_ = a.DB.Model(&model.Video{}).Where("deleted_at IS NULL").Count(&videos).Error
	_ = a.DB.Model(&model.VideoPart{}).Count(&parts).Error
	info, _ := os.Stat(a.Config.DBPath)
	var size int64
	if info != nil {
		size = info.Size()
	}
	var version string
	_ = a.DB.Raw("select sqlite_version()").Scan(&version).Error
	writeJSON(w, http.StatusOK, map[string]any{
		"app_version":      a.Config.Version,
		"go_version":       runtime.Version(),
		"database_version": version,
		"timezone":         a.Config.Timezone,
		"time_now":         dbtime.Now().Format("2006-01-02 15:04:05"),
		"database_usage": map[string]any{
			"favorite_lists": favs,
			"videos":         videos,
			"video_parts":    parts,
			"db_size":        size,
		},
		"media_usage": nil,
	})
}

func (a *App) mediaUsage(w http.ResponseWriter, r *http.Request) {
	root := filepath.Join(a.Config.DataDir, "public")
	writeJSON(w, http.StatusOK, map[string]any{
		"videos_size": humanSize(dirSize(filepath.Join(root, "videos"))),
		"images_size": humanSize(dirSize(filepath.Join(root, "images"))),
	})
}

func (a *App) migrateDB(w http.ResponseWriter, r *http.Request) {
	if err := migrate.Auto(a.DB); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "database is aligned"})
}

func dirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

func humanSize(n int64) string {
	if n <= 0 {
		return "0"
	}
	units := []string{"B", "K", "M", "G", "T"}
	value := float64(n)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d%s", n, units[unit])
	}
	return fmt.Sprintf("%.1f%s", value, units[unit])
}
