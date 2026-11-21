package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const (
	favListPageSize  = 20
	favListMaxPages  = 50
	favVideoMaxPages = 200
)

// Folder 是「我创建的收藏夹」列表里的一条，不含里面的视频。
type Folder struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Ctime      int64  `json:"ctime"`
	Mtime      int64  `json:"mtime"`
	MediaCount int    `json:"media_count"`
}

// ListCreatedFolders 分页拉取当前 Cookie 对应用户创建的收藏夹。
// PageInterval 用于页与页之间停顿，避免一次手动同步连续打满接口。
func ListCreatedFolders(ctx context.Context, client *Client, cookies []Cookie, pageInterval time.Duration) ([]Folder, error) {
	mid := cookieValue(cookies, "DedeUserID")
	if mid == "" {
		return nil, fmt.Errorf("cookie does not contain DedeUserID")
	}
	var folders []Folder
	for page := 1; page <= favListMaxPages; page++ {
		if page > 1 {
			if err := wait(ctx, pageInterval); err != nil {
				return nil, err
			}
		}
		query := url.Values{}
		query.Set("pn", fmt.Sprintf("%d", page))
		query.Set("ps", fmt.Sprintf("%d", favListPageSize))
		query.Set("up_mid", mid)
		body, err := client.Get(ctx, "/x/v3/fav/folder/created/list?"+query.Encode())
		if err != nil {
			return nil, err
		}
		raw, ok, err := client.Parse(body)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("bilibili did not return folder data")
		}
		var data struct {
			List    []Folder `json:"list"`
			HasMore bool     `json:"has_more"`
		}
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, fmt.Errorf("bilibili data: %w", err)
		}
		folders = append(folders, data.List...)
		if !data.HasMore {
			return folders, nil
		}
	}
	return nil, fmt.Errorf("folder list exceeds %d pages, stopped request", favListMaxPages)
}

// FavoriteMedia 是收藏夹资源列表里的一条视频，不含分 P。
type FavoriteMedia struct {
	ID       int64  `json:"id"`
	Type     int    `json:"type"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Intro    string `json:"intro"`
	Page     int    `json:"page"`
	Duration int    `json:"duration"`
	Upper    *struct {
		Mid int64 `json:"mid"`
	} `json:"upper"`
	Attr    int    `json:"attr"`
	Link    string `json:"link"`
	Bvid    string `json:"bvid"`
	Pubtime int64  `json:"pubtime"`
	FavTime int64  `json:"fav_time"`
}

// ListFavoriteMedias 拉取一个收藏夹里的视频。pageSize 小于 1 时用 40。
// onlyPage 为 0 时翻完全部页，大于 0 时只请求那一页。
func ListFavoriteMedias(ctx context.Context, client *Client, mediaID int64, pageSize int, onlyPage int, pageInterval time.Duration) ([]FavoriteMedia, error) {
	if onlyPage < 0 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if pageSize < 1 {
		pageSize = 40
	}
	start, end := 1, favVideoMaxPages
	if onlyPage > 0 {
		if onlyPage > favVideoMaxPages {
			return nil, fmt.Errorf("page %d exceeds %d", onlyPage, favVideoMaxPages)
		}
		start, end = onlyPage, onlyPage
	}
	var medias []FavoriteMedia
	for page := start; page <= end; page++ {
		if page > start {
			if err := wait(ctx, pageInterval); err != nil {
				return nil, err
			}
		}
		query := url.Values{}
		query.Set("media_id", fmt.Sprintf("%d", mediaID))
		query.Set("pn", fmt.Sprintf("%d", page))
		query.Set("ps", fmt.Sprintf("%d", pageSize))
		query.Set("keyword", "")
		query.Set("order", "mtime")
		query.Set("type", "0")
		query.Set("tid", "0")
		query.Set("platform", "web")
		body, err := client.Get(ctx, "/x/v3/fav/resource/list?"+query.Encode())
		if err != nil {
			return nil, err
		}
		raw, ok, err := client.Parse(body)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("bilibili did not return favorite videos")
		}
		var data struct {
			Medias  []json.RawMessage `json:"medias"`
			HasMore bool              `json:"has_more"`
		}
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, fmt.Errorf("bilibili data: %w", err)
		}
		for _, item := range data.Medias {
			if len(item) == 0 || string(item) == "null" {
				continue
			}
			var media FavoriteMedia
			if err := json.Unmarshal(item, &media); err != nil {
				return nil, fmt.Errorf("bilibili media: %w", err)
			}
			if media.ID == 0 {
				continue
			}
			medias = append(medias, media)
		}
		if onlyPage > 0 || !data.HasMore {
			return medias, nil
		}
	}
	return nil, fmt.Errorf("favorite %d exceeds %d pages, stopped request", mediaID, favVideoMaxPages)
}

// FolderResource 是移动端收藏夹列表里的一条。失效稿件仍可能带着原来的标题和封面。
type FolderResource struct {
	OID       int64  `json:"oid"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	IsInvalid bool   `json:"is_invalid"`
}

// ListFolderResources 拉取收藏夹某一页。接口没有这一页时返回空列表。
func ListFolderResources(ctx context.Context, client *Client, mediaID int64, page, pageSize int) ([]FolderResource, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if pageSize < 1 {
		pageSize = 40
	}
	query := url.Values{}
	query.Set("media_id", fmt.Sprintf("%d", mediaID))
	query.Set("pn", fmt.Sprintf("%d", page))
	query.Set("ps", fmt.Sprintf("%d", pageSize))
	query.Set("build", "85900200")
	query.Set("c_locale", "en")
	query.Set("device", "phone")
	query.Set("disable_rcmd", "0")
	query.Set("mobi_app", "iphone")
	query.Set("platform", "ios")
	query.Set("s_locale", "en")
	body, err := client.Get(ctx, "/x/v3/fav/folder/resources?"+query.Encode())
	if err != nil {
		return nil, err
	}
	raw, ok, err := client.Parse(body)
	if err != nil || !ok {
		return nil, err
	}
	var data struct {
		List []FolderResource `json:"list"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("bilibili data: %w", err)
	}
	return data.List, nil
}

func cookieValue(cookies []Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
