package download

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func TestReconcileDownloadsNotesCanonicalFileAndRecounts(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 11, 23, 6, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: video.ID, Cid: 99, Page: 1, From: "vupload", Part: "P1", Duration: 12,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	kept := videoFilePath(dir, video.ID, 99)
	if err := os.MkdirAll(filepath.Dir(kept), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(kept, []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	downloadedAt := time.Date(2026, 10, 30, 11, 22, 33, 0, dbtime.Location())
	if err := os.Chtimes(kept, downloadedAt, downloadedAt); err != nil {
		t.Fatal(err)
	}

	result, err := ReconcileDownloads(context.Background(), conn, dir, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Noted != 1 || result.Recounted != 1 {
		t.Fatalf("result = %+v", result)
	}
	var savedPart model.VideoPart
	if err := conn.First(&savedPart, part.ID).Error; err != nil {
		t.Fatal(err)
	}
	wantPath := storageVideoPath(video.ID, 99)
	if savedPart.VideoDownloadPath == nil || *savedPart.VideoDownloadPath != wantPath {
		t.Fatalf("path = %v", savedPart.VideoDownloadPath)
	}
	if savedPart.VideoDownloadedAt == nil || !savedPart.VideoDownloadedAt.Equal(downloadedAt) {
		t.Fatalf("part downloaded at = %v", savedPart.VideoDownloadedAt)
	}
	var saved model.Video
	if err := conn.First(&saved, video.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.VideoDownloadedNum != 1 || saved.AudioDownloadedNum != 0 {
		t.Fatalf("counts video=%d audio=%d", saved.VideoDownloadedNum, saved.AudioDownloadedNum)
	}
	if saved.VideoDownloadedAt == nil || !saved.VideoDownloadedAt.Equal(downloadedAt) {
		t.Fatalf("video downloaded at = %v", saved.VideoDownloadedAt)
	}
	var jobs int64
	if err := conn.Model(&model.Job{}).Count(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if jobs != 0 {
		t.Fatalf("tasks = %d", jobs)
	}

	again, err := ReconcileDownloads(context.Background(), conn, dir, video.ID, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if again.Noted != 0 || again.Recounted != 0 {
		t.Fatalf("second = %+v", again)
	}
}

func TestReconcileDownloadsRecountsExistingPathWithoutRewritingIt(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 11, 23, 6, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp, VideoDownloadedNum: 4,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	kept := "/storage/videos/kept.mp4"
	part := model.VideoPart{
		VideoID: video.ID, Cid: 5, Page: 1, From: "vupload", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", VideoDownloadPath: &kept, VideoDownloadedAt: &stamp, CreatedAt: &stamp,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	disk := resolveStored(dir, kept)
	if err := os.MkdirAll(filepath.Dir(disk), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(disk, []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := model.VideoPart{
		VideoID: video.ID, Cid: 6, Page: 2, From: "vupload", Part: "P2", Duration: 3,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&missing).Error; err != nil {
		t.Fatal(err)
	}

	result, err := ReconcileDownloads(context.Background(), conn, dir, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Noted != 0 || result.Recounted != 1 {
		t.Fatalf("result = %+v", result)
	}
	var savedPart model.VideoPart
	if err := conn.First(&savedPart, part.ID).Error; err != nil {
		t.Fatal(err)
	}
	if savedPart.VideoDownloadPath == nil || *savedPart.VideoDownloadPath != kept {
		t.Fatalf("path = %v", savedPart.VideoDownloadPath)
	}
	var saved model.Video
	if err := conn.First(&saved, video.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.VideoDownloadedNum != 1 {
		t.Fatalf("count = %d", saved.VideoDownloadedNum)
	}
}

func TestReconcileDownloadsPrefersLegacyNameAndRepairsBrokenPath(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 11, 23, 6, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	broken := "/storage/videos/missing.mp4"
	legacyPart := model.VideoPart{
		VideoID: video.ID, Cid: 8, Page: 1, From: "vupload", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&legacyPart).Error; err != nil {
		t.Fatal(err)
	}
	legacy := legacyVideoPath(video.ID, 1)
	if err := os.MkdirAll(filepath.Join(dir, "public", "videos"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(resolveStored(dir, legacy), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(videoFilePath(dir, video.ID, 8), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	brokenPart := model.VideoPart{
		VideoID: video.ID, Cid: 9, Page: 2, From: "vupload", Part: "P2", Duration: 3,
		Vid: "", Weblink: "", VideoDownloadPath: &broken, CreatedAt: &stamp,
	}
	if err := conn.Create(&brokenPart).Error; err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(videoFilePath(dir, video.ID, 9), []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ReconcileDownloads(context.Background(), conn, dir, video.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Noted != 2 || result.Recounted != 1 {
		t.Fatalf("result = %+v", result)
	}
	var savedLegacy model.VideoPart
	if err := conn.First(&savedLegacy, legacyPart.ID).Error; err != nil {
		t.Fatal(err)
	}
	if savedLegacy.VideoDownloadPath == nil || *savedLegacy.VideoDownloadPath != legacy {
		t.Fatalf("legacy path = %v", savedLegacy.VideoDownloadPath)
	}
	var savedBroken model.VideoPart
	if err := conn.First(&savedBroken, brokenPart.ID).Error; err != nil {
		t.Fatal(err)
	}
	want := storageVideoPath(video.ID, 9)
	if savedBroken.VideoDownloadPath == nil || *savedBroken.VideoDownloadPath != want {
		t.Fatalf("repaired path = %v", savedBroken.VideoDownloadPath)
	}
}

func TestReconcileDownloadsNotesAudioAndSkipsOthers(t *testing.T) {
	conn := openDownloadDB(t)
	dir := t.TempDir()
	now := time.Date(2026, 11, 23, 6, 0, 0, 0, dbtime.Location())
	stamp := dbtime.DateTime{Time: now}
	audioVideo := model.Video{
		Link: "https://www.bilibili.com/audio/au7", Title: "音频", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 12, CreatedAt: &stamp,
	}
	other := model.Video{
		Link: "https://www.bilibili.com/video/BV2", Title: "别的", Intro: "", Cover: "",
		Bvid: "BV2", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	deletedAt := stamp
	deleted := model.Video{
		Link: "https://www.bilibili.com/video/BV3", Title: "已删", Intro: "", Cover: "",
		Bvid: "BV3", Pubtime: stamp, Type: 2, CreatedAt: &stamp, DeletedAt: &deletedAt,
	}
	for _, video := range []*model.Video{&audioVideo, &other, &deleted} {
		if err := conn.Create(video).Error; err != nil {
			t.Fatal(err)
		}
	}
	audio := model.AudioPart{VideoID: audioVideo.ID, Sid: 7, Duration: 30, CreatedAt: &stamp}
	if err := conn.Create(&audio).Error; err != nil {
		t.Fatal(err)
	}
	otherPart := model.VideoPart{
		VideoID: other.ID, Cid: 3, Page: 1, From: "vupload", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&otherPart).Error; err != nil {
		t.Fatal(err)
	}
	deletedPart := model.VideoPart{
		VideoID: deleted.ID, Cid: 4, Page: 1, From: "vupload", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&deletedPart).Error; err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "public", "videos"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyAudioFilePath(dir, 7), []byte("mp3"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(videoFilePath(dir, other.ID, 3), []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(videoFilePath(dir, deleted.ID, 4), []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ReconcileDownloads(context.Background(), conn, dir, audioVideo.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Noted != 1 || result.Recounted != 1 {
		t.Fatalf("result = %+v", result)
	}
	var savedAudio model.AudioPart
	if err := conn.First(&savedAudio, audio.ID).Error; err != nil {
		t.Fatal(err)
	}
	want := legacyStorageAudioPath(7)
	if savedAudio.AudioDownloadPath == nil || *savedAudio.AudioDownloadPath != want {
		t.Fatalf("audio path = %v", savedAudio.AudioDownloadPath)
	}
	var saved model.Video
	if err := conn.First(&saved, audioVideo.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.AudioDownloadedNum != 1 || saved.VideoDownloadedNum != 0 {
		t.Fatalf("counts video=%d audio=%d", saved.VideoDownloadedNum, saved.AudioDownloadedNum)
	}
	var untouched model.VideoPart
	if err := conn.First(&untouched, otherPart.ID).Error; err != nil {
		t.Fatal(err)
	}
	if untouched.VideoDownloadPath != nil {
		t.Fatalf("other path = %v", untouched.VideoDownloadPath)
	}
	var stillDeleted model.VideoPart
	if err := conn.First(&stillDeleted, deletedPart.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stillDeleted.VideoDownloadPath != nil {
		t.Fatalf("deleted path = %v", stillDeleted.VideoDownloadPath)
	}
}
