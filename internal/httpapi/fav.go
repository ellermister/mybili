package httpapi

import (
	"net/http"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func (a *App) favIndex(w http.ResponseWriter, r *http.Request) {
	var favorites []model.FavoriteList
	if err := a.DB.Order("id").Find(&favorites).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	var subscriptions []model.Subscription
	if err := a.DB.Order("id").Find(&subscriptions).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	favIDs := make([]int64, len(favorites))
	for i, item := range favorites {
		favIDs[i] = item.ID
	}
	subIDs := make([]int64, len(subscriptions))
	for i, item := range subscriptions {
		subIDs[i] = item.ID
	}
	favCovers, err := a.coversFor(model.CoverableFavoriteList, favIDs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	subCovers, err := a.coversFor(model.CoverableSubscription, subIDs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}

	list := make([]any, 0, len(favorites)+len(subscriptions))
	for _, fav := range favorites {
		list = append(list, map[string]any{
			"id":          fav.ID,
			"type":        "favorite",
			"title":       fav.Title,
			"description": "",
			"cover":       fav.Cover,
			"media_count": fav.MediaCount,
			"created_at":  dbtime.Unix(dbtime.FromPtr(fav.CreatedAt)),
			"updated_at":  dbtime.Unix(dbtime.FromPtr(fav.UpdatedAt)),
			"ctime":       dbtime.Unix(fav.Ctime.Time),
			"mtime":       dbtime.Unix(fav.Mtime.Time),
			"cover_info":  coverJSON(favCovers[fav.ID]),
		})
	}
	for _, sub := range subscriptions {
		list = append(list, map[string]any{
			"id":          -sub.ID,
			"type":        "subscription",
			"title":       sub.Name,
			"description": sub.Description,
			"cover":       sub.Cover,
			"media_count": atoiDefault(sub.Total),
			"created_at":  dbtime.Unix(dbtime.FromPtr(sub.CreatedAt)),
			"updated_at":  dbtime.Unix(dbtime.FromPtr(sub.UpdatedAt)),
			"ctime":       dbtime.Unix(dbtime.FromPtr(sub.CreatedAt)),
			"mtime":       dbtime.Unix(dbtime.FromPtr(sub.UpdatedAt)),
			"cover_info":  coverJSON(subCovers[sub.ID]),
		})
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) favShow(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	if id < 0 {
		var sub model.Subscription
		if err := a.DB.First(&sub, -id).Error; err != nil {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                -sub.ID,
			"type":              "subscription",
			"title":             sub.Name,
			"description":       sub.Description,
			"cover":             sub.Cover,
			"media_count":       atoiDefault(sub.Total),
			"created_at":        dbtime.Unix(dbtime.FromPtr(sub.CreatedAt)),
			"updated_at":        dbtime.Unix(dbtime.FromPtr(sub.UpdatedAt)),
			"subscription_type": sub.Type,
			"mid":               sub.Mid,
			"season_id":         nil,
			"status":            sub.Status,
			"last_check_at":     dbtime.Unix(dbtime.FromPtr(sub.LastCheckAt)),
		})
		return
	}
	var fav model.FavoriteList
	if err := a.DB.First(&fav, id).Error; err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":          fav.ID,
		"type":        "favorite",
		"title":       fav.Title,
		"description": "",
		"cover":       fav.Cover,
		"media_count": fav.MediaCount,
		"created_at":  dbtime.Unix(dbtime.FromPtr(fav.CreatedAt)),
		"updated_at":  dbtime.Unix(dbtime.FromPtr(fav.UpdatedAt)),
		"fid":         nil,
		"mid":         nil,
	})
}

func (a *App) favVideos(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil || id == 0 {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	query := a.DB.Model(&model.Video{}).Where("deleted_at IS NULL")
	if id > 0 {
		query = query.Where("id IN (?)", a.DB.Model(&model.FavoriteListVideo{}).Select("video_id").Where("favorite_list_id = ?", id)).
			Order("fav_time DESC").Order("created_at DESC")
	} else {
		query = query.Where("id IN (?)", a.DB.Model(&model.SubscriptionVideo{}).Select("video_id").Where("subscription_id = ?", -id)).
			Order("pubtime DESC").Order("created_at DESC")
	}
	var videos []model.Video
	if err := query.Find(&videos).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	ids := videoIDs(videos)
	covers, err := a.coversFor(model.CoverableVideo, ids)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	list := make([]any, 0, len(videos))
	for _, video := range videos {
		list = append(list, lightweightJSON(video, covers[video.ID], dbtime.FormatLocal(dbtime.FromPtr(video.CreatedAt))))
	}
	writeJSON(w, http.StatusOK, list)
}
