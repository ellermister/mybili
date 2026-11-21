package bilibili

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	appAPIHost = "https://app.biliapi.com"
	appKey     = "1d8b6e7d45233436"
	appSecret  = "560c52ccd288fed045859ed18bffd973"
)

// SpaceArchive 是合集、系列或 UP 主投稿里的一条稿件。
type SpaceArchive struct {
	Aid      int64
	Bvid     string
	Title    string
	Cover    string
	Duration int
	Pages    int
	Pubtime  int64
}

// SeasonPage 是合集的一页。
type SeasonPage struct {
	Name        string
	Description string
	Cover       string
	Total       int
	Archives    []SpaceArchive
}

// SeriesMeta 是系列的标题信息。
type SeriesMeta struct {
	Name        string
	Description string
	Cover       string
	Total       int
}

// UpPage 是 UP 主投稿的一页。LastAid 用于向后翻。
type UpPage struct {
	Archives []SpaceArchive
	HasNext  bool
	LastAid  int64
}

// SeasonsArchives 拉取合集的一页稿件。
func SeasonsArchives(ctx context.Context, client *Client, mid, seasonID int64, page int) (SeasonPage, error) {
	path := fmt.Sprintf("/x/polymer/web-space/seasons_archives_list?mid=%d&season_id=%d&sort_reverse=false&page_size=30&page_num=%d", mid, seasonID, page)
	body, err := client.Get(ctx, path)
	if err != nil {
		return SeasonPage{}, err
	}
	data, ok, err := client.Parse(body)
	if err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("seasons list is empty")
		}
		return SeasonPage{}, err
	}
	var payload struct {
		Meta struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Cover       string `json:"cover"`
			Total       int    `json:"total"`
		} `json:"meta"`
		Archives []spaceArchiveJSON `json:"archives"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return SeasonPage{}, err
	}
	return SeasonPage{
		Name: payload.Meta.Name, Description: payload.Meta.Description, Cover: payload.Meta.Cover,
		Total: payload.Meta.Total, Archives: mapArchives(payload.Archives),
	}, nil
}

// SeriesInfo 拉取系列标题。
func SeriesInfo(ctx context.Context, client *Client, seriesID int64) (SeriesMeta, error) {
	body, err := client.Get(ctx, "/x/series/series?series_id="+strconv.FormatInt(seriesID, 10))
	if err != nil {
		return SeriesMeta{}, err
	}
	data, ok, err := client.Parse(body)
	if err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("series meta is empty")
		}
		return SeriesMeta{}, err
	}
	var payload struct {
		Meta struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Cover       string `json:"cover"`
			Total       int    `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return SeriesMeta{}, err
	}
	return SeriesMeta{
		Name: payload.Meta.Name, Description: payload.Meta.Description,
		Cover: payload.Meta.Cover, Total: payload.Meta.Total,
	}, nil
}

// SeriesArchives 拉取系列的一页稿件。
func SeriesArchives(ctx context.Context, client *Client, mid, seriesID int64, page int) ([]SpaceArchive, error) {
	path := fmt.Sprintf("/x/series/archives?mid=%d&series_id=%d&only_normal=true&sort=desc&ps=30&pn=%d", mid, seriesID, page)
	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	data, ok, err := client.Parse(body)
	if err != nil || !ok {
		if err == nil {
			return nil, nil
		}
		return nil, err
	}
	var payload struct {
		Archives []spaceArchiveJSON `json:"archives"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return mapArchives(payload.Archives), nil
}

// UpCard 拉取 UP 主昵称和头像。
func UpCard(ctx context.Context, client *Client, mid int64) (name, face, sign string, err error) {
	body, err := client.Get(ctx, "/x/web-interface/card?mid="+strconv.FormatInt(mid, 10))
	if err != nil {
		return "", "", "", err
	}
	data, ok, err := client.Parse(body)
	if err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("up card is empty")
		}
		return "", "", "", err
	}
	var payload struct {
		Card struct {
			Name string `json:"name"`
			Face string `json:"face"`
			Sign string `json:"sign"`
		} `json:"card"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", "", "", err
	}
	return payload.Card.Name, payload.Card.Face, payload.Card.Sign, nil
}

// UpArchives 用移动端投稿接口拉一页。offsetAid 为 0 时从最新开始。
func UpArchives(ctx context.Context, mid, offsetAid int64) (UpPage, error) {
	params := url.Values{}
	params.Set("vmid", strconv.FormatInt(mid, 10))
	params.Set("order", "pubdate")
	params.Set("ps", "20")
	params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	if offsetAid > 0 {
		params.Set("aid", strconv.FormatInt(offsetAid, 10))
	}
	params.Set("appkey", appKey)
	encoded := params.Encode()
	sum := md5.Sum([]byte(encoded + appSecret))
	rawURL := appAPIHost + "/x/v2/space/archive/cursor?" + encoded + "&sign=" + hex.EncodeToString(sum[:])
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return UpPage{}, err
	}
	req.Header.Set("User-Agent", "bili-universal/77100100 CFNetwork/1404.0.5 Darwin/22.3.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UpPage{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return UpPage{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return UpPage{}, &HTTPError{Status: resp.StatusCode, Snippet: snippet(body)}
	}
	data, ok, err := APIData(body)
	if err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("up videos are empty")
		}
		return UpPage{}, err
	}
	var payload struct {
		HasNext bool `json:"has_next"`
		Item    []struct {
			Param    string `json:"param"`
			Bvid     string `json:"bvid"`
			Title    string `json:"title"`
			Cover    string `json:"cover"`
			Duration int    `json:"duration"`
			Videos   int    `json:"videos"`
			Ctime    int64  `json:"ctime"`
		} `json:"item"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return UpPage{}, err
	}
	page := UpPage{HasNext: payload.HasNext}
	for _, item := range payload.Item {
		aid, _ := strconv.ParseInt(item.Param, 10, 64)
		page.Archives = append(page.Archives, SpaceArchive{
			Aid: aid, Bvid: item.Bvid, Title: item.Title, Cover: item.Cover,
			Duration: item.Duration, Pages: item.Videos, Pubtime: item.Ctime,
		})
		page.LastAid = aid
	}
	return page, nil
}

type spaceArchiveJSON struct {
	Aid      int64  `json:"aid"`
	Bvid     string `json:"bvid"`
	Title    string `json:"title"`
	Pic      string `json:"pic"`
	Duration int    `json:"duration"`
	Pubdate  int64  `json:"pubdate"`
}

func mapArchives(rows []spaceArchiveJSON) []SpaceArchive {
	out := make([]SpaceArchive, 0, len(rows))
	for _, row := range rows {
		out = append(out, SpaceArchive{
			Aid: row.Aid, Bvid: row.Bvid, Title: row.Title, Cover: row.Pic,
			Duration: row.Duration, Pages: 1, Pubtime: row.Pubdate,
		})
	}
	return out
}
