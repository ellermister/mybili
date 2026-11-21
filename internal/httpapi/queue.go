package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/download"
	"mybili/internal/jobs"
	"mybili/internal/model"
)

func (a *App) taskQueue(w http.ResponseWriter, r *http.Request) {
	if a.DB == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": "database is not configured"})
		return
	}
	page := queryInt(r, "page", 1)
	perPage := queryInt(r, "per_page", 30)
	if perPage > 100 {
		perPage = 100
	}
	store := jobs.Store{DB: a.DB}
	rows, total, err := store.List(r.Context(), parseStatuses(r.URL.Query().Get("status")), page, perPage)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	counts, err := store.CountByStatus(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	list := a.presentTasks(rows)
	writeJSON(w, http.StatusOK, map[string]any{
		"list":  list,
		"total": total,
		"stat":  taskStat(counts),
	})
}

func (a *App) taskQueueStat(w http.ResponseWriter, r *http.Request) {
	if a.DB == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": "database is not configured"})
		return
	}
	counts, err := jobs.Store{DB: a.DB}.CountByStatus(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, taskStat(counts))
}

func (a *App) taskCancel(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false})
		return
	}
	ok, err := jobs.Store{DB: a.DB}.Cancel(r.Context(), id, time.Now())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": ok})
}

func (a *App) taskRetry(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false})
		return
	}
	job, ok, err := jobs.Store{DB: a.DB}.Retry(r.Context(), id, time.Now())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
		return
	}
	if ok && a.Tasks != nil {
		if err := a.Tasks.Wake(r.Context(), job.Queue); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": ok})
}

func (a *App) taskOutput(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "任务不存在"})
		return
	}
	lines := []string{}
	done, total := 0, 0
	text := ""
	if a.Output != nil {
		lines, done, total, text, err = a.Output(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
			return
		}
	}
	if lines == nil {
		lines = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"lines": lines,
		"progress": map[string]any{
			"done":  done,
			"total": total,
			"text":  text,
		},
	})
}

func (a *App) refreshDanmaku(w http.ResponseWriter, r *http.Request) {
	if a.DB == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false})
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false})
		return
	}
	var wake func(ctx context.Context, queue string) error
	if a.Tasks != nil {
		wake = a.Tasks.Wake
	}
	queued, err := download.RefreshDanmaku(r.Context(), a.DB, id, time.Now(), wake)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 1, "message": err.Error(), "parts_queued": 0})
		return
	}
	if queued == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"code": 1, "message": "没有分 P 可以刷新", "parts_queued": 0})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"code": 0, "parts_queued": queued})
}

func (a *App) taskClear(w http.ResponseWriter, r *http.Request) {
	if a.DB == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false})
		return
	}
	res := a.DB.WithContext(r.Context()).Where("1 = 1").Delete(&model.Job{})
	if res.Error != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": res.Error.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "deleted": res.RowsAffected})
}

// taskPayload 是任务表里存的正文。字段按各任务自己的 JSON 取用。
type taskPayload struct {
	ID      int64  `json:"id"`
	VideoID int64  `json:"video_id"`
	Page    int    `json:"page"`
	PartID  int64  `json:"part_id"`
	Kind    string `json:"kind"`
	URL     string `json:"url"`
	Bvid    string `json:"bvid"`
	FavID   int64  `json:"fav_id"`
	PullAll bool   `json:"pull_all"`
}

// taskSubject 是给前端拼文案用的稿件数据，不含语言。
type taskSubject struct {
	Title   string `json:"title,omitempty"`
	Bvid    string `json:"bvid,omitempty"`
	Page    int    `json:"page,omitempty"`
	Part    string `json:"part,omitempty"`
	Kind    string `json:"kind,omitempty"`
	PullAll bool   `json:"pull_all,omitempty"`
}

