package httpapi

import (
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func coverJSON(cover *model.Cover) any {
	if cover == nil {
		return nil
	}
	return map[string]any{
		"id":         cover.ID,
		"type":       cover.Type,
		"url":        cover.URL,
		"filename":   cover.Filename,
		"path":       cover.Path,
		"mime_type":  cover.MimeType,
		"size":       cover.Size,
		"width":      cover.Width,
		"height":     cover.Height,
		"created_at": dbtime.FormatISO(dbtime.FromPtr(cover.CreatedAt)),
		"image_url":  publicURL(cover.Path),
	}
}

func videoJSON(video model.Video, cover *model.Cover) map[string]any {
	item := map[string]any{
		"id":                   video.ID,
		"link":                 video.Link,
		"title":                video.Title,
		"intro":                video.Intro,
		"cover":                video.Cover,
		"bvid":                 video.Bvid,
		"pubtime":              dbtime.Unix(video.Pubtime.Time),
		"duration":             video.Duration,
		"attr":                 video.Attr,
		"invalid":              video.Invalid,
		"frozen":               video.Frozen,
		"cache_image":          video.CacheImage,
		"page":                 video.Page,
		"fav_time":             dbtime.Unix(dbtime.FromPtr(video.FavTime)),
		"created_at":           dbtime.FormatISO(dbtime.FromPtr(video.CreatedAt)),
		"updated_at":           dbtime.FormatISO(dbtime.FromPtr(video.UpdatedAt)),
		"video_downloaded_at":  dbtime.FormatLocal(dbtime.FromPtr(video.VideoDownloadedAt)),
		"video_downloaded_num": video.VideoDownloadedNum,
		"audio_downloaded_num": video.AudioDownloadedNum,
		"deleted_at":           dbtime.FormatISO(dbtime.FromPtr(video.DeletedAt)),
		"upper_id":             video.UpperID,
		"type":                 video.Type,
		"cover_info":           coverJSON(cover),
	}
	return item
}

func favoriteJSON(fav model.FavoriteList, cover *model.Cover) map[string]any {
	return map[string]any{
		"id":          fav.ID,
		"title":       fav.Title,
		"cover":       fav.Cover,
		"ctime":       dbtime.Unix(fav.Ctime.Time),
		"mtime":       dbtime.Unix(fav.Mtime.Time),
		"media_count": fav.MediaCount,
		"cache_image": fav.CacheImage,
		"created_at":  dbtime.FormatISO(dbtime.FromPtr(fav.CreatedAt)),
		"updated_at":  dbtime.FormatISO(dbtime.FromPtr(fav.UpdatedAt)),
		"cover_info":  coverJSON(cover),
	}
}

func subscriptionJSON(sub model.Subscription, cover *model.Cover) map[string]any {
	return map[string]any{
		"id":            sub.ID,
		"type":          sub.Type,
		"mid":           sub.Mid,
		"list_id":       sub.ListID,
		"name":          sub.Name,
		"url":           sub.URL,
		"cover":         sub.Cover,
		"cache_image":   sub.CacheImage,
		"description":   sub.Description,
		"total":         sub.Total,
		"status":        sub.Status,
		"last_check_at": dbtime.FormatMinute(dbtime.FromPtr(sub.LastCheckAt)),
		"created_at":    dbtime.FormatISO(dbtime.FromPtr(sub.CreatedAt)),
		"updated_at":    dbtime.FormatISO(dbtime.FromPtr(sub.UpdatedAt)),
		"cover_info":    coverJSON(cover),
	}
}

func upperJSON(upper model.Upper, cover *model.Cover) map[string]any {
	return map[string]any{
		"id":         upper.ID,
		"mid":        upper.Mid,
		"name":       upper.Name,
		"face":       upper.Face,
		"created_at": dbtime.FormatISO(dbtime.FromPtr(upper.CreatedAt)),
		"updated_at": dbtime.FormatISO(dbtime.FromPtr(upper.UpdatedAt)),
		"cover_info": coverJSON(cover),
	}
}

func lightweightJSON(video model.Video, cover *model.Cover, createdAt any) map[string]any {
	var image any
	if cover != nil {
		image = publicURL(cover.Path)
	}
	return map[string]any{
		"id":                   video.ID,
		"title":                video.Title,
		"bvid":                 video.Bvid,
		"pubtime":              dbtime.Unix(video.Pubtime.Time),
		"fav_time":             dbtime.Unix(dbtime.FromPtr(video.FavTime)),
		"page":                 video.Page,
		"video_downloaded_num": video.VideoDownloadedNum,
		"audio_downloaded_num": video.AudioDownloadedNum,
		"frozen":               video.Frozen,
		"invalid":              video.Invalid,
		"cover":                video.Cover,
		"cover_image_url":      image,
		"created_at":           createdAt,
	}
}
