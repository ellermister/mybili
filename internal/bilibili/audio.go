package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const audioHost = "https://www.bilibili.com"

// AudioSong 是音频稿件的时长。sid 与收藏里的视频 ID 相同。
type AudioSong struct {
	Duration int `json:"duration"`
}

// AudioInfo 请求音频详情。测试客户端的 BaseURL 不是正式接口时，改打到那个地址。
func AudioInfo(ctx context.Context, client *Client, sid int64) (AudioSong, error) {
	if client == nil || sid <= 0 {
		return AudioSong{}, fmt.Errorf("audio id is empty")
	}
	query := url.Values{}
	query.Set("sid", strconv.FormatInt(sid, 10))
	target := audioHost + "/audio/music-service-c/web/song/info?" + query.Encode()
	if base := strings.TrimRight(client.BaseURL, "/"); base != "" && base != APIHost {
		target = base + "/audio/music-service-c/web/song/info?" + query.Encode()
	}
	body, err := client.Get(ctx, target)
	if err != nil {
		return AudioSong{}, err
	}
	raw, ok, err := client.Parse(body)
	if err != nil {
		return AudioSong{}, err
	}
	if !ok {
		return AudioSong{}, fmt.Errorf("bilibili did not return audio data")
	}
	var song AudioSong
	if err := json.Unmarshal(raw, &song); err != nil {
		return AudioSong{}, fmt.Errorf("bilibili data: %w", err)
	}
	return song, nil
}

// AudioFile 是音频稿件的可下载文件。地址来自 cdns，size 是字节数。
type AudioFile struct {
	Sid  int64
	Size int64
	URLs []string
}

// AudioURL 取音频下载地址。privilege 和 quality 按网页固定为 2。
func AudioURL(ctx context.Context, client *Client, sid int64) (AudioFile, error) {
	if client == nil || sid <= 0 {
		return AudioFile{}, fmt.Errorf("audio id is empty")
	}
	query := url.Values{}
	query.Set("sid", strconv.FormatInt(sid, 10))
	query.Set("privilege", "2")
	query.Set("quality", "2")
	target := audioHost + "/audio/music-service-c/web/url?" + query.Encode()
	if base := strings.TrimRight(client.BaseURL, "/"); base != "" && base != APIHost {
		target = base + "/audio/music-service-c/web/url?" + query.Encode()
	}
	body, err := client.Get(ctx, target)
	if err != nil {
		return AudioFile{}, err
	}
	raw, ok, err := client.Parse(body)
	if err != nil {
		return AudioFile{}, err
	}
	if !ok {
		return AudioFile{}, fmt.Errorf("bilibili did not return audio url")
	}
	var data struct {
		Sid  int64    `json:"sid"`
		Size int64    `json:"size"`
		CDNs []string `json:"cdns"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return AudioFile{}, fmt.Errorf("bilibili data: %w", err)
	}
	file := AudioFile{Sid: data.Sid, Size: data.Size}
	for _, item := range data.CDNs {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		file.URLs = append(file.URLs, item)
	}
	return file, nil
}
