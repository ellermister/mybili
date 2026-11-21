package syncbili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/schedule"
	"mybili/internal/taskrun"
)

// Video 同步一个稿件的详情和分 P。videoID 与 bvid 只传一个。
// 封面或头像有变化时写入封面队列。音频稿件不请求接口。
func (s Sync) Video(ctx context.Context, videoID int64, bvid string) error {
	if videoID <= 0 && strings.TrimSpace(bvid) == "" {
		return fmt.Errorf("video id is empty")
	}
	if videoID > 0 && strings.TrimSpace(bvid) != "" {
		return fmt.Errorf("pass only one of id or bvid")
	}
	now := s.now()
	if until, active, err := suspendUntil(s.DB, now); err != nil {
		return err
	} else if active {
		return &SuspendedError{Until: until}
	}
	return s.syncVideo(ctx, videoID, bvid, now)
}

// MissingParts 同步还没有分 P、且不是无效或音频的视频。limit 是本次最多处理多少个。
func (s Sync) MissingParts(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		return 0, fmt.Errorf("limit must be >= 1")
	}
	now := s.now()
	if until, active, err := suspendUntil(s.DB, now); err != nil {
		return 0, err
	} else if active {
		return 0, &SuspendedError{Until: until}
	}
	ids, err := s.videosWithoutParts(limit)
	if err != nil {
		return 0, err
	}
	var done int
	for i, id := range ids {
		if i > 0 {
			if err := waitInterval(ctx, s.PageInterval); err != nil {
				return done, err
			}
		}
		if err := s.syncVideo(ctx, id, "", now); err != nil {
			if bilibili.IsRiskControl(err) {
				return done, s.suspend(now)
			}
			return done, fmt.Errorf("video %d: %w", id, err)
		}
		done++
	}
	return done, nil
}

// videosWithoutParts 找出还没有分 P、且不是无效或音频的视频 ID。
func (s Sync) videosWithoutParts(limit int) ([]int64, error) {
	var ids []int64
	err := s.DB.Model(&model.Video{}).
		Where("deleted_at IS NULL").
		Where("invalid = 0").
		Where("type <> ?", 12).
		Where("NOT EXISTS (SELECT 1 FROM video_parts WHERE video_parts.video_id = videos.id)").
		Order("id").
		Limit(limit).
		Pluck("id", &ids).Error
	return ids, err
}

// MissingPartSpecs 每小时把还没有分 P 的有效视频写成详情任务。
// 不请求 Bilibili。24 小时内已经跑过的不重新打开。收藏夹全部被排除时跳过。
func MissingPartSpecs(db *gorm.DB) []schedule.Spec {
	return []schedule.Spec{{
		Name: "补拉分P",
		Due: func(now time.Time) bool {
			return now.Minute() == 0
		},
		Tasks: func(ctx context.Context, now time.Time) ([]schedule.Draft, error) {
			return missingPartDrafts(ctx, db, now)
		},
	}}
}

type missingPartCandidate struct {
	ID   int64  `gorm:"column:id"`
	Bvid string `gorm:"column:bvid"`
}