type taskItem struct {
	ID         int64            `json:"id"`
	Queue      string           `json:"queue"`
	Type       string           `json:"type"`
	UniqueKey  string           `json:"unique_key"`
	Status     string           `json:"status"`
	Subject    *taskSubject     `json:"subject,omitempty"`
	ImageURL   string           `json:"image_url"`
	Href       string           `json:"href"`
	Payload    string           `json:"payload"`
	ErrorMsg   *string          `json:"error_msg"`
	RetryCount int              `json:"retry_count"`
	MaxRetries int              `json:"max_retries"`
	CreatedAt  *dbtime.DateTime `json:"created_at"`
	StartedAt  *dbtime.DateTime `json:"started_at"`
	FinishedAt *dbtime.DateTime `json:"finished_at"`
	UpdatedAt  *dbtime.DateTime `json:"updated_at"`
}

func (a *App) presentTasks(rows []model.Job) []taskItem {
	bodies := make([]taskPayload, len(rows))
	parsed := make([]bool, len(rows))
	var videoIDs, upperIDs, favIDs, subIDs, partIDs []int64
	var bvids []string
	needsFavThumb := false
	for i, row := range rows {
		if json.Unmarshal([]byte(row.Payload), &bodies[i]) != nil {
			continue
		}
		parsed[i] = true
		body := bodies[i]
		switch row.Type {
		case "download-video", "download-danmaku", "download-audio":
			if body.VideoID > 0 {
				videoIDs = append(videoIDs, body.VideoID)
			}
		case "sync-audio":
			if body.ID > 0 {
				videoIDs = append(videoIDs, body.ID)
			}
		case "sync-video":
			if bvid := strings.TrimSpace(body.Bvid); bvid != "" {
				bvids = append(bvids, bvid)
			}
		case "subscription":
			if body.ID > 0 {
				subIDs = append(subIDs, body.ID)
			}
		case "cover":
			switch body.Kind {
			case "video":
				if body.ID > 0 {
					videoIDs = append(videoIDs, body.ID)
				}
			case "avatar":
				if body.ID > 0 {
					upperIDs = append(upperIDs, body.ID)
				}
			case "favorite":
				if body.ID > 0 {
					favIDs = append(favIDs, body.ID)
				}
			}
		case "fix-invalid-fav":
			if body.FavID > 0 {
				favIDs = append(favIDs, body.FavID)
			}
		case "readable-link":
			if body.PartID > 0 {
				partIDs = append(partIDs, body.PartID)
			}
		case "fav-list", "fav-videos":
			needsFavThumb = true
		}
	}
	parts := loadParts(a.DB, partIDs)
	for _, part := range parts {
		if part.VideoID > 0 {
			videoIDs = append(videoIDs, part.VideoID)
		}
	}
	videos := loadVideos(a.DB, videoIDs)
	byBvid := loadVideosByBvid(a.DB, bvids)
	for _, video := range byBvid {
		videos[video.ID] = video
		videoIDs = append(videoIDs, video.ID)
	}
	videoCovers, _ := a.coversFor(model.CoverableVideo, videoIDs)
	uppers := loadUppers(a.DB, upperIDs)
	favs := loadFavorites(a.DB, favIDs)
	favCovers, _ := a.coversFor(model.CoverableFavoriteList, favIDs)
	subs := loadSubscriptions(a.DB, subIDs)
	subCovers, _ := a.coversFor(model.CoverableSubscription, subIDs)
	favThumb := ""
	if needsFavThumb {
		favThumb = a.favoriteTaskCover()
	}
	list := make([]taskItem, 0, len(rows))
	for i, row := range rows {
		item := newTaskItem(row)
		body := bodies[i]
		switch row.Type {
		case "download-video", "download-danmaku", "download-audio":
			if body.VideoID > 0 {
				item.Href = "/video/" + strconv.FormatInt(body.VideoID, 10)
			}
			subject := taskSubject{Page: body.Page}
			if video, ok := videos[body.VideoID]; ok {
				subject.Title = video.Title
				subject.Bvid = video.Bvid
				item.ImageURL = videoImage(video, videoCovers)
			}
			setSubject(&item, subject)
		case "sync-video":
			video, ok := byBvid[strings.TrimSpace(body.Bvid)]
			subject := taskSubject{Bvid: strings.TrimSpace(body.Bvid)}
			if ok {
				subject.Title = video.Title
				item.Href = "/video/" + strconv.FormatInt(video.ID, 10)
				item.ImageURL = videoImage(video, videoCovers)
			}
			setSubject(&item, subject)
		case "sync-audio":
			if body.ID > 0 {
				item.Href = "/video/" + strconv.FormatInt(body.ID, 10)
			}
			if video, ok := videos[body.ID]; ok {
				item.ImageURL = videoImage(video, videoCovers)
				setSubject(&item, taskSubject{Title: video.Title, Bvid: video.Bvid})
			}
		case "readable-link":
			part, ok := parts[body.PartID]
			if !ok {
				break
			}
			subject := taskSubject{Page: part.Page, Part: part.Part}
			if video, ok := videos[part.VideoID]; ok {
				subject.Title = video.Title
				subject.Bvid = video.Bvid
				item.Href = "/video/" + strconv.FormatInt(video.ID, 10)
				item.ImageURL = videoImage(video, videoCovers)
			}
			setSubject(&item, subject)
		case "subscription":
			if body.ID > 0 {
				item.Href = "/subscription/" + strconv.FormatInt(body.ID, 10)
			}
			subject := taskSubject{PullAll: body.PullAll}
			if sub, ok := subs[body.ID]; ok {
				subject.Title = sub.Name
				item.ImageURL = sub.Cover
				if cover := subCovers[body.ID]; cover != nil {
					if url, ok := publicURL(cover.Path).(string); ok && url != "" {
						item.ImageURL = url
					}
				}
			}
			setSubject(&item, subject)
		case "cover":
			item.ImageURL = body.URL
			subject := taskSubject{Kind: body.Kind}
			switch body.Kind {
			case "video":
				if body.ID > 0 {
					item.Href = "/video/" + strconv.FormatInt(body.ID, 10)
				}
				if video, ok := videos[body.ID]; ok {
					subject.Title = video.Title
					subject.Bvid = video.Bvid
				}
			case "avatar":
				if upper, ok := uppers[body.ID]; ok {
					subject.Title = upper.Name
				}
			case "favorite":
				if body.ID > 0 {
					item.Href = "/fav/" + strconv.FormatInt(body.ID, 10)
				}
				if fav, ok := favs[body.ID]; ok {
					subject.Title = fav.Title
				}
			}
			setSubject(&item, subject)
		case "fix-invalid-fav":
			subject := taskSubject{Page: body.Page}
			if fav, ok := favs[body.FavID]; ok {
				subject.Title = fav.Title
				item.Href = "/fav/" + strconv.FormatInt(fav.ID, 10)
				item.ImageURL = favoriteImage(fav, favCovers)
			}
			setSubject(&item, subject)
		case "fav-list", "fav-videos":
			item.ImageURL = favThumb
			if row.Type == "fav-videos" && parsed[i] {
				setSubject(&item, taskSubject{Page: body.Page})
			}
		}
		list = append(list, item)
	}
	return list
}

