package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"mybili/internal/config"
	"mybili/internal/model"
)

// TaskWaker 在任务重新变成 pending 时通知 Feeder。
type TaskWaker interface {
	Wake(ctx context.Context, queue string) error
}

type App struct {
	DB     *gorm.DB
	Config config.Config
	Tasks  TaskWaker
	// Output 读取任务执行时写到 Redis 的输出。为空时接口返回空日志。
	Output func(ctx context.Context, id int64) (lines []string, done, total int, text string, err error)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

func readJSON(r *http.Request, dest any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	return nil
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func pathInt64(r *http.Request, key string) (int64, error) {
	parsed, err := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func notReady(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{
		"message": "this operation is still migrating, and the current version cannot be executed",
	})
}

func (a *App) coversFor(coverableType string, ids []int64) (map[int64]*model.Cover, error) {
	result := map[int64]*model.Cover{}
	if len(ids) == 0 {
		return result, nil
	}
	var links []model.Coverable
	if err := a.DB.Where("coverable_type = ? AND coverable_id IN ?", coverableType, ids).Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return result, nil
	}
	coverIDs := make([]int64, 0, len(links))
	for _, link := range links {
		coverIDs = append(coverIDs, link.CoverID)
	}
	var covers []model.Cover
	if err := a.DB.Where("id IN ?", coverIDs).Find(&covers).Error; err != nil {
		return nil, err
	}
	byID := map[int64]*model.Cover{}
	for i := range covers {
		byID[covers[i].ID] = &covers[i]
	}
	for _, link := range links {
		if cover, ok := byID[link.CoverID]; ok {
			result[link.CoverableID] = cover
		}
	}
	return result, nil
}