// missingPartDrafts 选出还没有分 P、未下载、未冻结、未失效的视频。
func missingPartDrafts(ctx context.Context, db *gorm.DB, now time.Time) ([]schedule.Draft, error) {
	var rows []missingPartCandidate
	err := db.WithContext(ctx).Model(&model.Video{}).
		Select("videos.id, videos.bvid").
		Where("videos.deleted_at IS NULL").
		Where("videos.invalid = 0 AND videos.frozen = 0 AND videos.video_downloaded_num = 0").
		Where("videos.type <> ?", 12).
		Where("videos.bvid <> ''").
		Where("NOT EXISTS (SELECT 1 FROM video_parts WHERE video_parts.video_id = videos.id)").
		Order("videos.id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	audio, err := missingAudioDrafts(ctx, db, now)
	if err != nil {
		return nil, err
	}
	excluded, err := excludedFavoriteIDs(db)
	if err != nil {
		return nil, err
	}
	var favs map[int64][]int64
	if len(excluded) > 0 && len(rows) > 0 {
		favs, err = favoriteIDsByVideo(ctx, db, rows)
		if err != nil {
			return nil, err
		}
	}
	reopen := now.Add(-24 * time.Hour)
	drafts := make([]schedule.Draft, 0, len(rows))
	for _, row := range rows {
		bvid := strings.TrimSpace(row.Bvid)
		if bvid == "" || allFavoritesExcluded(favs[row.ID], excluded) {
			continue
		}
		body, err := json.Marshal(map[string]string{"bvid": bvid})
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, schedule.Draft{
			Queue:       taskrun.QueueBiliFast,
			Type:        TypeSyncVideo,
			UniqueKey:   "sync:video:" + bvid,
			Payload:     string(body),
			MaxRetries:  jobs.DefaultMaxRetries,
			ReopenAfter: reopen,
		})
	}
	return append(drafts, audio...), nil
}

// missingAudioDrafts 选出还没下载过的收藏夹音频。24 小时内跑过的不重新打开。
func missingAudioDrafts(ctx context.Context, db *gorm.DB, now time.Time) ([]schedule.Draft, error) {
	var rows []missingPartCandidate
	err := db.WithContext(ctx).Model(&model.Video{}).
		Select("videos.id").
		Where("videos.deleted_at IS NULL").
		Where("videos.type = ?", 12).
		Where("videos.invalid = 0 AND videos.frozen = 0").
		Where("videos.video_downloaded_num = 0 AND videos.audio_downloaded_num = 0").
		Order("videos.id").
		Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	excluded, err := excludedFavoriteIDs(db)
	if err != nil {
		return nil, err
	}
	var favs map[int64][]int64
	if len(excluded) > 0 {
		favs, err = favoriteIDsByVideo(ctx, db, rows)
		if err != nil {
			return nil, err
		}
	}
	reopen := now.Add(-24 * time.Hour)
	drafts := make([]schedule.Draft, 0, len(rows))
	for _, row := range rows {
		if allFavoritesExcluded(favs[row.ID], excluded) {
			continue
		}
		body, err := json.Marshal(map[string]int64{"id": row.ID})
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, schedule.Draft{
			Queue:       taskrun.QueueBiliFast,
			Type:        TypeSyncAudio,
			UniqueKey:   fmt.Sprintf("sync:audio:%d", row.ID),
			Payload:     string(body),
			MaxRetries:  jobs.DefaultMaxRetries,
			ReopenAfter: reopen,
		})
	}
	return drafts, nil
}

// excludedFavoriteIDs 读取收藏夹排除。开关关着或内容解析失败时返回空，表示都不排除。
func excludedFavoriteIDs(db *gorm.DB) (map[int64]bool, error) {
	text, err := readString(db, "fav_exclude")
	if err != nil || strings.TrimSpace(text) == "" {
		return nil, err
	}
	var body struct {
		Enabled  bool    `json:"enabled"`
		Selected []int64 `json:"selected"`
	}
	if err := json.Unmarshal([]byte(text), &body); err != nil || !body.Enabled {
		return map[int64]bool{}, nil
	}
	out := make(map[int64]bool, len(body.Selected))
	for _, id := range body.Selected {
		out[id] = true
	}
	return out, nil
}

type videoFavoriteLink struct {
	VideoID int64 `gorm:"column:video_id"`
	FavID   int64 `gorm:"column:favorite_list_id"`
}

// favoriteIDsByVideo 取出这些视频各自属于哪些收藏夹。
func favoriteIDsByVideo(ctx context.Context, db *gorm.DB, rows []missingPartCandidate) (map[int64][]int64, error) {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	var links []videoFavoriteLink
	err := db.WithContext(ctx).Table("favorite_list_videos").
		Select("video_id, favorite_list_id").
		Where("video_id IN ?", ids).
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64][]int64, len(links))
	for _, link := range links {
		out[link.VideoID] = append(out[link.VideoID], link.FavID)
	}
	return out, nil
}

