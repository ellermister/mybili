package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

// Video 用播放地址接口下载一个分 P 的画面和声音，再用 ffmpeg 原样合成 mp4。
// 文件落盘后返回 PartReady。旁路由调用方处理，不在这里入队。
func (r Runner) Video(ctx context.Context, taskID int64, payload partPayload) (PartReady, error) {
	if r.DB == nil {
		return PartReady{}, fmt.Errorf("downloader is incomplete")
	}
	var part model.VideoPart
	if err := r.DB.WithContext(ctx).First(&part, payload.PartID).Error; err != nil {
		return PartReady{}, err
	}
	var video model.Video
	if err := r.DB.WithContext(ctx).First(&video, part.VideoID).Error; err != nil {
		return PartReady{}, err
	}
	if video.DeletedAt != nil || video.Invalid != 0 {
		return PartReady{}, fmt.Errorf("视频已失效")
	}
	cid := part.Cid
	if cid <= 0 {
		cid = payload.Cid
	}
	outPath := videoFilePath(r.DataDir, video.ID, cid)
	if info, err := os.Stat(outPath); err == nil && info.Size() > 0 {
		r.log(ctx, taskID, "文件已存在，跳过下载")
		r.progress(ctx, taskID, 100, 100, "文件已存在")
		if err := r.markVideoFile(ctx, video.ID, part.ID, outPath); err != nil {
			return PartReady{}, err
		}
		return PartReady{PartID: part.ID, Fresh: false}, nil
	}
	preset, skip, err := qualityForVideo(ctx, r.DB.WithContext(ctx), video.ID)
	if err != nil {
		return PartReady{}, err
	}
	if skip {
		r.log(ctx, taskID, "清晰度为不下载，跳过")
		return PartReady{}, errSkipDownload
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return PartReady{}, err
	}
	tmpRoot := filepath.Join(r.DataDir, "public", "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return PartReady{}, err
	}
	tmp, err := os.MkdirTemp(tmpRoot, "part-")
	if err != nil {
		return PartReady{}, err
	}
	defer os.RemoveAll(tmp)
	client, err := r.playClient(ctx)
	if err != nil {
		return PartReady{}, err
	}
	r.log(ctx, taskID, "获取播放地址 cid="+strconv.FormatInt(cid, 10))
	dash, err := fetchPlay(ctx, client, cid, video.ID, video.Bvid)
	if err != nil {
		return PartReady{}, err
	}
	r.log(ctx, taskID, "清晰度 "+qualityPresetName(preset))
	picture, sound, err := pickStreams(streamsOf(dash.Video), streamsOf(dash.Audio), preset)
	if err != nil {
		return PartReady{}, err
	}
	r.log(ctx, taskID, "画面 "+picture.label())
	r.log(ctx, taskID, "声音 "+sound.label())
	meter := byteMeter{logged: -1}
	videoPath := filepath.Join(tmp, "video.m4s")
	audioPath := filepath.Join(tmp, "audio.m4s")
	if err := r.downloadStream(ctx, taskID, "视频", picture, videoPath, &meter, true, client.CookieHeader); err != nil {
		return PartReady{}, err
	}
	if err := r.downloadStream(ctx, taskID, "音频", sound, audioPath, &meter, false, client.CookieHeader); err != nil {
		return PartReady{}, err
	}
	r.progress(ctx, taskID, 99, 100, "合并文件")
	r.log(ctx, taskID, "合并音视频")
	merged := filepath.Join(tmp, "merged.mp4")
	if err := r.merge(ctx, videoPath, audioPath, merged, picture.CodecID == codecHEVC); err != nil {
		return PartReady{}, err
	}
	if err := os.Rename(merged, outPath); err != nil {
		return PartReady{}, err
	}
	if info, err := os.Stat(outPath); err != nil || info.Size() == 0 {
		_ = os.Remove(outPath)
		return PartReady{}, fmt.Errorf("合并结束但没有生成文件")
	}
	r.progress(ctx, taskID, 100, 100, "下载完成")
	if err := r.markVideoFile(ctx, video.ID, part.ID, outPath); err != nil {
		return PartReady{}, err
	}
	return PartReady{PartID: part.ID, Fresh: true}, nil
}

func videoFilePath(dataDir string, videoID, cid int64) string {
	name := strconv.FormatInt(videoID, 10) + "." + strconv.FormatInt(cid, 10) + ".mp4"
	return filepath.Join(dataDir, "public", "videos", name)
}

func storageVideoPath(videoID, cid int64) string {
	return "/storage/videos/" + strconv.FormatInt(videoID, 10) + "." + strconv.FormatInt(cid, 10) + ".mp4"
}

func (r Runner) markVideoFile(ctx context.Context, videoID, partID int64, outPath string) error {
	now := dbtime.DateTime{Time: dbtime.Now()}
	rel := storageVideoPath(videoID, cidFromPath(outPath))
	err := r.DB.WithContext(ctx).Model(&model.VideoPart{}).Where("id = ?", partID).Updates(map[string]any{
		"video_download_path": rel,
		"video_downloaded_at": now,
		"updated_at":          now,
	}).Error
	if err != nil {
		return err
	}
	var count int64
	if err := r.DB.WithContext(ctx).Model(&model.VideoPart{}).
		Where("video_id = ? AND video_download_path IS NOT NULL AND video_download_path <> ''", videoID).
		Count(&count).Error; err != nil {
		return err
	}
	return r.DB.WithContext(ctx).Model(&model.Video{}).Where("id = ?", videoID).Updates(map[string]any{
		"video_downloaded_num": count,
		"video_downloaded_at":  now,
		"updated_at":           now,
	}).Error
}

func cidFromPath(path string) int64 {
	base := strings.TrimSuffix(filepath.Base(path), ".mp4")
	parts := strings.Split(base, ".")
	if len(parts) != 2 {
		return 0
	}
	cid, _ := strconv.ParseInt(parts[1], 10, 64)
	return cid
}
