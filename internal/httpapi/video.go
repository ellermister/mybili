package httpapi

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
	"mybili/internal/download"
	"mybili/internal/model"
	"mybili/internal/readable"
)

func (a *App) videoIndex(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	if page < 1 {
		page = 1
	}
	perPage := queryInt(r, "page_size", 30)
	if perPage < 1 {
		perPage = 30
	}
	query := a.DB.Model(&model.Video{}).Where("deleted_at IS NULL")
	if text := strings.TrimSpace(r.URL.Query().Get("query")); text != "" {
		lower := strings.ToLower(text)
		if strings.HasPrefix(lower, "bv") {
			query = query.Where("bvid = ?", text)
		} else if _, err := strconv.ParseInt(text, 10, 64); err == nil {
			query = query.Where("id = ?", text)
		} else {
			query = query.Where("title LIKE ?", "%"+text+"%")
		}
	}
	switch r.URL.Query().Get("status") {
	case "valid":
		query = query.Where("invalid = 0")
	case "invalid":
		query = query.Where("invalid = 1")
	case "frozen":
		query = query.Where("frozen = 1")
	}
	switch r.URL.Query().Get("downloaded") {
	case "yes":
		query = query.Where("video_downloaded_num > 0")
	case "no":
		query = query.Where("video_downloaded_num = 0")
	}
	switch r.URL.Query().Get("multi_part") {
	case "yes":
		query = query.Where("id IN (?)", a.DB.Model(&model.VideoPart{}).Select("video_id").Where("page > 1"))
	case "no":
		query = query.Where("id NOT IN (?)", a.DB.Model(&model.VideoPart{}).Select("video_id").Where("page = 1"))
	}
	if favID := r.URL.Query().Get("fav_id"); favID != "" {
		id, err := strconv.ParseInt(favID, 10, 64)
		if err == nil && id > 0 {
			query = query.Where("id IN (?)", a.DB.Model(&model.FavoriteListVideo{}).Select("video_id").Where("favorite_list_id = ?", id))
		} else if err == nil && id < 0 {
			query = query.Where("id IN (?)", a.DB.Model(&model.SubscriptionVideo{}).Select("video_id").Where("subscription_id = ?", -id))
		}
	}

	stat := map[string]int64{}
	countOf := func(extra string, args ...any) int64 {
		var count int64
		q := query.Session(&gorm.Session{})
		if extra != "" {
			q = q.Where(extra, args...)
		}
		_ = q.Count(&count).Error
		return count
	}
	stat["count"] = countOf("")
	stat["downloaded"] = countOf("video_downloaded_num > 0")
	stat["invalid"] = countOf("invalid = 1")
	stat["valid"] = countOf("invalid = 0")
	stat["frozen"] = countOf("frozen = 1")

	var videos []model.Video
	if err := query.Session(&gorm.Session{}).Offset((page - 1) * perPage).Limit(perPage).Find(&videos).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	covers, err := a.coversFor(model.CoverableVideo, videoIDs(videos))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	list := make([]any, 0, len(videos))
	for _, video := range videos {
		list = append(list, videoJSON(video, covers[video.ID]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"stat": stat, "list": list})
}

func (a *App) videoDestroy(w http.ResponseWriter, r *http.Request) {
	if a.Config.SettingReadOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "read-only mode"})
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusOK, map[string]any{"code": 1, "message": "Video deletion failed"})
		return
	}
	var body struct {
		ExtendIDs []int64 `json:"extend_ids"`
		Permanent *bool   `json:"permanent"`
		Requeue   bool    `json:"requeue"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": 1, "message": "Video deletion failed"})
		return
	}
	permanent := true
	if body.Permanent != nil {
		permanent = *body.Permanent
	}
	ids := append([]int64{id}, body.ExtendIDs...)
	deleted, err := download.DeleteVideos(r.Context(), a.DB, a.Config.DataDir, ids, download.DeleteOptions{
		Permanent: permanent,
		Requeue:   body.Requeue,
	}, time.Now(), a.wake)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 1, "message": err.Error()})
		return
	}
	if len(deleted) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"code": 1, "message": "Video deletion failed"})
		return
	}
	if err := readable.UnlinkVideos(r.Context(), a.DB, a.Config.DataDir, a.Config.HumanReadableDir, deleted); err != nil {
		log.Printf("可读链接清理失败: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"code":        0,
		"message":     "Video deleted successfully",
		"deleted_ids": deleted,
	})
}

func (a *App) videoShow(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var video model.Video
	if err := a.DB.Where("deleted_at IS NULL").First(&video, id).Error; err != nil {
		http.NotFound(w, r)
		return
	}
	covers, err := a.coversFor(model.CoverableVideo, []int64{video.ID})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	item := videoJSON(video, covers[video.ID])

	var parts []model.VideoPart
	_ = a.DB.Where("video_id = ?", video.ID).Order("page").Find(&parts).Error
	if video.IsAudio() {
		var audio model.AudioPart
		if err := a.DB.Where("video_id = ?", video.ID).First(&audio).Error; err == nil {
			path := ""
			if audio.AudioDownloadPath != nil {
				path = *audio.AudioDownloadPath
			}
			item["audio_part"] = map[string]any{
				"id":                  audio.ID,
				"video_id":            audio.VideoID,
				"sid":                 audio.Sid,
				"duration":            audio.Duration,
				"audio_downloaded_at": dbtime.FormatLocal(dbtime.FromPtr(audio.AudioDownloadedAt)),
				"audio_download_path": audio.AudioDownloadPath,
				"audio_download_url":  publicURL(path),
			}
			item["video_parts"] = []any{map[string]any{
				"id":         audio.Sid,
				"part":       1,
				"url":        publicURL(path),
				"title":      video.Title,
				"downloaded": path != "",
			}}
		} else {
			item["video_parts"] = []any{}
		}
	} else {
		viewParts := make([]any, 0, len(parts))
		for _, part := range parts {
			path := ""
			if part.VideoDownloadPath != nil {
				path = *part.VideoDownloadPath
			}
			title := part.Part
			if title == "" {
				title = "P" + strconv.Itoa(part.Page)
			}
			viewParts = append(viewParts, map[string]any{
				"id":         part.Cid,
				"part":       part.Page,
				"url":        publicURL(path),
				"title":      title,
				"downloaded": path != "",
			})
		}
		item["video_parts"] = viewParts
	}

	var count int64
	if len(parts) > 0 {
		cids := make([]int64, len(parts))
		for i, part := range parts {
			cids[i] = part.Cid
		}
		_ = a.DB.Model(&model.Danmaku{}).Where("cid IN ?", cids).Count(&count).Error
	}
	item["danmaku_count"] = count
	item["favorite"] = a.videoFavorites(video.ID)
	item["subscriptions"] = a.videoSubscriptions(video.ID)
	item["upper"] = a.videoUpper(video.UpperID)
	writeJSON(w, http.StatusOK, item)
}

func (a *App) videoFavorites(videoID int64) []any {
	var favorites []model.FavoriteList
	err := a.DB.Where("id IN (?)", a.DB.Model(&model.FavoriteListVideo{}).Select("favorite_list_id").Where("video_id = ?", videoID)).Find(&favorites).Error
	if err != nil || len(favorites) == 0 {
		return []any{}
	}
	ids := make([]int64, len(favorites))
	for i, fav := range favorites {
		ids[i] = fav.ID
	}
	covers, _ := a.coversFor(model.CoverableFavoriteList, ids)
	list := make([]any, 0, len(favorites))
	for _, fav := range favorites {
		list = append(list, favoriteJSON(fav, covers[fav.ID]))
	}
	return list
}

func (a *App) videoSubscriptions(videoID int64) []any {
	var subs []model.Subscription
	err := a.DB.Where("id IN (?)", a.DB.Model(&model.SubscriptionVideo{}).Select("subscription_id").Where("video_id = ?", videoID)).Find(&subs).Error
	if err != nil || len(subs) == 0 {
		return []any{}
	}
	ids := make([]int64, len(subs))
	for i, sub := range subs {
		ids[i] = sub.ID
	}
	covers, _ := a.coversFor(model.CoverableSubscription, ids)
	list := make([]any, 0, len(subs))
	for _, sub := range subs {
		list = append(list, subscriptionJSON(sub, covers[sub.ID]))
	}
	return list
}

func (a *App) videoUpper(mid *int64) any {
	if mid == nil {
		return nil
	}
	var upper model.Upper
	if err := a.DB.Where("mid = ?", *mid).First(&upper).Error; err != nil {
		return nil
	}
	covers, _ := a.coversFor(model.CoverableUpper, []int64{upper.ID})
	return upperJSON(upper, covers[upper.ID])
}

func (a *App) progress(w http.ResponseWriter, r *http.Request) {
	var videos []model.Video
	if err := a.DB.Where("deleted_at IS NULL").Order("COALESCE(fav_time, created_at) DESC").Find(&videos).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	covers, err := a.coversFor(model.CoverableVideo, videoIDs(videos))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	list := make([]any, 0, len(videos))
	var downloaded, invalid, valid, frozen int64
	for _, video := range videos {
		list = append(list, lightweightJSON(video, covers[video.ID], dbtime.FormatISO(dbtime.FromPtr(video.CreatedAt))))
		if video.VideoDownloadedNum > 0 {
			downloaded++
		}
		if video.Invalid == 1 {
			invalid++
		} else {
			valid++
		}
		if video.Frozen == 1 {
			frozen++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": list,
		"stat": map[string]int64{
			"count":      int64(len(videos)),
			"downloaded": downloaded,
			"invalid":    invalid,
			"valid":      valid,
			"frozen":     frozen,
		},
	})
}

func (a *App) danmaku(w http.ResponseWriter, r *http.Request) {
	cid := r.URL.Query().Get("id")
	if cid == "" {
		writeJSON(w, http.StatusOK, map[string]any{"code": 1, "message": "CID parameter cannot be empty", "data": []any{}})
		return
	}
	var rows []model.Danmaku
	if err := a.DB.Where("cid = ?", cid).Find(&rows).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	// 转换在 danmaku 包，这里只组响应。真正的播放器格式由 handler 文件完成。
	_ = rows
	writeJSON(w, http.StatusOK, map[string]any{"code": 0, "data": a.playerDanmaku(rows)})
}

func videoIDs(videos []model.Video) []int64 {
	ids := make([]int64, len(videos))
	for i, video := range videos {
		ids[i] = video.ID
	}
	return ids
}

func atoiDefault(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