func setSubject(item *taskItem, subject taskSubject) {
	if subject == (taskSubject{}) {
		return
	}
	item.Subject = &subject
}

func videoImage(video model.Video, covers map[int64]*model.Cover) string {
	if covers != nil {
		if cover := covers[video.ID]; cover != nil {
			if url, ok := publicURL(cover.Path).(string); ok && url != "" {
				return url
			}
		}
	}
	return strings.TrimSpace(video.Cover)
}

func favoriteImage(fav model.FavoriteList, covers map[int64]*model.Cover) string {
	if covers != nil {
		if cover := covers[fav.ID]; cover != nil {
			if url, ok := publicURL(cover.Path).(string); ok && url != "" {
				return url
			}
		}
	}
	return strings.TrimSpace(fav.Cover)
}

func loadVideos(db *gorm.DB, ids []int64) map[int64]model.Video {
	out := map[int64]model.Video{}
	if db == nil || len(ids) == 0 {
		return out
	}
	var rows []model.Video
	if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return out
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out
}

func loadVideosByBvid(db *gorm.DB, bvids []string) map[string]model.Video {
	out := map[string]model.Video{}
	if db == nil || len(bvids) == 0 {
		return out
	}
	var rows []model.Video
	if err := db.Where("bvid IN ?", bvids).Find(&rows).Error; err != nil {
		return out
	}
	for _, row := range rows {
		out[row.Bvid] = row
	}
	return out
}

