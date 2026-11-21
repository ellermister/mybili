package download

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"mybili/internal/taskrun"
)

const (
	// TypeVideo 是视频文件下载。
	TypeVideo = "download-video"
	// TypeDanmaku 是一个分 P 的弹幕下载。
	TypeDanmaku = "download-danmaku"
	// TypeAudio 是音频稿件下载。
	TypeAudio = "download-audio"
)

// Logs 把执行过程写到 Redis。worker 传 *taskrun.Tasks。
type Logs interface {
	AppendLog(ctx context.Context, id int64, line string) error
	SetProgress(ctx context.Context, id int64, progress taskrun.Progress) error
}

// PartReady 表示这一分 P 的文件已经记到库里。
// Fresh 为 true 表示这次确实执行了下载，文件本来就在时为 false。
type PartReady struct {
	PartID int64
	Fresh  bool
}

// Hooks 是下载成功之后的旁路。调用失败不得影响下载任务本身。
type Hooks struct {
	PartReady func(context.Context, PartReady)
}

// Runner 下载视频和弹幕。它不负责入队。
type Runner struct {
	DB      *gorm.DB
	DataDir string
	Logs    Logs
	// PlayBase 是播放地址接口的主机。为空时用 Bilibili 正式接口。
	PlayBase string
	// CDN 下载画面和声音。为空时用带请求头超时的客户端。
	CDN *http.Client
	// FFmpeg 合并音视频。为空时用 PATH 里的 ffmpeg。
	FFmpeg string
	// OnCookie 在弹幕接口报 Cookie 过期时调用。
	OnCookie func(context.Context)
}

func (r Runner) log(ctx context.Context, id int64, line string) {
	if r.Logs == nil || line == "" {
		return
	}
	_ = r.Logs.AppendLog(ctx, id, line)
}

func (r Runner) progress(ctx context.Context, id int64, done, total int, text string) {
	if r.Logs == nil {
		return
	}
	_ = r.Logs.SetProgress(ctx, id, taskrun.Progress{Done: done, Total: total, Text: text})
}

type partPayload struct {
	VideoID int64 `json:"video_id"`
	PartID  int64 `json:"part_id"`
	Page    int   `json:"page"`
	Cid     int64 `json:"cid"`
}
