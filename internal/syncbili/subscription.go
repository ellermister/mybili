package syncbili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/download"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/schedule"
	"mybili/internal/taskrun"
)

const (
	// TypeSubscription 是一条订阅的更新。
	TypeSubscription = "subscription"
	// TypeSyncVideo 是单个稿件的详情和分 P。
	TypeSyncVideo = "sync-video"

	subTypeUp      = "up"
	subTypeSeasons = "seasons"
	subTypeSeries  = "series"
	subPageLimit   = 200
)

type subscriptionPayload struct {
	ID      int64 `json:"id"`
	PullAll bool  `json:"pull_all"`
}

var (
	listURLPattern   = regexp.MustCompile(`/(\d+)/lists/(\d+)`)
	uploadURLPattern = regexp.MustCompile(`/(\d+)/upload`)
	spaceURLPattern  = regexp.MustCompile(`space\.bilibili\.com/(\d+)`)
)

// SubSpecs 每 10 分钟为每个启用中的订阅写一条 pending。不在这里请求 Bilibili。
func SubSpecs(db *gorm.DB) []schedule.Spec {
	return []schedule.Spec{{
		Name: "订阅",
		Due:  everyTenMinutes,
		Tasks: func(ctx context.Context, now time.Time) ([]schedule.Draft, error) {
			return subscriptionDrafts(ctx, db, now, false)
		},
	}}
}

func subscriptionDrafts(ctx context.Context, db *gorm.DB, now time.Time, pullAll bool) ([]schedule.Draft, error) {
	var rows []model.Subscription
	if err := db.WithContext(ctx).Where("status = ?", 1).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	drafts := make([]schedule.Draft, 0, len(rows))
	for _, row := range rows {
		body, err := json.Marshal(subscriptionPayload{ID: row.ID, PullAll: pullAll})
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, schedule.Draft{
			Queue:       taskrun.QueueBiliFast,
			Type:        TypeSubscription,
			UniqueKey:   fmt.Sprintf("sync:subscription:%d", row.ID),
			Payload:     string(body),
			MaxRetries:  jobs.DefaultMaxRetries,
			ReopenAfter: now.Truncate(time.Minute),
		})
	}
	return drafts, nil
}

// CreateSubscription 按链接写入或更新一条订阅，不请求稿件列表。
func CreateSubscription(ctx context.Context, db *gorm.DB, kind, rawURL string, now time.Time) (model.Subscription, error) {
	resolved, err := resolveShortURL(ctx, strings.TrimSpace(rawURL))
	if err != nil {
		return model.Subscription{}, err
	}
	mid, listID, err := parseSubscriptionURL(kind, resolved)
	if err != nil {
		return model.Subscription{}, err
	}
	var row model.Subscription
	query := db.WithContext(ctx).Where("mid = ? AND type = ?", mid, kind)
	if listID != "" {
		query = query.Where("list_id = ?", listID)
	}
	err = query.Limit(1).Find(&row).Error
	if err != nil {
		return model.Subscription{}, err
	}
	stamp := dbtime.DateTime{Time: now}
	if row.ID == 0 {
		row = model.Subscription{
			Type: kind, Mid: mid, Name: "", URL: resolved, Cover: "", CacheImage: "",
			Description: "", Total: "0", Status: 1, CreatedAt: &stamp, UpdatedAt: &stamp,
		}
		if listID != "" {
			row.ListID = &listID
		}
		if err := db.WithContext(ctx).Create(&row).Error; err != nil {
			return model.Subscription{}, err
		}
		return row, nil
	}
	updates := map[string]any{"url": resolved, "updated_at": stamp}
	if err := db.WithContext(ctx).Model(&row).Updates(updates).Error; err != nil {
		return model.Subscription{}, err
	}
	row.URL = resolved
	return row, nil
}