// allFavoritesExcluded 在视频至少属于一个收藏夹、且这些收藏夹都被排除时为 true。
func allFavoritesExcluded(favs []int64, excluded map[int64]bool) bool {
	if len(favs) == 0 || len(excluded) == 0 {
		return false
	}
	for _, id := range favs {
		if !excluded[id] {
			return false
		}
	}
	return true
}

// syncVideo 请求一个稿件的 view 并写详情。已删除的跳过，音频稿件不请求。
func (s Sync) syncVideo(ctx context.Context, videoID int64, bvid string, now time.Time) error {
	local, err := s.findVideo(videoID, bvid)
	if err != nil {
		return err
	}
	if local != nil && local.DeletedAt != nil && !local.DeletedAt.Time.IsZero() {
		return nil
	}
	if local != nil && local.IsAudio() {
		return fmt.Errorf("video %d is audio, skipped", local.ID)
	}
	key := strings.TrimSpace(bvid)
	if videoID > 0 {
		key = strconv.FormatInt(videoID, 10)
	}
	detail, err := bilibili.VideoView(ctx, s.API, key)
	if err != nil {
		if bilibili.IsVideoUnavailable(err) {
			return s.markUnavailable(ctx, local, now, err)
		}
		return err
	}
	return s.saveVideoDetail(ctx, local, detail, now)
}

// findVideo 按稿件 ID 或 bvid 找本地视频。没有记录时返回 nil。
func (s Sync) findVideo(videoID int64, bvid string) (*model.Video, error) {
	var row model.Video
	var err error
	if videoID > 0 {
		err = s.DB.First(&row, videoID).Error
	} else {
		err = s.DB.Where("bvid = ?", bvid).First(&row).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// markUnavailable 把本地稿件标为无效。已经下载过的同时冻结。本地没有行时原样返回错误。
func (s Sync) markUnavailable(ctx context.Context, local *model.Video, now time.Time, cause error) error {
	if local == nil {
		return cause
	}
	if local.Invalid != 0 {
		return nil
	}
	wasFrozen := local.Frozen != 0
	local.Invalid = 1
	if local.VideoDownloadedNum > 0 || local.AudioDownloadedNum > 0 {
		local.Frozen = 1
	} else {
		local.Frozen = 0
	}
	updated := dbtime.DateTime{Time: now}
	local.UpdatedAt = &updated
	if err := s.DB.Save(local).Error; err != nil {
		return err
	}
	if local.Frozen != 0 && !wasFrozen {
		s.noteFrozen(ctx, local.ID)
	}
	return nil
}

// saveVideoDetail 写入稿件详情、UP 主和分 P，并在需要时写入封面任务。
func (s Sync) saveVideoDetail(ctx context.Context, local *model.Video, detail bilibili.VideoDetail, now time.Time) error {
	wasFrozen := local != nil && local.Frozen != 0
	row := model.Video{ID: detail.AID, CacheImage: "", Type: 2}
	if local != nil {
		row = *local
		row.ID = detail.AID
	} else {
		created := dbtime.DateTime{Time: now}
		row.CreatedAt = &created
	}
	row.Link = fmt.Sprintf("bilibili://video/%d", detail.AID)
	row.Title = detail.Title
	row.Intro = detail.Desc
	row.Cover = detail.Pic
	row.Bvid = detail.Bvid
	row.Pubtime = unixDateTime(detail.Pubdate)
	row.Page = len(detail.Pages)
	invalid := detail.State != 0
	row.Invalid = 0
	row.Frozen = 0
	if invalid {
		row.Invalid = 1
		if row.VideoDownloadedNum > 0 || row.AudioDownloadedNum > 0 {
			row.Frozen = 1
		}
	}
	if detail.Owner != nil && detail.Owner.Mid > 0 {
		mid := detail.Owner.Mid
		row.UpperID = &mid
	}
	updated := dbtime.DateTime{Time: now}
	row.UpdatedAt = &updated

	var savedUpper model.Upper
	enqueueFace := false
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		if detail.Owner != nil && detail.Owner.Mid > 0 {
			upper, err := saveUpper(tx, *detail.Owner, now)
			if err != nil {
				return err
			}
			savedUpper = upper
			enqueueFace = !strings.Contains(detail.Owner.Name, "注销") && strings.TrimSpace(upper.Face) != ""
		}
		if invalid {
			return nil
		}
		return saveVideoParts(tx, row.ID, detail.Pages, now)
	})
	if err != nil {
		return err
	}
	if row.Frozen != 0 && !wasFrozen {
		s.noteFrozen(ctx, row.ID)
	}
	if !invalid {
		if _, err := s.EnqueueCover(nil, CoverKindVideo, row.ID, row.Cover); err != nil {
			return err
		}
		if err := s.queuePartDownloads(ctx, row.ID, now); err != nil {
			return err
		}
		if err := s.queueDanmaku(ctx, row.ID, now); err != nil {
			return err
		}
	}
	if enqueueFace {
		_, err = s.EnqueueCover(nil, CoverKindAvatar, savedUpper.ID, savedUpper.Face)
		return err
	}
	return nil
}

