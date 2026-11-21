package download

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/model"
)

func TestDeleteVideosRemovesFilesAndDanmaku(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	video := model.Video{
		Link: "l", Title: "要删", Intro: "", Cover: "c", Bvid: "BV1",
		Pubtime: stamp, Type: 2, VideoDownloadedNum: 1, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	stored := "/storage/videos/" + strconv.FormatInt(video.ID, 10) + ".5.mp4"
	part := model.VideoPart{
		VideoID: video.ID, Cid: 5, Page: 1, From: "v", Part: "P", Duration: 3,
		Vid: "", Weblink: "", FirstFrame: "", VideoDownloadPath: &stored, VideoDownloadedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	audioPath := "/storage/videos/au8.mp3"
	audio := model.AudioPart{VideoID: video.ID, Sid: 8, Duration: 3, AudioDownloadPath: &audioPath, AudioDownloadedAt: &stamp}
	if err := conn.Create(&audio).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.Danmaku{VideoID: video.ID, Cid: 5, Content: "弹幕", CreatedAt: stamp}).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.Coverable{CoverID: 1, CoverableID: video.ID, CoverableType: model.CoverableVideo}).Error; err != nil {
		t.Fatal(err)
	}
	canonical := filepath.Join(dir, "public", "videos", strconv.FormatInt(video.ID, 10)+".5.mp4")
	legacy := filepath.Join(dir, "public", "videos", strconv.FormatInt(video.ID, 10)+".mp4")
	audioFile := filepath.Join(dir, "public", "videos", "au8.mp3")
	outside := filepath.Join(dir, "secret.mp4")
	for _, path := range []string{canonical, canonical + ".hash", legacy, audioFile, outside} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := conn.Model(&part).Update("video_download_path", outside).Error; err != nil {
		t.Fatal(err)
	}
	deleted, err := DeleteVideos(context.Background(), conn, dir, []int64{video.ID, video.ID, 0}, DeleteOptions{Permanent: true}, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted) != 1 || deleted[0] != video.ID {
		t.Fatalf("deleted = %v", deleted)
	}
	for _, path := range []string{canonical, canonical + ".hash", legacy, audioFile} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("file still exists %s err=%v", path, err)
		}
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatal("file outside public was removed")
	}
	var saved model.Video
	if err := conn.First(&saved, video.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.DeletedAt == nil {
		t.Fatal("video was not soft-deleted")
	}
	var danmaku int64
	if err := conn.Model(&model.Danmaku{}).Where("video_id = ?", video.ID).Count(&danmaku).Error; err != nil {
		t.Fatal(err)
	}
	if danmaku != 0 {
		t.Fatalf("danmaku = %d", danmaku)
	}
	var covers int64
	if err := conn.Model(&model.Coverable{}).Where("coverable_id = ?", video.ID).Count(&covers).Error; err != nil {
		t.Fatal(err)
	}
	if covers != 1 {
		t.Fatal("cover link was removed")
	}
	var savedPart model.VideoPart
	if err := conn.First(&savedPart, part.ID).Error; err != nil {
		t.Fatal(err)
	}
	if savedPart.VideoDownloadPath != nil || savedPart.VideoDownloadedAt != nil {
		t.Fatal("part download state was kept")
	}
}

func TestDeleteVideosCanRequeue(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	on := `"on"`
	if err := conn.Create(&model.Setting{Name: "video_download_enabled", Value: &on, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{
		Link: "l", Title: "临时", Intro: "", Cover: "c", Bvid: "BV2",
		Pubtime: stamp, Type: 2, VideoDownloadedNum: 1, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: video.ID, Cid: 6, Page: 1, From: "v", Part: "P", Duration: 3,
		Vid: "", Weblink: "", FirstFrame: "", VideoDownloadedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.Danmaku{VideoID: video.ID, Cid: 6, Content: "留着", CreatedAt: stamp}).Error; err != nil {
		t.Fatal(err)
	}
	done := model.Job{
		Queue: "bili-default", Type: TypeVideo, UniqueKey: "download:video:" + strconv.FormatInt(part.ID, 10),
		Status: jobs.StatusDone, Payload: "{}", MaxRetries: 3, CreatedAt: &stamp, UpdatedAt: &stamp,
	}
	if err := conn.Create(&done).Error; err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "public", "videos", strconv.FormatInt(video.ID, 10)+".6.mp4")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	deleted, err := DeleteVideos(context.Background(), conn, dir, []int64{video.ID}, DeleteOptions{Permanent: false, Requeue: true}, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted) != 1 {
		t.Fatalf("deleted = %v", deleted)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatal("file was kept")
	}
	var saved model.Video
	if err := conn.First(&saved, video.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.DeletedAt != nil || saved.VideoDownloadedNum != 0 {
		t.Fatalf("video = %+v", saved)
	}
	var danmaku int64
	if err := conn.Model(&model.Danmaku{}).Count(&danmaku).Error; err != nil {
		t.Fatal(err)
	}
	if danmaku != 1 {
		t.Fatalf("danmaku = %d", danmaku)
	}
	var task model.Job
	if err := conn.Where("unique_key = ?", done.UniqueKey).First(&task).Error; err != nil {
		t.Fatal(err)
	}
	if task.Status != jobs.StatusPending {
		t.Fatalf("status = %s", task.Status)
	}
}
