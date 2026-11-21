package httpapi

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"mybili/internal/covercolor"
	"mybili/internal/dbtime"
	"mybili/internal/model"
	"mybili/internal/readable"
	"mybili/internal/syncbili"
)

func (a *App) subscriptionIndex(w http.ResponseWriter, r *http.Request) {
	var subs []model.Subscription
	if err := a.DB.Order("id").Find(&subs).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	ids := make([]int64, len(subs))
	for i, sub := range subs {
		ids[i] = sub.ID
	}
	covers, err := a.coversFor(model.CoverableSubscription, ids)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	list := make([]any, 0, len(subs))
	for _, sub := range subs {
		item := subscriptionJSON(sub, covers[sub.ID])
		item["dominant_color"] = dominantColor(a.Config.DataDir, sub, covers[sub.ID])
		list = append(list, item)
	}
	writeJSON(w, http.StatusOK, list)
}

func dominantColor(dataDir string, sub model.Subscription, cover *model.Cover) any {
	if sub.Type != "up" || cover == nil {
		return nil
	}
	path := coverFile(dataDir, cover.Path)
	if path == "" {
		return nil
	}
	hex, ok := covercolor.Dominant(path)
	if !ok {
		return nil
	}
	return hex
}

func coverFile(dataDir, stored string) string {
	stored = strings.TrimSpace(stored)
	stored = strings.TrimPrefix(stored, "/storage/")
	stored = strings.TrimPrefix(stored, "storage/")
	stored = strings.TrimPrefix(stored, "/")
	if stored == "" || strings.Contains(stored, "..") {
		return ""
	}
	return filepath.Join(dataDir, "public", filepath.FromSlash(stored))
}

func (a *App) subscriptionShow(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Subscription not found"})
		return
	}
	var sub model.Subscription
	if err := a.DB.First(&sub, id).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Subscription not found"})
		return
	}
	covers, _ := a.coversFor(model.CoverableSubscription, []int64{sub.ID})
	item := subscriptionJSON(sub, covers[sub.ID])
	var videos []model.Video
	_ = a.DB.Where("deleted_at IS NULL AND id IN (?)", a.DB.Model(&model.SubscriptionVideo{}).Select("video_id").Where("subscription_id = ?", sub.ID)).
		Order("pubtime DESC").Find(&videos).Error
	videoCovers, _ := a.coversFor(model.CoverableVideo, videoIDs(videos))
	nested := make([]any, 0, len(videos))
	for _, video := range videos {
		nested = append(nested, videoJSON(video, videoCovers[video.ID]))
	}
	item["videos"] = nested
	writeJSON(w, http.StatusOK, item)
}

func (a *App) subscriptionStore(w http.ResponseWriter, r *http.Request) {
	if a.Config.SettingReadOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "read-only mode"})
		return
	}
	var body struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "request format is incorrect"})
		return
	}
	row, err := syncbili.CreateSubscription(r.Context(), a.DB, body.Type, body.URL, time.Now())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": err.Error()})
		return
	}
	if err := syncbili.EnqueueSubscription(r.Context(), a.DB, row.ID, true, time.Now(), a.wake); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "Subscription added successfully", "id": row.ID})
}

func (a *App) subscriptionUpdate(w http.ResponseWriter, r *http.Request) {
	if a.Config.SettingReadOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "read-only mode"})
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Subscription not found"})
		return
	}
	var body struct {
		Status *int `json:"status"`
	}
	if err := readJSON(r, &body); err != nil || body.Status == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "request format is incorrect"})
		return
	}
	if *body.Status != 0 && *body.Status != 1 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "status is incorrect"})
		return
	}
	res := a.DB.WithContext(r.Context()).Model(&model.Subscription{}).Where("id = ?", id).Updates(map[string]any{
		"status":     *body.Status,
		"updated_at": dbtime.DateTime{Time: time.Now()},
	})
	if res.Error != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Subscription not found"})
		return
	}
	var row model.Subscription
	_ = a.DB.First(&row, id).Error
	covers, _ := a.coversFor(model.CoverableSubscription, []int64{id})
	writeJSON(w, http.StatusOK, subscriptionJSON(row, covers[id]))
}

func (a *App) subscriptionDestroy(w http.ResponseWriter, r *http.Request) {
	if a.Config.SettingReadOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "read-only mode"})
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Subscription not found"})
		return
	}
	var row model.Subscription
	if err := a.DB.WithContext(r.Context()).First(&row, id).Error; err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Subscription not found"})
		return
	}
	ids, err := syncbili.DeleteSubscription(r.Context(), a.DB, id, time.Now())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	if err := readable.UnlinkFolder(r.Context(), a.DB, a.Config.DataDir, a.Config.HumanReadableDir, row.Name, ids); err != nil {
		log.Printf("可读链接清理失败: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "Subscription deleted successfully"})
}

func (a *App) wake(ctx context.Context, queue string) error {
	if a.Tasks == nil {
		return nil
	}
	return a.Tasks.Wake(ctx, queue)
}
