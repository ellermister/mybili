package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// VideoDetail 是 view 接口里的稿件详情和分 P。
type VideoDetail struct {
	AID     int64       `json:"aid"`
	Bvid    string      `json:"bvid"`
	Title   string      `json:"title"`
	Desc    string      `json:"desc"`
	Pic     string      `json:"pic"`
	Pubdate int64       `json:"pubdate"`
	State   int         `json:"state"`
	Owner   *VideoOwner `json:"owner"`
	Pages   []VideoPage `json:"pages"`
}

type VideoOwner struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type VideoPage struct {
	Cid       int64  `json:"cid"`
	Page      int    `json:"page"`
	From      string `json:"from"`
	Part      string `json:"part"`
	Duration  int    `json:"duration"`
	Vid       string `json:"vid"`
	Weblink   string `json:"weblink"`
	Dimension struct {
		Width  int `json:"width"`
		Height int `json:"height"`
		Rotate int `json:"rotate"`
	} `json:"dimension"`
	FirstFrame string `json:"first_frame"`
}

// VideoView 请求稿件详情。一次响应里包含分 P，不再抓视频网页。
func VideoView(ctx context.Context, client *Client, key string) (VideoDetail, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return VideoDetail{}, fmt.Errorf("video id is empty")
	}
	query := url.Values{}
	if strings.HasPrefix(strings.ToLower(key), "bv") {
		query.Set("bvid", key)
	} else {
		query.Set("aid", key)
	}
	body, err := client.Get(ctx, "/x/web-interface/view?"+query.Encode())
	if err != nil {
		return VideoDetail{}, err
	}
	raw, ok, err := client.Parse(body)
	if err != nil {
		return VideoDetail{}, err
	}
	if !ok {
		return VideoDetail{}, fmt.Errorf("bilibili did not return video data")
	}
	var detail VideoDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return VideoDetail{}, fmt.Errorf("bilibili data: %w", err)
	}
	if detail.AID == 0 {
		return VideoDetail{}, fmt.Errorf("bilibili video aid is empty")
	}
	return detail, nil
}
