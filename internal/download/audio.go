package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

type audioPayload struct {
	VideoID int64 `json:"video_id"`
	PartID  int64 `json:"part_id"`
	Sid     int64 `json:"sid"`
}

// Audio 用音频地址接口下载一条稿件。文件本身是 m4a，不再转成 mp3。
func (r Runner) Audio(ctx context.Context, taskID int64, payload audioPayload) error {
	if r.DB == nil {
		return fmt.Errorf("downloader is incomplete")
	}
	var part model.AudioPart
	if err := r.DB.WithContext(ctx).First(&part, payload.PartID).Error; err != nil {
		return err
	}
	var video model.Video
	if err := r.DB.WithContext(ctx).First(&video, part.VideoID).Error; err != nil {
		return err
	}
	if video.DeletedAt != nil || video.Invalid != 0 {
		return fmt.Errorf("视频已失效")
	}
	sid := part.Sid
	if sid <= 0 {
		sid = payload.Sid
	}
	outPath := audioFilePath(r.DataDir, sid)
	if info, err := os.Stat(outPath); err == nil && info.Size() > 0 {
		r.log(ctx, taskID, "文件已存在，跳过下载")
		r.progress(ctx, taskID, 100, 100, "文件已存在")
		return r.markAudioFile(ctx, video.ID, part.ID, storageAudioPath(sid))
	}
	legacy := legacyAudioFilePath(r.DataDir, sid)
	if info, err := os.Stat(legacy); err == nil && info.Size() > 0 {
		r.log(ctx, taskID, "文件已存在，跳过下载")
		r.progress(ctx, taskID, 100, 100, "文件已存在")
		return r.markAudioFile(ctx, video.ID, part.ID, legacyStorageAudioPath(sid))
	}
	if _, skip, err := qualityForVideo(ctx, r.DB.WithContext(ctx), video.ID); err != nil {
		return err
	} else if skip {
		r.log(ctx, taskID, "清晰度为不下载，跳过")
		return errSkipDownload
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	client, err := r.playClient(ctx)
	if err != nil {
		return err
	}
	r.log(ctx, taskID, "获取音频地址 sid="+strconv.FormatInt(sid, 10))
	file, err := bilibili.AudioURL(ctx, client, sid)
	if err != nil {
		return err
	}
	if len(file.URLs) == 0 {
		return fmt.Errorf("没有音频地址")
	}
	tmpRoot := filepath.Join(r.DataDir, "public", "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(tmpRoot, "audio-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	partPath := filepath.Join(tmp, "audio.m4a")
	title := "下载音频"
	if file.Size > 0 {
		title += " " + sizeText(file.Size)
	}
	if err := r.downloadOne(ctx, taskID, title, file.URLs, partPath, file.Size, client.CookieHeader); err != nil {
		return err
	}
	if err := os.Rename(partPath, outPath); err != nil {
		return err
	}
	if info, err := os.Stat(outPath); err != nil || info.Size() == 0 {
		_ = os.Remove(outPath)
		return fmt.Errorf("下载结束但没有生成文件")
	}
	r.progress(ctx, taskID, 100, 100, "下载完成")
	return r.markAudioFile(ctx, video.ID, part.ID, storageAudioPath(sid))
}

func audioFilePath(dataDir string, sid int64) string {
	return filepath.Join(dataDir, "public", "videos", audioName(sid, ".m4a"))
}

func legacyAudioFilePath(dataDir string, sid int64) string {
	return filepath.Join(dataDir, "public", "videos", audioName(sid, ".mp3"))
}

func storageAudioPath(sid int64) string {
	return "/storage/videos/" + audioName(sid, ".m4a")
}

func legacyStorageAudioPath(sid int64) string {
	return "/storage/videos/" + audioName(sid, ".mp3")
}

func audioName(sid int64, ext string) string {
	return "au" + strconv.FormatInt(sid, 10) + ext
}

func audioReady(dataDir string, stored *string, sid int64) bool {
	if fileReady(dataDir, stored, audioFilePath(dataDir, sid)) {
		return true
	}
	return fileExists(legacyAudioFilePath(dataDir, sid))
}

func (r Runner) markAudioFile(ctx context.Context, videoID, partID int64, stored string) error {
	now := dbtime.DateTime{Time: dbtime.Now()}
	err := r.DB.WithContext(ctx).Model(&model.AudioPart{}).Where("id = ?", partID).Updates(map[string]any{
		"audio_download_path": stored,
		"audio_downloaded_at": now,
		"updated_at":          now,
	}).Error
	if err != nil {
		return err
	}
	var count int64
	if err := r.DB.WithContext(ctx).Model(&model.AudioPart{}).
		Where("video_id = ? AND audio_download_path IS NOT NULL AND audio_download_path <> ''", videoID).
		Count(&count).Error; err != nil {
		return err
	}
	return r.DB.WithContext(ctx).Model(&model.Video{}).Where("id = ?", videoID).Updates(map[string]any{
		"audio_downloaded_num": count,
		"updated_at":           now,
	}).Error
}