// UpdateSubscription 拉取一条订阅的稿件并写入关联。pullAll 为假时只拉第一页。
func (s Sync) UpdateSubscription(ctx context.Context, id int64, pullAll bool) error {
	if s.DB == nil || s.API == nil {
		return fmt.Errorf("syncer is incomplete")
	}
	now := s.now()
	if until, active, err := suspendUntil(s.DB, now); err != nil {
		return err
	} else if active {
		return &SuspendedError{Until: until}
	}
	var row model.Subscription
	if err := s.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return err
	}
	if row.Status != 1 {
		return nil
	}
	var err error
	switch row.Type {
	case subTypeSeasons, subTypeSeries:
		err = s.syncList(ctx, &row, pullAll, now)
	case subTypeUp:
		err = s.syncUp(ctx, &row, pullAll, now)
	default:
		return fmt.Errorf("未知订阅类型 %s", row.Type)
	}
	if err != nil {
		if bilibili.IsRiskControl(err) {
			return s.suspend(now)
		}
		return err
	}
	if err := s.queueSubscriptionDownloads(ctx, row.ID, now); err != nil {
		if bilibili.IsRiskControl(err) {
			return s.suspend(now)
		}
		return err
	}
	stamp := dbtime.DateTime{Time: now}
	return s.DB.WithContext(ctx).Model(&row).Updates(map[string]any{
		"last_check_at": stamp,
		"updated_at":    stamp,
	}).Error
}

// DeleteSubscription 删掉订阅和只属于它的视频关联。被收藏夹或其他订阅引用的视频留着。
// 返回删除前属于这条订阅的稿件 ID，方便调用方清理这一层目录。
func DeleteSubscription(ctx context.Context, db *gorm.DB, id int64, now time.Time) ([]int64, error) {
	var links []model.SubscriptionVideo
	if err := db.WithContext(ctx).Where("subscription_id = ?", id).Find(&links).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(links))
	stamp := dbtime.DateTime{Time: now}
	for _, link := range links {
		ids = append(ids, link.VideoID)
		var favs int64
		if err := db.WithContext(ctx).Model(&model.FavoriteListVideo{}).Where("video_id = ?", link.VideoID).Count(&favs).Error; err != nil {
			return nil, err
		}
		var others int64
		if err := db.WithContext(ctx).Model(&model.SubscriptionVideo{}).
			Where("video_id = ? AND subscription_id <> ?", link.VideoID, id).Count(&others).Error; err != nil {
			return nil, err
		}
		if favs > 0 || others > 0 {
			continue
		}
		if err := db.WithContext(ctx).Model(&model.Video{}).Where("id = ?", link.VideoID).
			Update("deleted_at", stamp).Error; err != nil {
			return nil, err
		}
	}
	if err := db.WithContext(ctx).Where("subscription_id = ?", id).Delete(&model.SubscriptionVideo{}).Error; err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).Where("coverable_type = ? AND coverable_id = ?", model.CoverableSubscription, id).
		Delete(&model.Coverable{}).Error; err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).Delete(&model.Subscription{}, id).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (s Sync) syncList(ctx context.Context, row *model.Subscription, pullAll bool, now time.Time) error {
	mid, err := strconv.ParseInt(row.Mid, 10, 64)
	if err != nil || row.ListID == nil {
		return fmt.Errorf("订阅缺少列表编号")
	}
	listID, err := strconv.ParseInt(*row.ListID, 10, 64)
	if err != nil {
		return err
	}
	if row.Type == subTypeSeries {
		meta, err := bilibili.SeriesInfo(ctx, s.API, listID)
		if err != nil {
			return err
		}
		if err := s.saveSubMeta(ctx, row, meta.Name, meta.Description, meta.Cover, meta.Total, now); err != nil {
			return err
		}
	}
	total := 0
	if row.Total != "" {
		total, _ = strconv.Atoi(row.Total)
	}
	loaded := 0
	for page := 1; page <= subPageLimit; page++ {
		var archives []bilibili.SpaceArchive
		if row.Type == subTypeSeasons {
			season, err := bilibili.SeasonsArchives(ctx, s.API, mid, listID, page)
			if err != nil {
				return err
			}
			if page == 1 {
				if err := s.saveSubMeta(ctx, row, season.Name, season.Description, season.Cover, season.Total, now); err != nil {
					return err
				}
				total = season.Total
			}
			archives = season.Archives
		} else {
			var err error
			archives, err = bilibili.SeriesArchives(ctx, s.API, mid, listID, page)
			if err != nil {
				return err
			}
			if page == 1 && len(archives) > 0 && archives[0].Cover != "" {
				if err := s.saveSubMeta(ctx, row, row.Name, row.Description, archives[0].Cover, total, now); err != nil {
					return err
				}
			}
		}
		if len(archives) == 0 {
			break
		}
		if err := s.linkArchives(ctx, row.ID, archives, false, now); err != nil {
			return err
		}
		loaded += len(archives)
		if !pullAll || (total > 0 && loaded >= total) {
			break
		}
		if err := waitInterval(ctx, s.PageInterval); err != nil {
			return err
		}
	}
	return nil
}