func (s Sync) noteFrozen(ctx context.Context, videoID int64) {
	if s.NotifyFrozen != nil && videoID > 0 {
		s.NotifyFrozen(ctx, videoID)
	}
}

// saveUpper 按 mid 写入或更新 UP 主。名字含「注销」时不改头像。
func saveUpper(tx *gorm.DB, owner bilibili.VideoOwner, now time.Time) (model.Upper, error) {
	var row model.Upper
	err := tx.Where("mid = ?", owner.Mid).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		created := dbtime.DateTime{Time: now}
		row = model.Upper{Mid: owner.Mid, CreatedAt: &created}
	} else if err != nil {
		return model.Upper{}, err
	}
	if strings.Contains(owner.Name, "注销") {
		if !strings.Contains(row.Name, "注销") {
			row.Name = owner.Name + " (注销)"
		}
	} else {
		row.Name = owner.Name
		row.Face = owner.Face
	}
	updated := dbtime.DateTime{Time: now}
	row.UpdatedAt = &updated
	if err := tx.Save(&row).Error; err != nil {
		return model.Upper{}, err
	}
	return row, nil
}

// saveVideoParts 按 cid 写入或更新分 P。本地多出来的分 P 不删除。
func saveVideoParts(tx *gorm.DB, videoID int64, pages []bilibili.VideoPage, now time.Time) error {
	var existing []model.VideoPart
	if err := tx.Where("video_id = ?", videoID).Find(&existing).Error; err != nil {
		return err
	}
	byCID := make(map[int64]model.VideoPart, len(existing))
	for _, part := range existing {
		byCID[part.Cid] = part
	}
	for _, page := range pages {
		if page.Cid == 0 {
			continue
		}
		row, ok := byCID[page.Cid]
		if !ok {
			created := dbtime.DateTime{Time: now}
			row = model.VideoPart{VideoID: videoID, Cid: page.Cid, CreatedAt: &created}
		}
		row.Page = page.Page
		row.From = page.From
		row.Part = page.Part
		row.Duration = page.Duration
		row.Vid = page.Vid
		row.Weblink = page.Weblink
		row.Width = page.Dimension.Width
		row.Height = page.Dimension.Height
		row.Rotate = page.Dimension.Rotate
		row.FirstFrame = page.FirstFrame
		updated := dbtime.DateTime{Time: now}
		row.UpdatedAt = &updated
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
