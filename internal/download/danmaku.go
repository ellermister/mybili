package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

const danmakuSegment = 6 * time.Minute

// Danmaku 按每 6 分钟一段拉取弹幕，进度写到任务输出里，全部到齐后再入库。
func (r Runner) Danmaku(ctx context.Context, taskID int64, payload partPayload) error {
	if r.DB == nil {
		return fmt.Errorf("downloader is incomplete")
	}
	var part model.VideoPart
	if err := r.DB.WithContext(ctx).First(&part, payload.PartID).Error; err != nil {
		return err
	}
	var video model.Video
	if err := r.DB.WithContext(ctx).First(&video, part.VideoID).Error; err != nil {
		return err
	}
	content, err := readSettingString(r.DB, "cookies_content")
	if err != nil {
		return err
	}
	cookies, err := bilibili.ParseNetscape(content)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	imgKey, subKey, err := wbiKeys(ctx, client, bilibili.Header(cookies))
	if err != nil {
		if bilibili.IsCookieExpired(err) && r.OnCookie != nil {
			r.OnCookie(ctx)
		}
		return err
	}
	total := segmentCount(part.Duration)
	var elems []danmakuElem
	for i := 1; i <= total; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		r.progress(ctx, taskID, i-1, total, fmt.Sprintf("弹幕 %d/%d", i-1, total))
		r.log(ctx, taskID, fmt.Sprintf("拉取弹幕分段 %d/%d", i, total))
		query := signWBI(map[string]string{
			"type":          "1",
			"oid":           strconv.FormatInt(part.Cid, 10),
			"segment_index": strconv.Itoa(i),
			"web_location":  "1315873",
			"wts":           strconv.FormatInt(time.Now().Unix(), 10),
		}, imgKey, subKey)
		body, err := getBytes(ctx, client, bilibili.APIHost+"/x/v2/dm/wbi/web/seg.so?"+query, bilibili.Header(cookies))
		if err != nil {
			return err
		}
		chunk := parseDanmakuSeg(body)
		r.log(ctx, taskID, fmt.Sprintf("分段 %d/%d 得到 %d 条", i, total, len(chunk)))
		elems = append(elems, chunk...)
		r.progress(ctx, taskID, i, total, fmt.Sprintf("弹幕 %d/%d", i, total))
	}
	if err := saveDanmaku(ctx, r.DB, video.ID, part.Cid, elems); err != nil {
		return err
	}
	now := dbtime.DateTime{Time: dbtime.Now()}
	r.log(ctx, taskID, fmt.Sprintf("弹幕已写入，共 %d 条", len(elems)))
	r.progress(ctx, taskID, total, total, "弹幕完成")
	return r.DB.WithContext(ctx).Model(&model.VideoPart{}).Where("id = ?", part.ID).Updates(map[string]any{
		"danmaku_downloaded_at": now,
		"updated_at":            now,
	}).Error
}

func segmentCount(duration int) int {
	if duration <= 0 {
		return 1
	}
	return int(math.Ceil(float64(duration) / danmakuSegment.Seconds()))
}

func wbiKeys(ctx context.Context, client *http.Client, cookieHeader string) (string, string, error) {
	body, err := getBytes(ctx, client, bilibili.APIHost+"/x/web-interface/nav", cookieHeader)
	if err != nil {
		return "", "", err
	}
	data, ok, err := bilibili.APIData(body)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", fmt.Errorf("nav 没有返回 wbi")
	}
	var nav struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	}
	if err := json.Unmarshal(data, &nav); err != nil {
		return "", "", err
	}
	img := imageKey(nav.WbiImg.ImgURL)
	sub := imageKey(nav.WbiImg.SubURL)
	if img == "" || sub == "" {
		return "", "", fmt.Errorf("nav 没有返回 wbi")
	}
	return img, sub, nil
}

func getBytes(ctx context.Context, client *http.Client, rawURL, cookieHeader string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > 180 {
			snippet = snippet[:180]
		}
		return nil, &bilibili.HTTPError{Status: resp.StatusCode, Snippet: snippet}
	}
	return body, nil
}

func saveDanmaku(ctx context.Context, db *gorm.DB, videoID, cid int64, elems []danmakuElem) error {
	var existing []int64
	if err := db.WithContext(ctx).Model(&model.Danmaku{}).Where("cid = ?", cid).Pluck("id", &existing).Error; err != nil {
		return err
	}
	seen := make(map[int64]struct{}, len(existing))
	for _, id := range existing {
		seen[id] = struct{}{}
	}
	now := dbtime.DateTime{Time: dbtime.Now()}
	rows := make([]model.Danmaku, 0, len(elems))
	for _, elem := range elems {
		if elem.ID <= 0 || elem.Content == "" {
			continue
		}
		if _, ok := seen[elem.ID]; ok {
			continue
		}
		seen[elem.ID] = struct{}{}
		created := now
		if elem.Ctime > 0 {
			created = dbtime.DateTime{Time: time.Unix(elem.Ctime, 0).In(dbtime.Location())}
		}
		row := model.Danmaku{
			ID:        elem.ID,
			VideoID:   videoID,
			Cid:       cid,
			Progress:  elem.Progress,
			Mode:      elem.Mode,
			Content:   elem.Content,
			CreatedAt: created,
		}
		if elem.hasColor {
			color := elem.Color
			row.Color = &color
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil
	}
	return db.WithContext(ctx).CreateInBatches(rows, 500).Error
}