func (s Sync) syncUp(ctx context.Context, row *model.Subscription, pullAll bool, now time.Time) error {
	mid, err := strconv.ParseInt(row.Mid, 10, 64)
	if err != nil {
		return err
	}
	name, face, sign, err := bilibili.UpCard(ctx, s.API, mid)
	if err != nil {
		return err
	}
	if err := s.saveSubMeta(ctx, row, name, sign, face, 0, now); err != nil {
		return err
	}
	var offset int64
	loaded := 0
	for page := 1; page <= subPageLimit; page++ {
		up, err := bilibili.UpArchives(ctx, mid, offset)
		if err != nil {
			return err
		}
		if len(up.Archives) == 0 {
			break
		}
		if err := s.linkArchives(ctx, row.ID, up.Archives, true, now); err != nil {
			return err
		}
		loaded += len(up.Archives)
		if !pullAll || !up.HasNext || up.LastAid == 0 {
			break
		}
		offset = up.LastAid
		if err := waitInterval(ctx, s.PageInterval); err != nil {
			return err
		}
	}
	total := loaded
	if !pullAll {
		var count int64
		if err := s.DB.WithContext(ctx).Model(&model.SubscriptionVideo{}).Where("subscription_id = ?", row.ID).Count(&count).Error; err != nil {
			return err
		}
		total = int(count)
	}
	return s.DB.WithContext(ctx).Model(row).Update("total", strconv.Itoa(total)).Error
}

func (s Sync) saveSubMeta(ctx context.Context, row *model.Subscription, name, description, cover string, total int, now time.Time) error {
	updates := map[string]any{"updated_at": dbtime.DateTime{Time: now}}
	if name != "" {
		updates["name"] = name
		row.Name = name
	}
	if description != "" {
		updates["description"] = description
		row.Description = description
	}
	if cover != "" && cover != row.Cover {
		updates["cover"] = cover
		row.Cover = cover
	}
	if total > 0 {
		updates["total"] = strconv.Itoa(total)
		row.Total = strconv.Itoa(total)
	}
	if err := s.DB.WithContext(ctx).Model(row).Updates(updates).Error; err != nil {
		return err
	}
	if cover != "" {
		_, err := s.EnqueueCover(nil, CoverKindSubscription, row.ID, cover)
		return err
	}
	return nil
}

