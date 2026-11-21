package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"mybili/internal/bilibili"
)

const (
	playFnval = "80"
	fetchUA   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0"
	fetchFrom = "https://www.bilibili.com"

	fetchDialTimeout   = 5 * time.Second
	fetchTLSTimeout    = 5 * time.Second
	fetchHeaderTimeout = 15 * time.Second
	// fetchIdleTimeout 是正文停住多久算失败。没有整段下载的总时长限制。
	fetchIdleTimeout = 60 * time.Second
)

type dashBody struct {
	Video []dashStream `json:"video"`
	Audio []dashStream `json:"audio"`
}

type dashStream struct {
	ID        int      `json:"id"`
	BaseURL   string   `json:"baseUrl"`
	BaseURL2  string   `json:"base_url"`
	Backup    []string `json:"backupUrl"`
	Backup2   []string `json:"backup_url"`
	Bandwidth int      `json:"bandwidth"`
	CodecID   int      `json:"codecid"`
	Codecs    string   `json:"codecs"`
	MimeType  string   `json:"mimeType"`
	MimeType2 string   `json:"mime_type"`
}

func (s dashStream) toStream() stream {
	return stream{
		ID:        s.ID,
		Bandwidth: s.Bandwidth,
		CodecID:   s.CodecID,
		Codecs:    s.Codecs,
		URLs:      s.urls(),
	}
}

func (s dashStream) urls() []string {
	var out []string
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		for _, item := range out {
			if item == raw {
				return
			}
		}
		out = append(out, raw)
	}
	add(s.BaseURL)
	add(s.BaseURL2)
	for _, item := range s.Backup {
		add(item)
	}
	for _, item := range s.Backup2 {
		add(item)
	}
	return out
}

func playQuery(cid, avid int64, bvid string) (string, error) {
	if cid <= 0 {
		return "", fmt.Errorf("缺少 cid")
	}
	query := url.Values{}
	query.Set("fnval", playFnval)
	query.Set("cid", strconv.FormatInt(cid, 10))
	if avid > 0 {
		query.Set("avid", strconv.FormatInt(avid, 10))
	} else if strings.TrimSpace(bvid) != "" {
		query.Set("bvid", strings.TrimSpace(bvid))
	} else {
		return "", fmt.Errorf("缺少 avid 或 bvid")
	}
	return "/x/player/playurl?" + query.Encode(), nil
}

func (r Runner) playClient(ctx context.Context) (*bilibili.Client, error) {
	content, err := readSettingString(r.DB, "cookies_content")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("没有 Cookie")
	}
	cookies, err := bilibili.ParseNetscape(content)
	if err != nil {
		return nil, err
	}
	client := bilibili.NewClient(cookies)
	if r.PlayBase != "" {
		client.BaseURL = r.PlayBase
	}
	client.OnError = func(err error) {
		if bilibili.IsCookieExpired(err) && r.OnCookie != nil {
			r.OnCookie(ctx)
		}
	}
	return client, nil
}

