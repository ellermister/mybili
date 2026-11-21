package download

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/migrate"
	"mybili/internal/model"
)

func TestEnqueueLostOnlyMissingRecordedFiles(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 31, 8, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	on := `"on"`
	if err := conn.Create(&model.Setting{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	lost := "/storage/videos/lost.mp4"
	part := model.VideoPart{
		VideoID: video.ID, Cid: 99, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", VideoDownloadPath: &lost, VideoDownloadedAt: &stamp, CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	fresh := model.VideoPart{
		VideoID: video.ID, Cid: 100, Page: 2, From: "vupload", Part: "P2", Duration: 8,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&fresh).Error; err != nil {
		t.Fatal(err)
	}
	audioPath := "/storage/videos/au7.mp3"
	audio := model.AudioPart{
		VideoID: video.ID, Sid: 7, Duration: 30, AudioDownloadPath: &audioPath, AudioDownloadedAt: &stamp, CreatedAt: &stamp,
	}
	if err := conn.Create(&audio).Error; err != nil {
		t.Fatal(err)
	}
	count, err := EnqueueLost(context.Background(), conn, dir, 0, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("enqueued = %d", count)
	}
	var jobs int64
	if err := conn.Model(&model.Job{}).Count(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if jobs != 2 {
		t.Fatalf("tasks = %d", jobs)
	}

	kept := videoFilePath(dir, video.ID, 99)
	if err := os.MkdirAll(filepath.Dir(kept), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(kept, []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := EnqueueLost(context.Background(), conn, dir, video.ID, now.Add(time.Minute), nil)
	if err != nil {
		t.Fatal(err)
	}
	if again != 0 {
		t.Fatalf("second scan enqueued = %d", again)
	}
}

func TestScanStoredPartIDs(t *testing.T) {
	conn := openDownloadDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: video.ID, Cid: 5, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	var rows []storedPart
	err := conn.Table("video_parts").
		Select(`video_parts.id, video_parts.video_id, video_parts.page, video_parts.cid, video_parts.duration,
			videos.title, videos.duration as video_duration, videos.fav_time, videos.pubtime,
			video_parts.video_download_path as download_path, video_parts.video_downloaded_at as downloaded_at`).
		Joins("JOIN videos ON videos.id = video_parts.video_id").
		Scan(&rows).Error
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != part.ID || rows[0].VideoID != video.ID || rows[0].Cid != 5 {
		t.Fatalf("scanned = %+v", rows)
	}
}

func TestEnqueuePartsOnceWhenFileMissing(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	on := `"on"`
	if err := conn.Create(&model.Setting{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: video.ID, Cid: 5, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	count, err := EnqueueParts(context.Background(), conn, dir, video.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("enqueued = %d", count)
	}
	done := model.Job{}
	if err := conn.Where("type = ?", TypeVideo).First(&done).Error; err != nil {
		t.Fatal(err)
	}
	var body partPayload
	if err := json.Unmarshal([]byte(done.Payload), &body); err != nil {
		t.Fatal(err)
	}
	if body.VideoID != video.ID || body.PartID != part.ID || body.Cid != 5 || body.Page != 1 {
		t.Fatalf("payload = %s", done.Payload)
	}
	if err := conn.Model(&done).Update("status", "done").Error; err != nil {
		t.Fatal(err)
	}
	again, err := EnqueueParts(context.Background(), conn, dir, video.ID, now.Add(time.Minute), nil)
	if err != nil {
		t.Fatal(err)
	}
	if again != 0 {
		t.Fatalf("finished task reopened = %d", again)
	}
	kept := videoFilePath(dir, video.ID, 5)
	if err := os.MkdirAll(filepath.Dir(kept), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(kept, []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	other := model.Video{
		Link: "b", Title: "有文件", Intro: "", Cover: "", Bvid: "BV2", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	ready := model.VideoPart{
		VideoID: other.ID, Cid: 6, Page: 1, From: "vupload", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&ready).Error; err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(videoFilePath(dir, other.ID, 6), []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	readyPath := videoFilePath(dir, other.ID, 6)
	downloadedAt := time.Date(2026, 10, 30, 11, 22, 33, 0, dbtime.Location())
	if err := os.Chtimes(readyPath, downloadedAt, downloadedAt); err != nil {
		t.Fatal(err)
	}
	skipped, err := EnqueueParts(context.Background(), conn, dir, other.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("existing file enqueued = %d", skipped)
	}
	var savedPart model.VideoPart
	if err := conn.First(&savedPart, ready.ID).Error; err != nil {
		t.Fatal(err)
	}
	wantPath := storageVideoPath(other.ID, 6)
	if savedPart.VideoDownloadPath == nil || *savedPart.VideoDownloadPath != wantPath {
		t.Fatalf("path = %v", savedPart.VideoDownloadPath)
	}
	if savedPart.VideoDownloadedAt == nil || !savedPart.VideoDownloadedAt.Equal(downloadedAt) {
		t.Fatalf("part downloaded at = %v", savedPart.VideoDownloadedAt)
	}
	var savedVideo model.Video
	if err := conn.First(&savedVideo, other.ID).Error; err != nil {
		t.Fatal(err)
	}
	if savedVideo.VideoDownloadedNum != 1 || savedVideo.VideoDownloadedAt == nil || !savedVideo.VideoDownloadedAt.Equal(downloadedAt) {
		t.Fatalf("video downloaded num=%d at=%v", savedVideo.VideoDownloadedNum, savedVideo.VideoDownloadedAt)
	}
}

func TestEnqueueDanmakuSkipsFreshAndRepeats(t *testing.T) {
	conn := openDownloadDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	on := `"on"`
	if err := conn.Create(&[]model.Setting{
		{Name: "danmaku_download_enabled", Value: &on, CreatedAt: &stamp},
		{Name: "multi_partition_download_enabled", Value: &on, CreatedAt: &stamp},
	}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	freshAt := dbtime.DateTime{Time: now.Add(-24 * time.Hour)}
	oldAt := dbtime.DateTime{Time: now.Add(-8 * 24 * time.Hour)}
	parts := []model.VideoPart{
		{VideoID: video.ID, Cid: 1, Page: 1, From: "vupload", Part: "新", Duration: 3, Vid: "", Weblink: "", CreatedAt: &stamp},
		{VideoID: video.ID, Cid: 2, Page: 2, From: "vupload", Part: "七天内", Duration: 3, Vid: "", Weblink: "", DanmakuDownloadedAt: &freshAt, CreatedAt: &stamp},
		{VideoID: video.ID, Cid: 3, Page: 3, From: "vupload", Part: "过期", Duration: 3, Vid: "", Weblink: "", DanmakuDownloadedAt: &oldAt, CreatedAt: &stamp},
	}
	if err := conn.Create(&parts).Error; err != nil {
		t.Fatal(err)
	}
	count, err := EnqueueDanmaku(context.Background(), conn, video.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("enqueued = %d", count)
	}
	var fresh int64
	if err := conn.Model(&model.Job{}).Where("unique_key = ?", fmt.Sprintf("download:danmaku:%d", parts[1].ID)).Count(&fresh).Error; err != nil {
		t.Fatal(err)
	}
	if fresh != 0 {
		t.Fatal("fresh danmaku was queued")
	}
	again, err := EnqueueDanmaku(context.Background(), conn, video.ID, now.Add(time.Minute), nil)
	if err != nil {
		t.Fatal(err)
	}
	if again != 0 {
		t.Fatalf("pending danmaku reopened = %d", again)
	}
}

func TestEnqueueDanmakuFollowsMultiPartition(t *testing.T) {
	conn := openDownloadDB(t)
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	on := `"on"`
	off := `"off"`
	if err := conn.Create(&[]model.Setting{
		{Name: "danmaku_download_enabled", Value: &on, CreatedAt: &stamp},
		{Name: "multi_partition_download_enabled", Value: &off, CreatedAt: &stamp},
	}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	parts := []model.VideoPart{
		{VideoID: video.ID, Cid: 1, Page: 1, From: "vupload", Part: "P1", Duration: 3, Vid: "", Weblink: "", CreatedAt: &stamp},
		{VideoID: video.ID, Cid: 2, Page: 2, From: "vupload", Part: "P2", Duration: 3, Vid: "", Weblink: "", CreatedAt: &stamp},
	}
	if err := conn.Create(&parts).Error; err != nil {
		t.Fatal(err)
	}
	count, err := EnqueueDanmaku(context.Background(), conn, video.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("enqueued = %d", count)
	}
	refreshed, err := RefreshDanmaku(context.Background(), conn, video.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed != 1 {
		t.Fatalf("refreshed = %d", refreshed)
	}
	var jobs int64
	if err := conn.Model(&model.Job{}).Where("type = ?", TypeDanmaku).Count(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if jobs != 1 {
		t.Fatalf("tasks = %d", jobs)
	}
}

func TestEnqueuePartsSkipsWhenEverySourceDeclines(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	on := `"on"`
	def := `"1"`
	sources := `{"7":"skip","-3":"2"}`
	if err := conn.Create(&[]model.Setting{
		{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp},
		{Name: "video_quality", Value: &def, CreatedAt: &stamp},
		{Name: "source_quality", Value: &sources, CreatedAt: &stamp},
	}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: video.ID, Cid: 5, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&[]model.FavoriteListVideo{{FavoriteListID: 7, VideoID: video.ID}}).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.SubscriptionVideo{SubscriptionID: 3, VideoID: video.ID, Bvid: "BV1"}).Error; err != nil {
		t.Fatal(err)
	}
	count, err := EnqueueParts(context.Background(), conn, dir, video.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("mixed source enqueued = %d", count)
	}
	sources = `{"7":"skip","-3":"skip"}`
	if err := conn.Model(&model.Setting{}).Where("name = ?", "source_quality").Update("value", sources).Error; err != nil {
		t.Fatal(err)
	}
	other := model.Video{
		Link: "b", Title: "全跳过", Intro: "", Cover: "", Bvid: "BV2", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	otherPart := model.VideoPart{
		VideoID: other.ID, Cid: 6, Page: 1, From: "vupload", Part: "P1", Duration: 8,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&otherPart).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: 7, VideoID: other.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.SubscriptionVideo{SubscriptionID: 3, VideoID: other.ID, Bvid: "BV2"}).Error; err != nil {
		t.Fatal(err)
	}
	skipped, err := EnqueueParts(context.Background(), conn, dir, other.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("all skip enqueued = %d", skipped)
	}
}

func TestEnqueuePartsSkipsLaterPagesWhenMultiOff(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	on := `"on"`
	off := `"off"`
	if err := conn.Create(&[]model.Setting{
		{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp},
		{Name: "multi_partition_download_enabled", Value: &off, CreatedAt: &stamp},
	}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	parts := []model.VideoPart{
		{VideoID: video.ID, Cid: 5, Page: 1, From: "vupload", Part: "P1", Duration: 12, Vid: "", Weblink: "", CreatedAt: &stamp},
		{VideoID: video.ID, Cid: 6, Page: 2, From: "vupload", Part: "P2", Duration: 8, Vid: "", Weblink: "", CreatedAt: &stamp},
		{VideoID: video.ID, Cid: 7, Page: 3, From: "vupload", Part: "P3", Duration: 8, Vid: "", Weblink: "", CreatedAt: &stamp},
	}
	if err := conn.Create(&parts).Error; err != nil {
		t.Fatal(err)
	}
	count, err := EnqueueParts(context.Background(), conn, dir, video.ID, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("enqueued = %d", count)
	}
	var jobs []model.Job
	if err := conn.Find(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("tasks = %d", len(jobs))
	}
	var body partPayload
	if err := json.Unmarshal([]byte(jobs[0].Payload), &body); err != nil {
		t.Fatal(err)
	}
	if body.Page != 1 || body.Cid != 5 {
		t.Fatalf("payload = %s", jobs[0].Payload)
	}
}

func openDownloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "download.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}
