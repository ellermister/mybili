package download

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const (
	// QualityBest 是最高清晰度，同一档优先 H.265，音频取最大码率。
	QualityBest = 1
	// Quality1080 是 1080P。没有这一档时取更低清晰度里的最高档，音频取最大码率。
	Quality1080 = 2
	// QualityLow 是最低码率的画面和声音。
	QualityLow = 3
	// QualityHighVideo 是最高清晰度配最低码率音频。
	QualityHighVideo = 4
	// QualityHighAudio 是最低码率画面配最高码率音频。
	QualityHighAudio = 5

	qn1080    = 80
	codecAVC  = 7
	codecHEVC = 12
)

// stream 是 playurl 里的一条画面或声音。
type stream struct {
	ID        int
	Bandwidth int
	CodecID   int
	Codecs    string
	URLs      []string
}

func (s stream) label() string {
	name := s.Codecs
	switch s.CodecID {
	case codecAVC:
		name = "H.264"
	case codecHEVC:
		name = "H.265"
	}
	if name == "" {
		name = "编码 " + strconv.Itoa(s.CodecID)
	}
	return qualityName(s.ID) + " " + name + " " + fmt.Sprintf("%.2fMbps", float64(s.Bandwidth)/1_000_000)
}

func qualityName(id int) string {
	switch id {
	case 16:
		return "360P"
	case 32:
		return "480P"
	case 64:
		return "720P"
	case 80:
		return "1080P"
	case 112:
		return "1080P+"
	case 116:
		return "1080P60"
	case 120:
		return "4K"
	case 125:
		return "HDR"
	case 126:
		return "杜比视界"
	case 127:
		return "8K"
	case 30216:
		return "64K"
	case 30232:
		return "132K"
	case 30280:
		return "192K"
	default:
		return strconv.Itoa(id)
	}
}

// errSkipDownload 表示当前来源解析后不该下载。任务行会被删掉，方便以后改回清晰度时重新排队。
var errSkipDownload = errors.New("清晰度为不下载，跳过")

// videoQuality 读取设置里的默认清晰度。没有这项或值不在 1 到 5 时，用最高清晰度。
func videoQuality(db *gorm.DB) int {
	text, err := readSettingString(db, "video_quality")
	if err != nil {
		return QualityBest
	}
	preset, skip := parseDefault(text)
	if skip {
		return QualityBest
	}
	return preset
}

// qualityForVideo 按这个稿件现在还挂着的收藏夹和订阅决定档位。
// 不看来源任务是从哪一个收藏夹投进来的。没有任何来源时用默认档。
func qualityForVideo(ctx context.Context, db *gorm.DB, videoID int64) (int, bool, error) {
	settings, err := loadQualitySettings(db)
	if err != nil {
		return 0, false, err
	}
	favs, subs, err := relations(ctx, db, []Candidate{{VideoID: videoID}})
	if err != nil {
		return 0, false, err
	}
	preset, skip := settings.choose(favs[videoID], subs[videoID])
	return preset, skip, nil
}

type qualitySettings struct {
	defaultValue string
	bySource     map[int64]string
	excluded     map[int64]bool
}

