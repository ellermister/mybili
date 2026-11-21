package download

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func TestPickStreams(t *testing.T) {
	videos := []stream{
		{ID: 112, Bandwidth: 7505465, CodecID: codecAVC, URLs: []string{"h264-112"}},
		{ID: 112, Bandwidth: 3947970, CodecID: codecHEVC, URLs: []string{"h265-112"}},
		{ID: 80, Bandwidth: 4397545, CodecID: codecAVC, URLs: []string{"h264-80"}},
		{ID: 80, Bandwidth: 2087221, CodecID: codecHEVC, URLs: []string{"h265-80"}},
		{ID: 64, Bandwidth: 1828185, CodecID: codecAVC, URLs: []string{"h264-64"}},
	}
	audios := []stream{
		{ID: 30232, Bandwidth: 76610, URLs: []string{"a32"}},
		{ID: 30280, Bandwidth: 99596, URLs: []string{"a80"}},
		{ID: 30216, Bandwidth: 65827, URLs: []string{"a16"}},
	}
	video, audio, err := pickStreams(videos, audios, QualityBest)
	if err != nil || video.URLs[0] != "h265-112" || audio.URLs[0] != "a80" {
		t.Fatalf("best = %v %v %v", video.URLs, audio.URLs, err)
	}
	video, audio, err = pickStreams(videos, audios, Quality1080)
	if err != nil || video.URLs[0] != "h265-80" || audio.URLs[0] != "a80" {
		t.Fatalf("1080 = %v %v %v", video.URLs, audio.URLs, err)
	}
	video, audio, err = pickStreams(videos, audios, QualityLow)
	if err != nil || video.URLs[0] != "h264-64" || audio.URLs[0] != "a16" {
		t.Fatalf("low = %v %v %v", video.URLs, audio.URLs, err)
	}
	video, audio, err = pickStreams(videos, audios, QualityHighVideo)
	if err != nil || video.URLs[0] != "h265-112" || audio.URLs[0] != "a16" {
		t.Fatalf("high video = %v %v %v", video.URLs, audio.URLs, err)
	}
	video, audio, err = pickStreams(videos, audios, QualityHighAudio)
	if err != nil || video.URLs[0] != "h264-64" || audio.URLs[0] != "a80" {
		t.Fatalf("high audio = %v %v %v", video.URLs, audio.URLs, err)
	}
}

func TestVideoQualityReadsSetting(t *testing.T) {
	conn := openDownloadDB(t)
	if got := videoQuality(conn); got != QualityBest {
		t.Fatalf("missing = %d", got)
	}
	stamp := dbtime.DateTime{Time: dbtime.Now()}
	value := `"3"`
	if err := conn.Create(&model.Setting{Name: "video_quality", Value: &value, CreatedAt: &stamp}).Error; err != nil {
		t.Fatal(err)
	}
	if got := videoQuality(conn); got != QualityLow {
		t.Fatalf("stored = %d", got)
	}
	value = `"9"`
	if err := conn.Model(&model.Setting{}).Where("name = ?", "video_quality").Update("value", value).Error; err != nil {
		t.Fatal(err)
	}
	if got := videoQuality(conn); got != QualityBest {
		t.Fatalf("invalid = %d", got)
	}
}

func TestStreamLabelUsesNameAndMbps(t *testing.T) {
	video := stream{ID: 112, Bandwidth: 3947970, CodecID: codecHEVC}.label()
	if video != "1080P+ H.265 3.95Mbps" {
		t.Fatalf("video label = %s", video)
	}
	audio := stream{ID: 30280, Bandwidth: 99596, Codecs: "mp4a.40.2"}.label()
	if audio != "192K mp4a.40.2 0.10Mbps" {
		t.Fatalf("audio label = %s", audio)
	}
}

func TestPick1080FallsBack(t *testing.T) {
	videos := []stream{
		{ID: 112, Bandwidth: 10, CodecID: codecHEVC, URLs: []string{"h265-112"}},
		{ID: 64, Bandwidth: 4, CodecID: codecAVC, URLs: []string{"h264-64"}},
		{ID: 32, Bandwidth: 2, CodecID: codecHEVC, URLs: []string{"h265-32"}},
	}
	chosen := pickVideo1080(videos)
	if chosen.URLs[0] != "h264-64" {
		t.Fatalf("below 1080 = %v", chosen.URLs)
	}
	onlyHigh := []stream{{ID: 112, Bandwidth: 10, CodecID: codecAVC, URLs: []string{"h264-112"}}}
	chosen = pickVideo1080(onlyHigh)
	if chosen.URLs[0] != "h264-112" {
		t.Fatalf("only higher = %v", chosen.URLs)
	}
}