func loadParts(db *gorm.DB, ids []int64) map[int64]model.VideoPart {
	out := map[int64]model.VideoPart{}
	if db == nil || len(ids) == 0 {
		return out
	}
	var rows []model.VideoPart
	if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return out
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out
}

func loadUppers(db *gorm.DB, ids []int64) map[int64]model.Upper {
	out := map[int64]model.Upper{}
	if db == nil || len(ids) == 0 {
		return out
	}
	var rows []model.Upper
	if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return out
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out
}

// favoriteTaskCover 取一个收藏夹封面，给收藏夹同步任务当缩略图。
// 优先用已经下到本地的图，没有再退回接口地址。
func (a *App) favoriteTaskCover() string {
	if a.DB == nil {
		return ""
	}
	var favs []model.FavoriteList
	if err := a.DB.Order("id").Find(&favs).Error; err != nil || len(favs) == 0 {
		return ""
	}
	ids := make([]int64, 0, len(favs))
	for _, fav := range favs {
		ids = append(ids, fav.ID)
	}
	covers, err := a.coversFor(model.CoverableFavoriteList, ids)
	if err == nil {
		for _, fav := range favs {
			cover := covers[fav.ID]
			if cover == nil {
				continue
			}
			if url, ok := publicURL(cover.Path).(string); ok && url != "" {
				return url
			}
		}
	}
	for _, fav := range favs {
		if strings.TrimSpace(fav.Cover) != "" {
			return fav.Cover
		}
	}
	return ""
}

func loadSubscriptions(db *gorm.DB, ids []int64) map[int64]model.Subscription {
	out := map[int64]model.Subscription{}
	if db == nil || len(ids) == 0 {
		return out
	}
	var rows []model.Subscription
	if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return out
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out
}

func loadFavorites(db *gorm.DB, ids []int64) map[int64]model.FavoriteList {
	out := map[int64]model.FavoriteList{}
	if db == nil || len(ids) == 0 {
		return out
	}
	var rows []model.FavoriteList
	if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return out
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out
}

func newTaskItem(row model.Job) taskItem {
	if row.ErrorMsg != nil && *row.ErrorMsg == "" {
		row.ErrorMsg = nil
	}
	return taskItem{
		ID:         row.ID,
		Queue:      row.Queue,
		Type:       row.Type,
		UniqueKey:  row.UniqueKey,
		Status:     row.Status,
		Payload:    row.Payload,
		ErrorMsg:   row.ErrorMsg,
		RetryCount: row.RetryCount,
		MaxRetries: row.MaxRetries,
		CreatedAt:  row.CreatedAt,
		StartedAt:  row.StartedAt,
		FinishedAt: row.FinishedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func taskStat(counts map[string]int64) map[string]int64 {
	keys := []string{
		jobs.StatusPending,
		jobs.StatusQueued,
		jobs.StatusRunning,
		jobs.StatusDone,
		jobs.StatusFailed,
		jobs.StatusCancelled,
	}
	stat := map[string]int64{}
	for _, key := range keys {
		stat[key] = counts[key]
	}
	return stat
}

func parseStatuses(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	allowed := map[string]bool{
		jobs.StatusPending:   true,
		jobs.StatusQueued:    true,
		jobs.StatusRunning:   true,
		jobs.StatusDone:      true,
		jobs.StatusFailed:    true,
		jobs.StatusCancelled: true,
	}
	var statuses []string
	seen := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		status := strings.TrimSpace(part)
		if !allowed[status] || seen[status] {
			continue
		}
		seen[status] = true
		statuses = append(statuses, status)
	}
	return statuses
}