func loadQualitySettings(db *gorm.DB) (qualitySettings, error) {
	var out qualitySettings
	out.bySource = map[int64]string{}
	out.excluded = map[int64]bool{}
	text, err := readSettingString(db, "video_quality")
	if err != nil {
		return out, err
	}
	out.defaultValue = text
	var sources map[string]any
	if err := readSettingJSON(db, "source_quality", &sources); err != nil {
		return out, err
	}
	for key, value := range sources {
		id, err := strconv.ParseInt(strings.TrimSpace(key), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		out.bySource[id] = settingText(value)
	}
	var fav struct {
		Enabled  bool    `json:"enabled"`
		Selected []int64 `json:"selected"`
	}
	if err := readSettingJSON(db, "fav_exclude", &fav); err != nil {
		return out, err
	}
	if fav.Enabled {
		for _, id := range fav.Selected {
			out.excluded[id] = true
		}
	}
	return out, nil
}

func (q qualitySettings) choose(favIDs, subIDs []int64) (int, bool) {
	if len(favIDs) == 0 && len(subIDs) == 0 {
		return chooseQuality(q.defaultValue, nil)
	}
	values := make([]string, 0, len(favIDs)+len(subIDs))
	for _, id := range favIDs {
		values = append(values, q.choice(id))
	}
	for _, id := range subIDs {
		values = append(values, q.choice(-id))
	}
	return chooseQuality(q.defaultValue, values)
}

func (q qualitySettings) choice(id int64) string {
	if value, ok := q.bySource[id]; ok && strings.TrimSpace(value) != "" {
		return value
	}
	if q.excluded[id] {
		return "skip"
	}
	return ""
}

// chooseQuality 在多个来源里取允许下载的更高一档。
// 来源值为空或 0 时跟随默认。没有任何来源时也用默认。全部是不下载才跳过。
func chooseQuality(defaultValue string, sourceValues []string) (int, bool) {
	fallback, defaultSkip := parseDefault(defaultValue)
	if len(sourceValues) == 0 {
		if defaultSkip {
			return 0, true
		}
		return fallback, false
	}
	picked := make([]int, 0, len(sourceValues))
	for _, value := range sourceValues {
		preset, skip, follow := parseChoice(value)
		if follow {
			if defaultSkip {
				continue
			}
			preset = fallback
		}
		if skip {
			continue
		}
		picked = append(picked, preset)
	}
	if len(picked) == 0 {
		return 0, true
	}
	return higherPreset(picked), false
}

func parseDefault(value string) (int, bool) {
	preset, skip, follow := parseChoice(value)
	if skip {
		return 0, true
	}
	if follow {
		return QualityBest, false
	}
	return preset, false
}

func parseChoice(value string) (preset int, skip bool, follow bool) {
	switch strings.TrimSpace(value) {
	case "", "0":
		return 0, false, true
	case "skip":
		return 0, true, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < QualityBest || n > QualityHighAudio {
		return 0, false, true
	}
	return n, false, false
}

func higherPreset(presets []int) int {
	seen := map[int]struct{}{}
	for _, preset := range presets {
		if preset < QualityBest || preset > QualityHighAudio {
			continue
		}
		seen[preset] = struct{}{}
	}
	if _, ok := seen[QualityBest]; ok {
		return QualityBest
	}
	_, mid := seen[Quality1080]
	_, low := seen[QualityLow]
	_, highVideo := seen[QualityHighVideo]
	_, highAudio := seen[QualityHighAudio]
	if (highVideo && highAudio) || ((highVideo || highAudio) && (mid || low)) {
		// 最高画面最低音频和最低画面最高音频是两种取舍，比不出谁更高。
		// 它们互相冲突，或和 1080P、最低码率放在一起时，没有唯一更高的一档，直接指定最高清晰度。
		return QualityBest
	}
	if highVideo {
		return QualityHighVideo
	}
	if highAudio {
		return QualityHighAudio
	}
	if mid {
		return Quality1080
	}
	if low {
		return QualityLow
	}
	return QualityBest
}

func settingText(value any) string {
	switch n := value.(type) {
	case string:
		return n
	case float64:
		return strconv.FormatInt(int64(n), 10)
	default:
		return fmt.Sprint(value)
	}
}

func qualityPresetName(preset int) string {
	switch preset {
	case QualityBest:
		return "最高清晰度"
	case Quality1080:
		return "1080P"
	case QualityLow:
		return "最低码率"
	case QualityHighVideo:
		return "最高画面，最低音频"
	case QualityHighAudio:
		return "最低画面，最高音频"
	default:
		return "最高清晰度"
	}
}

// pickStreams 按档位从 playurl 的画面和声音里各选一条。
func pickStreams(videos, audios []stream, preset int) (stream, stream, error) {
	videos = withURL(videos)
	audios = withURL(audios)
	if len(videos) == 0 || len(audios) == 0 {
		return stream{}, stream{}, errors.New("播放地址里没有画面或声音")
	}
	switch preset {
	case Quality1080:
		return pickVideo1080(videos), pickAudio(audios, true), nil
	case QualityLow:
		return pickVideoBandwidth(videos, false), pickAudio(audios, false), nil
	case QualityHighVideo:
		return pickVideoBest(videos), pickAudio(audios, false), nil
	case QualityHighAudio:
		return pickVideoBandwidth(videos, false), pickAudio(audios, true), nil
	default:
		return pickVideoBest(videos), pickAudio(audios, true), nil
	}
}

func pickVideoBest(videos []stream) stream {
	return pickAtQuality(videos, maxID(videos))
}

func pickVideo1080(videos []stream) stream {
	if chosen, ok := firstAt(videos, qn1080); ok {
		return pickAtQuality(videos, chosen.ID)
	}
	below := maxIDBelow(videos, qn1080)
	if below > 0 {
		return pickAtQuality(videos, below)
	}
	return pickVideoBest(videos)
}

func pickAtQuality(videos []stream, id int) stream {
	var hevc, other []stream
	for _, item := range videos {
		if item.ID != id {
			continue
		}
		if item.CodecID == codecHEVC {
			hevc = append(hevc, item)
			continue
		}
		other = append(other, item)
	}
	if len(hevc) > 0 {
		return maxBandwidth(hevc)
	}
	return maxBandwidth(other)
}

func pickVideoBandwidth(videos []stream, highest bool) stream {
	chosen := videos[0]
	for _, item := range videos[1:] {
		if highest {
			if betterBandwidth(item, chosen) {
				chosen = item
			}
			continue
		}
		if item.Bandwidth < chosen.Bandwidth || (item.Bandwidth == chosen.Bandwidth && item.ID < chosen.ID) {
			chosen = item
		}
	}
	return chosen
}

func pickAudio(audios []stream, highest bool) stream {
	return pickVideoBandwidth(audios, highest)
}

func betterBandwidth(item, chosen stream) bool {
	if item.Bandwidth != chosen.Bandwidth {
		return item.Bandwidth > chosen.Bandwidth
	}
	return item.ID > chosen.ID
}

func maxBandwidth(items []stream) stream {
	return pickVideoBandwidth(items, true)
}

func maxID(items []stream) int {
	id := items[0].ID
	for _, item := range items[1:] {
		if item.ID > id {
			id = item.ID
		}
	}
	return id
}

func maxIDBelow(items []stream, limit int) int {
	id := 0
	for _, item := range items {
		if item.ID < limit && item.ID > id {
			id = item.ID
		}
	}
	return id
}

func firstAt(items []stream, id int) (stream, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return stream{}, false
}

func withURL(items []stream) []stream {
	out := make([]stream, 0, len(items))
	for _, item := range items {
		if len(item.URLs) == 0 || item.URLs[0] == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}