func (s Sync) linkArchives(ctx context.Context, subID int64, archives []bilibili.SpaceArchive, fillVideo bool, now time.Time) error {
	stamp := dbtime.DateTime{Time: now}
	for _, archive := range archives {
		if archive.Aid <= 0 {
			continue
		}
		var link model.SubscriptionVideo
		err := s.DB.WithContext(ctx).Where("subscription_id = ? AND video_id = ?", subID, archive.Aid).Limit(1).Find(&link).Error
		if err != nil {
			return err
		}
		if link.VideoID == 0 {
			link = model.SubscriptionVideo{
				SubscriptionID: subID, VideoID: archive.Aid, Bvid: archive.Bvid,
				CreatedAt: &stamp, UpdatedAt: &stamp,
			}
			if err := s.DB.WithContext(ctx).Create(&link).Error; err != nil {
				return err
			}
			if s.AfterLinked != nil {
				s.AfterLinked(ctx, []int64{archive.Aid})
			}
		} else if link.Bvid != archive.Bvid && archive.Bvid != "" {
			if err := s.DB.WithContext(ctx).Model(&link).Updates(map[string]any{"bvid": archive.Bvid, "updated_at": stamp}).Error; err != nil {
				return err
			}
		}
		missing, err := s.videoMissing(ctx, archive.Aid)
		if err != nil {
			return err
		}
		if missing && s.QueueVideo != nil && archive.Bvid != "" {
			if err := s.QueueVideo(ctx, archive.Bvid); err != nil {
				return err
			}
		}
		if _, err := s.ensureVideo(ctx, archive, fillVideo, stamp); err != nil {
			return err
		}
	}
	return nil
}

func (s Sync) videoMissing(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := s.DB.WithContext(ctx).Model(&model.Video{}).Where("id = ?", id).Count(&count).Error
	return count == 0, err
}

func (s Sync) ensureVideo(ctx context.Context, archive bilibili.SpaceArchive, fill bool, stamp dbtime.DateTime) (bool, error) {
	var video model.Video
	err := s.DB.WithContext(ctx).Where("id = ?", archive.Aid).Limit(1).Find(&video).Error
	if err != nil {
		return false, err
	}
	if video.ID != 0 {
		if video.DeletedAt != nil {
			if err := s.DB.WithContext(ctx).Model(&video).Update("deleted_at", nil).Error; err != nil {
				return false, err
			}
		}
		return false, nil
	}
	if !fill {
		return true, nil
	}
	pub := stamp
	if archive.Pubtime > 0 {
		pub = dbtime.DateTime{Time: time.Unix(archive.Pubtime, 0).In(dbtime.Location())}
	}
	pages := archive.Pages
	if pages < 1 {
		pages = 1
	}
	row := model.Video{
		ID: archive.Aid, Link: "https://www.bilibili.com/video/" + archive.Bvid + "/",
		Title: archive.Title, Intro: "", Cover: archive.Cover, Bvid: archive.Bvid,
		Pubtime: pub, Duration: archive.Duration, CacheImage: "", Page: pages, Type: 2,
		CreatedAt: &stamp, UpdatedAt: &stamp,
	}
	if err := s.DB.WithContext(ctx).Create(&row).Error; err != nil {
		return false, err
	}
	return true, nil
}

// queuePartDownloads 给这个稿件里还没下的分 P 写下载任务。没有数据目录时跳过。
func (s Sync) queuePartDownloads(ctx context.Context, videoID int64, now time.Time) error {
	if s.DataDir == "" || videoID <= 0 {
		return nil
	}
	_, err := download.EnqueueParts(ctx, s.DB, s.DataDir, videoID, now, s.Wake)
	return err
}

// queueDanmaku 给这个稿件里还没下过、或超过 7 天的分 P 写弹幕任务。
func (s Sync) queueDanmaku(ctx context.Context, videoID int64, now time.Time) error {
	if s.DataDir == "" || videoID <= 0 {
		return nil
	}
	_, err := download.EnqueueDanmaku(ctx, s.DB, videoID, now, s.Wake)
	return err
}