func fetchPlay(ctx context.Context, client *bilibili.Client, cid, avid int64, bvid string) (dashBody, error) {
	path, err := playQuery(cid, avid, bvid)
	if err != nil {
		return dashBody{}, err
	}
	body, err := client.Get(ctx, path)
	if err != nil {
		return dashBody{}, err
	}
	raw, ok, err := client.Parse(body)
	if err != nil {
		return dashBody{}, err
	}
	if !ok {
		return dashBody{}, fmt.Errorf("播放地址没有数据")
	}
	var data struct {
		Dash *dashBody `json:"dash"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return dashBody{}, fmt.Errorf("播放地址解析失败: %w", err)
	}
	if data.Dash == nil {
		return dashBody{}, fmt.Errorf("播放地址没有 DASH")
	}
	return *data.Dash, nil
}

func streamsOf(items []dashStream) []stream {
	out := make([]stream, 0, len(items))
	for _, item := range items {
		out = append(out, item.toStream())
	}
	return out
}

type byteMeter struct {
	videoSize int64
	audioSize int64
	videoGot  int64
	audioGot  int64
	videoDone bool
	audioDone bool
	logged    int
}

func (m byteMeter) percent() int {
	total := m.videoSize + m.audioSize
	got := m.videoGot + m.audioGot
	if total > 0 {
		value := int(got * 99 / total)
		if value > 99 {
			return 99
		}
		return value
	}
	value := 0
	if m.videoDone {
		value = 49
	}
	if m.audioDone {
		value = 99
	}
	return value
}

func (r Runner) downloadStream(ctx context.Context, taskID int64, name string, item stream, dest string, meter *byteMeter, video bool, cookie string) error {
	r.log(ctx, taskID, "下载"+name+" "+item.label())
	known := meter.videoSize
	if !video {
		known = meter.audioSize
	}
	err := r.fetchFirst(ctx, taskID, name, item.URLs, dest, known, cookie, func(got, total int64) {
		if video {
			meter.videoGot = got
			if total > 0 {
				meter.videoSize = total
			}
		} else {
			meter.audioGot = got
			if total > 0 {
				meter.audioSize = total
			}
		}
		r.reportDownload(ctx, taskID, meter)
	})
	if err != nil {
		return err
	}
	if video {
		meter.videoDone = true
		if meter.videoSize == 0 {
			meter.videoSize = meter.videoGot
		}
	} else {
		meter.audioDone = true
		if meter.audioSize == 0 {
			meter.audioSize = meter.audioGot
		}
	}
	r.reportDownload(ctx, taskID, meter)
	return nil
}

func (r Runner) reportDownload(ctx context.Context, taskID int64, meter *byteMeter) {
	percent := meter.percent()
	text := "已下载 " + sizeText(meter.videoGot+meter.audioGot)
	total := meter.videoSize + meter.audioSize
	if total > 0 {
		text += " / " + sizeText(total)
	}
	r.progress(ctx, taskID, percent, 100, text)
	if percent == meter.logged {
		return
	}
	meter.logged = percent
	r.log(ctx, taskID, strconv.Itoa(percent)+"% "+text)
}

func (r Runner) cdnClient() *http.Client {
	if r.CDN != nil {
		return r.CDN
	}
	return &http.Client{Transport: &http.Transport{
		DialContext:           (&net.Dialer{Timeout: fetchDialTimeout}).DialContext,
		TLSHandshakeTimeout:   fetchTLSTimeout,
		ResponseHeaderTimeout: fetchHeaderTimeout,
	}}
}

func (r Runner) downloadOne(ctx context.Context, taskID int64, title string, urls []string, dest string, size int64, cookie string) error {
	r.log(ctx, taskID, title)
	logged := -1
	return r.fetchFirst(ctx, taskID, "音频", urls, dest, size, cookie, func(got, total int64) {
		if total > 0 {
			size = total
		}
		percent := 0
		if size > 0 {
			percent = int(got * 99 / size)
			if percent > 99 {
				percent = 99
			}
		}
		text := "已下载 " + sizeText(got)
		if size > 0 {
			text += " / " + sizeText(size)
		}
		r.progress(ctx, taskID, percent, 100, text)
		if percent == logged {
			return
		}
		logged = percent
		r.log(ctx, taskID, strconv.Itoa(percent)+"% "+text)
	})
}

func (r Runner) fetchFirst(ctx context.Context, taskID int64, name string, urls []string, dest string, known int64, cookie string, onBytes func(got, total int64)) error {
	if len(urls) == 0 || urls[0] == "" {
		return fmt.Errorf("没有下载地址")
	}
	err := r.fetchLogged(ctx, taskID, name, urls[0], dest, known, cookie, onBytes)
	if err == nil || ctx.Err() != nil || !connectFailed(err) {
		return err
	}
	var last error
	for _, raw := range urls[1:] {
		raw = strings.TrimSpace(raw)
		if raw == "" || raw == urls[0] {
			continue
		}
		_ = os.Remove(dest)
		r.log(ctx, taskID, name+" 更换备用地址 "+raw)
		last = r.fetchLogged(ctx, taskID, name, raw, dest, known, cookie, onBytes)
		if last == nil || ctx.Err() != nil || !connectFailed(last) {
			return last
		}
	}
	if last != nil {
		return last
	}
	return err
}

func connectFailed(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

func (r Runner) fetchLogged(ctx context.Context, taskID int64, name, raw, dest string, total int64, cookie string, onBytes func(got, total int64)) error {
	r.log(ctx, taskID, name+" 地址 "+raw)
	err := fetchOne(ctx, r.cdnClient(), raw, dest, total, cookie, onBytes)
	if err != nil && ctx.Err() == nil {
		r.log(ctx, taskID, name+" 失败 "+raw+" "+err.Error())
	}
	return err
}

func fetchOne(ctx context.Context, client *http.Client, raw, dest string, total int64, cookie string, onBytes func(got, total int64)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return err
	}
	applyFetchHeaders(req, cookie)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 180))
		return fmt.Errorf("下载 HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}
	if resp.ContentLength > 0 {
		total = resp.ContentLength
	}
	if onBytes != nil && total > 0 {
		onBytes(0, total)
	}
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := make([]byte, 256*1024)
	var got int64
	var lastReport int64
	for {
		n, readErr := readWithIdle(ctx, resp.Body, buf, fetchIdleTimeout)
		if n > 0 {
			if _, err := file.Write(buf[:n]); err != nil {
				return err
			}
			got += int64(n)
			if onBytes != nil && (got-lastReport >= 1024*1024 || readErr == io.EOF) {
				onBytes(got, total)
				lastReport = got
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if onBytes != nil {
		onBytes(got, total)
	}
	if got == 0 {
		return fmt.Errorf("下载结果是空文件")
	}
	return nil
}

func readWithIdle(ctx context.Context, body io.ReadCloser, buf []byte, idle time.Duration) (int, error) {
	type result struct {
		n   int
		err error
	}
	done := make(chan result, 1)
	go func() {
		n, err := body.Read(buf)
		done <- result{n, err}
	}()
	timer := time.NewTimer(idle)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		_ = body.Close()
		return 0, ctx.Err()
	case <-timer.C:
		_ = body.Close()
		return 0, fmt.Errorf("连续 %d 秒没有收到数据", int(idle.Seconds()))
	case res := <-done:
		return res.n, res.err
	}
}

func applyFetchHeaders(req *http.Request, cookie string) {
	req.Header.Set("User-Agent", fetchUA)
	req.Header.Set("Referer", fetchFrom)
	if strings.TrimSpace(cookie) != "" {
		req.Header.Set("Cookie", cookie)
	}
}

func (r Runner) merge(ctx context.Context, videoPath, audioPath, outPath string, hevc bool) error {
	args := []string{
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-map", "0:v:0",
		"-map", "1:a:0",
		"-c", "copy",
	}
	if hevc {
		args = append(args, "-tag:v", "hvc1")
	}
	args = append(args, "-movflags", "+faststart", outPath)
	cmd := exec.CommandContext(ctx, r.ffmpegBin(), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(out))
		if len(text) > 400 {
			text = text[len(text)-400:]
		}
		if text == "" {
			return fmt.Errorf("ffmpeg 合并失败: %w", err)
		}
		return fmt.Errorf("ffmpeg 合并失败: %s", text)
	}
	return nil
}

func (r Runner) ffmpegBin() string {
	if strings.TrimSpace(r.FFmpeg) != "" {
		return r.FFmpeg
	}
	return "ffmpeg"
}

func sizeText(n int64) string {
	if n < 1024 {
		return strconv.FormatInt(n, 10) + "B"
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(n)/1024)
	}
	return fmt.Sprintf("%.1fMB", float64(n)/(1024*1024))
}