func TestSamplePlayURLPicks(t *testing.T) {
	body, err := os.ReadFile("../../docs/download-playurl.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Data struct {
			Dash dashBody `json:"dash"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	videos := streamsOf(payload.Data.Dash.Video)
	audios := streamsOf(payload.Data.Dash.Audio)
	video, audio, err := pickStreams(videos, audios, QualityBest)
	if err != nil || video.ID != 112 || video.CodecID != codecHEVC || audio.ID != 30280 {
		t.Fatalf("best = %+v %+v %v", video, audio, err)
	}
	video, audio, err = pickStreams(videos, audios, Quality1080)
	if err != nil || video.ID != 80 || video.CodecID != codecHEVC || audio.ID != 30280 {
		t.Fatalf("1080 = %+v %+v %v", video, audio, err)
	}
	video, audio, err = pickStreams(videos, audios, QualityLow)
	if err != nil || video.ID != 16 || audio.ID != 30216 {
		t.Fatalf("low = %+v %+v %v", video, audio, err)
	}
	video, audio, err = pickStreams(videos, audios, QualityHighVideo)
	if err != nil || video.ID != 112 || video.CodecID != codecHEVC || audio.ID != 30216 {
		t.Fatalf("high video = %+v %+v %v", video, audio, err)
	}
	video, audio, err = pickStreams(videos, audios, QualityHighAudio)
	if err != nil || video.ID != 16 || audio.ID != 30280 {
		t.Fatalf("high audio = %+v %+v %v", video, audio, err)
	}
}

func TestPlayQueryUsesAvidOrBvid(t *testing.T) {
	query, err := playQuery(9, 117, "BV1")
	if err != nil || !contains(query, "cid=9") || !contains(query, "avid=117") || contains(query, "bvid=") {
		t.Fatalf("avid query = %s %v", query, err)
	}
	query, err = playQuery(9, 0, "BV1")
	if err != nil || !contains(query, "bvid=BV1") || contains(query, "avid=") {
		t.Fatalf("bvid query = %s %v", query, err)
	}
	if _, err := playQuery(0, 1, "BV1"); err == nil {
		t.Fatal("cid is required")
	}
}

func TestChooseQuality(t *testing.T) {
	cases := []struct {
		name    string
		def     string
		sources []string
		preset  int
		skip    bool
	}{
		{name: "no source uses default", def: "2", preset: Quality1080},
		{name: "missing default is best", preset: QualityBest},
		{name: "default skip with no source", def: "skip", skip: true},
		{name: "follow default", def: "3", sources: []string{"", "0"}, preset: QualityLow},
		{name: "higher of 1080 and low", def: "3", sources: []string{"2", "3"}, preset: Quality1080},
		{name: "best beats the rest", def: "3", sources: []string{"1", "5", "skip"}, preset: QualityBest},
		{name: "skip loses to an explicit tier", def: "skip", sources: []string{"skip", "2"}, preset: Quality1080},
		{name: "all skip", def: "1", sources: []string{"skip", "skip"}, skip: true},
		{name: "only high video", def: "4", sources: []string{"4", "0"}, preset: QualityHighVideo},
		{name: "only high audio", def: "2", sources: []string{"5"}, preset: QualityHighAudio},
		{name: "mixed presets conflict", def: "2", sources: []string{"4", "5"}, preset: QualityBest},
		{name: "mixed preset with 1080 conflicts", def: "3", sources: []string{"4", "2"}, preset: QualityBest},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			preset, skip := chooseQuality(item.def, item.sources)
			if preset != item.preset || skip != item.skip {
				t.Fatalf("preset=%d skip=%v, want preset=%d skip=%v", preset, skip, item.preset, item.skip)
			}
		})
	}
}

func TestQualityForVideoUsesCurrentSources(t *testing.T) {
	conn := openDownloadDB(t)
	stamp := dbtime.DateTime{Time: dbtime.Now()}
	video := model.Video{
		Link: "https://www.bilibili.com/video/BV1", Title: "标题", Intro: "", Cover: "",
		Bvid: "BV1", Pubtime: stamp, Type: 2, CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&[]model.FavoriteListVideo{
		{FavoriteListID: 10, VideoID: video.ID},
		{FavoriteListID: 11, VideoID: video.ID},
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.SubscriptionVideo{SubscriptionID: 4, VideoID: video.ID, Bvid: "BV1"}).Error; err != nil {
		t.Fatal(err)
	}
	def := `"3"`
	sources := `{"10":"2","-4":"skip"}`
	if err := conn.Create(&[]model.Setting{
		{Name: "video_quality", Value: &def, CreatedAt: &stamp},
		{Name: "source_quality", Value: &sources, CreatedAt: &stamp},
	}).Error; err != nil {
		t.Fatal(err)
	}
	preset, skip, err := qualityForVideo(context.Background(), conn, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if skip || preset != Quality1080 {
		t.Fatalf("preset=%d skip=%v", preset, skip)
	}
	sources = `{"10":"skip","11":"skip","-4":"skip"}`
	if err := conn.Model(&model.Setting{}).Where("name = ?", "source_quality").Update("value", sources).Error; err != nil {
		t.Fatal(err)
	}
	_, skip, err = qualityForVideo(context.Background(), conn, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !skip {
		t.Fatal("every source is skip")
	}
}

func contains(text, part string) bool {
	return len(part) == 0 || (len(text) >= len(part) && (text == part || stringIndex(text, part) >= 0))
}

func stringIndex(text, part string) int {
	for i := 0; i+len(part) <= len(text); i++ {
		if text[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