// queueSubscriptionDownloads 给这条订阅里已经有分 P 的稿件补下载任务。
// 只看这一条订阅，不扫全库。
// 本机文件不在、且这次会新开下载任务时，先请求稿件详情。下架则标无效，不再写下载和弹幕。
// 下载任务已经在队列里或已经失败时，不再请求，也不补弹幕。文件已在的稿件只按原规则补弹幕。
func (s Sync) queueSubscriptionDownloads(ctx context.Context, subID int64, now time.Time) error {
	if s.DataDir == "" {
		return nil
	}
	var ids []int64
	if err := s.DB.WithContext(ctx).Model(&model.SubscriptionVideo{}).Where("subscription_id = ?", subID).Pluck("video_id", &ids).Error; err != nil {
		return err
	}
	viewed := false
	for _, id := range ids {
		missing, newTask, err := download.VideoDownloadGap(ctx, s.DB, s.DataDir, id, now)
		if err != nil {
			return err
		}
		if newTask {
			if viewed {
				if err := waitInterval(ctx, s.PageInterval); err != nil {
					return err
				}
			}
			viewed = true
			if err := s.syncVideo(ctx, id, "", now); err != nil {
				return err
			}
			continue
		}
		if missing {
			continue
		}
		if err := s.queueDanmaku(ctx, id, now); err != nil {
			return err
		}
	}
	return nil
}

func parseSubscriptionURL(kind, raw string) (string, string, error) {
	switch kind {
	case subTypeSeasons, subTypeSeries:
		match := listURLPattern.FindStringSubmatch(raw)
		if match == nil {
			return "", "", fmt.Errorf("链接里没有合集或系列编号")
		}
		return match[1], match[2], nil
	case subTypeUp:
		if match := uploadURLPattern.FindStringSubmatch(raw); match != nil {
			return match[1], "", nil
		}
		if match := spaceURLPattern.FindStringSubmatch(raw); match != nil {
			return match[1], "", nil
		}
		return "", "", fmt.Errorf("链接里没有 UP 主编号")
	default:
		return "", "", fmt.Errorf("订阅类型只能是 up、seasons、series")
	}
}

func resolveShortURL(ctx context.Context, raw string) (string, error) {
	if !strings.Contains(raw, "b23.tv") {
		return raw, nil
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	if resp.Request == nil || resp.Request.URL == nil {
		return raw, nil
	}
	return resp.Request.URL.String(), nil
}

// EnqueueSubscription 把一条订阅写成 pending，并在需要时通知 Feeder。
func EnqueueSubscription(ctx context.Context, db *gorm.DB, id int64, pullAll bool, now time.Time, wake func(context.Context, string) error) error {
	body, err := json.Marshal(subscriptionPayload{ID: id, PullAll: pullAll})
	if err != nil {
		return err
	}
	_, err = jobs.Store{DB: db}.Enqueue(ctx, jobs.Draft{
		Queue:      taskrun.QueueBiliFast,
		Type:       TypeSubscription,
		UniqueKey:  fmt.Sprintf("sync:subscription:%d", id),
		Payload:    string(body),
		MaxRetries: jobs.DefaultMaxRetries,
	}, now, jobs.SaveOptions{}, wake)
	return err
}

// EnqueueVideoDetail 把单个稿件详情写成 pending。已经跑完的不重新打开。
func EnqueueVideoDetail(ctx context.Context, db *gorm.DB, bvid string, now time.Time, wake func(context.Context, string) error) error {
	bvid = strings.TrimSpace(bvid)
	if bvid == "" {
		return nil
	}
	body, err := json.Marshal(map[string]string{"bvid": bvid})
	if err != nil {
		return err
	}
	_, err = jobs.Store{DB: db}.Enqueue(ctx, jobs.Draft{
		Queue:      taskrun.QueueBiliFast,
		Type:       TypeSyncVideo,
		UniqueKey:  "sync:video:" + bvid,
		Payload:    string(body),
		MaxRetries: jobs.DefaultMaxRetries,
	}, now, jobs.SaveOptions{ReopenAfter: time.Unix(1, 0)}, wake)
	return err
}

func subscriptionTask(payload string) (subscriptionPayload, error) {
	var body subscriptionPayload
	if err := json.Unmarshal([]byte(payload), &body); err != nil || body.ID <= 0 {
		return body, errors.New("订阅任务不完整")
	}
	return body, nil
}
