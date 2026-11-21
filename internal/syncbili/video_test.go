package syncbili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
	"mybili/internal/schedule"
)

func TestSyncVideoWritesDetailAndParts(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 29, 4, 0, 0, 0, time.UTC)
	fav := dbtime.DateTime{Time: time.Unix(1743737512, 0)}
	existing := sampleVideo(11, "旧标题")
	existing.CacheImage = "keep"
	existing.VideoDownloadedNum = 2
	existing.FavTime = &fav
	existing.Bvid = "BVold"
	if err := conn.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	downloadPath := "video/11.mp4"
	oldPart := model.VideoPart{
		VideoID: 11, Cid: 100, Page: 1, From: "old", Part: "旧分P", Duration: 1,
		Vid: "v", Weblink: "", FirstFrame: "", VideoDownloadPath: &downloadPath,
	}
	extra := model.VideoPart{
		VideoID: 11, Cid: 999, Page: 9, From: "old", Part: "多出来的", Duration: 1,
		Vid: "", Weblink: "", FirstFrame: "",
	}
	if err := conn.Create(&oldPart).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&extra).Error; err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("aid") != "11" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"aid":11,"bvid":"BV11","title":"新标题","desc":"介绍","pic":"cover","pubdate":1499900654,"state":0,"owner":{"mid":9,"name":"UP","face":"face"},"pages":[{"cid":100,"page":1,"from":"vupload","part":"新分P","duration":30,"vid":"","weblink":"","dimension":{"width":1280,"height":720,"rotate":0},"first_frame":"frame"},{"cid":101,"page":2,"from":"vupload","part":"P2","duration":10}]}}`))
	}))
	defer server.Close()

	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	if err := newTestSync(conn, client, nil, now, 20).Video(context.Background(), 11, ""); err != nil {
		t.Fatal(err)
	}

	var video model.Video
	if err := conn.First(&video, 11).Error; err != nil {
		t.Fatal(err)
	}
	if video.Title != "新标题" || video.Intro != "介绍" || video.Link != "bilibili://video/11" || video.Page != 2 {
		t.Fatalf("video = %+v", video)
	}
	if video.CacheImage != "keep" || video.VideoDownloadedNum != 2 || video.FavTime == nil || video.FavTime.Unix() != 1743737512 {
		t.Fatalf("kept fields = %+v", video)
	}
	if video.UpperID == nil || *video.UpperID != 9 || video.Invalid != 0 || video.Frozen != 0 {
		t.Fatalf("flags = %+v", video)
	}

	var upper model.Upper
	if err := conn.Where("mid = ?", 9).First(&upper).Error; err != nil {
		t.Fatal(err)
	}
	if upper.Name != "UP" || upper.Face != "face" {
		t.Fatalf("upper = %+v", upper)
	}

	var parts []model.VideoPart
	if err := conn.Where("video_id = ?", 11).Order("cid").Find(&parts).Error; err != nil {
		t.Fatal(err)
	}
	if len(parts) != 3 {
		t.Fatalf("parts = %+v", parts)
	}
	if parts[0].Cid != 100 || parts[0].Part != "新分P" || parts[0].VideoDownloadPath == nil || *parts[0].VideoDownloadPath != downloadPath {
		t.Fatalf("updated part = %+v", parts[0])
	}
	if parts[1].Cid != 101 || parts[1].Part != "P2" {
		t.Fatalf("new part = %+v", parts[1])
	}
	if parts[2].Cid != 999 || parts[2].Part != "多出来的" {
		t.Fatalf("extra part = %+v", parts[2])
	}
}

func TestSyncVideoInvalidStateSkipsParts(t *testing.T) {
	conn := openTestDB(t)
	existing := sampleVideo(11, "旧标题")
	existing.VideoDownloadedNum = 1
	if err := conn.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"aid":11,"bvid":"BV11","title":"失效了","desc":"","pic":"c","pubdate":1,"state":-1,"owner":{"mid":3,"name":"已注销","face":"f"},"pages":[{"cid":5,"page":1,"part":"不该写入","duration":1}]}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	if err := newTestSync(conn, client, nil, time.Now(), 20).Video(context.Background(), 11, ""); err != nil {
		t.Fatal(err)
	}
	var video model.Video
	if err := conn.First(&video, 11).Error; err != nil {
		t.Fatal(err)
	}
	if video.Title != "失效了" || video.Invalid != 1 || video.Frozen != 1 {
		t.Fatalf("video = %+v", video)
	}
	var parts int64
	if err := conn.Model(&model.VideoPart{}).Where("video_id = ?", 11).Count(&parts).Error; err != nil {
		t.Fatal(err)
	}
	if parts != 0 {
		t.Fatalf("parts = %d", parts)
	}
	var upper model.Upper
	if err := conn.Where("mid = ?", 3).First(&upper).Error; err != nil {
		t.Fatal(err)
	}
	if upper.Name != "已注销 (注销)" {
		t.Fatalf("upper = %+v", upper)
	}
}

func TestSyncVideoUnavailableMarksDownloadedFrozen(t *testing.T) {
	conn := openTestDB(t)
	existing := sampleVideo(11, "本地标题")
	existing.VideoDownloadedNum = 1
	if err := conn.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":62002,"message":"稿件不可见","data":null}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	if err := newTestSync(conn, client, nil, time.Now(), 20).Video(context.Background(), 11, ""); err != nil {
		t.Fatal(err)
	}
	var video model.Video
	if err := conn.First(&video, 11).Error; err != nil {
		t.Fatal(err)
	}
	if video.Title != "本地标题" || video.Invalid != 1 || video.Frozen != 1 {
		t.Fatalf("video = %+v", video)
	}
}

func TestUnavailableNotifiesFreezeOnce(t *testing.T) {
	conn := openTestDB(t)
	existing := sampleVideo(11, "本地标题")
	existing.VideoDownloadedNum = 1
	if err := conn.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":62002,"message":"稿件不可见","data":null}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	var notified []int64
	syncer := newTestSync(conn, client, nil, time.Now(), 20)
	syncer.NotifyFrozen = func(_ context.Context, id int64) {
		notified = append(notified, id)
	}
	if err := syncer.Video(context.Background(), 11, ""); err != nil {
		t.Fatal(err)
	}
	if err := syncer.Video(context.Background(), 11, ""); err != nil {
		t.Fatal(err)
	}
	if len(notified) != 1 || notified[0] != 11 {
		t.Fatalf("notified = %v", notified)
	}
}

func TestMissingPartsSkipsFilledAndAudio(t *testing.T) {
	conn := openTestDB(t)
	plain := sampleVideo(1, "没分P")
	filled := sampleVideo(2, "有分P")
	audio := sampleVideo(3, "音频")
	audio.Type = 12
	for _, row := range []model.Video{plain, filled, audio} {
		if err := conn.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	part := model.VideoPart{VideoID: 2, Cid: 8, Page: 1, From: "v", Part: "P", Duration: 1, Vid: "", Weblink: "", FirstFrame: ""}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Query().Get("aid") != "1" {
			t.Errorf("aid = %s", r.URL.Query().Get("aid"))
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"aid":1,"bvid":"BV1","title":"没分P","desc":"","pic":"c","pubdate":1,"state":0,"pages":[{"cid":7,"page":1,"from":"vupload","part":"P1","duration":4}]}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	count, err := newTestSync(conn, client, nil, time.Now(), 20).MissingParts(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || hits.Load() != 1 {
		t.Fatalf("count=%d hits=%d", count, hits.Load())
	}
}

func TestSyncVideoStopsOnRiskControl(t *testing.T) {
	conn := openTestDB(t)
	for _, id := range []int64{1, 2} {
		row := sampleVideo(id, "视频")
		if err := conn.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Query().Get("aid") == "1" {
			_, _ = w.Write([]byte(`{"code":0,"data":{"aid":1,"bvid":"BV1","title":"甲","desc":"","pic":"c","pubdate":1,"state":0,"pages":[{"cid":7,"page":1,"part":"P","duration":1}]}}`))
			return
		}
		w.WriteHeader(http.StatusPreconditionFailed)
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	count, err := newTestSync(conn, client, nil, time.Now(), 20).MissingParts(context.Background(), 10)
	if _, ok := err.(*SuspendedError); !ok {
		t.Fatalf("err = %v", err)
	}
	if count != 1 || hits.Load() != 2 {
		t.Fatalf("count=%d hits=%d", count, hits.Load())
	}
}

func TestMissingPartSpecsSkipsExcludedAndRecent(t *testing.T) {
	conn := openTestDB(t)
	rows := []model.Video{
		partCandidate(1, "BV1", 0, 0, 0, 2, nil),
		partCandidate(2, "BV2", 0, 0, 0, 2, nil),
		partCandidate(3, "BV3", 0, 0, 0, 12, nil),
		partCandidate(4, "BV4", 1, 0, 0, 2, nil),
		partCandidate(5, "BV5", 0, 1, 0, 2, nil),
		partCandidate(6, "BV6", 0, 0, 1, 2, nil),
		partCandidate(7, "BV7", 0, 0, 0, 2, &[]dbtime.DateTime{stamp(1)}[0]),
		partCandidate(8, "", 0, 0, 0, 2, nil),
		partCandidate(9, "BV9", 0, 0, 0, 2, nil),
		partCandidate(10, "BV10", 0, 0, 0, 2, nil),
		partCandidate(11, "BV11", 0, 0, 0, 2, nil),
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.VideoPart{VideoID: 2, Cid: 8, Page: 1, From: "v", Part: "P", Duration: 1, Vid: "", Weblink: "", FirstFrame: ""}).Error; err != nil {
		t.Fatal(err)
	}
	links := []model.FavoriteListVideo{
		{FavoriteListID: 9, VideoID: 9},
		{FavoriteListID: 9, VideoID: 10},
		{FavoriteListID: 10, VideoID: 10},
	}
	if err := conn.Create(&links).Error; err != nil {
		t.Fatal(err)
	}
	raw := `{"enabled":true,"selected":[9]}`
	if err := conn.Create(&model.Setting{Name: "fav_exclude", Value: &raw}).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 31, 5, 0, 0, 0, time.Local)
	specs := MissingPartSpecs(conn)
	if !specs[0].Due(now) || specs[0].Due(now.Add(time.Minute)) {
		t.Fatal("due window")
	}
	drafts, err := specs[0].Tasks(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if keys := draftKeys(drafts); len(keys) != 4 || keys[0] != "sync:video:BV1" || keys[1] != "sync:video:BV10" || keys[2] != "sync:video:BV11" || keys[3] != "sync:audio:3" {
		t.Fatalf("keys = %v", keys)
	}
	if !drafts[0].ReopenAfter.Equal(now.Add(-24 * time.Hour)) {
		t.Fatalf("reopen = %s", drafts[0].ReopenAfter)
	}
}

func partCandidate(id int64, bvid string, invalid, frozen, downloaded, kind int, deleted *dbtime.DateTime) model.Video {
	row := sampleVideo(id, bvid)
	row.Bvid = bvid
	row.Invalid = invalid
	row.Frozen = frozen
	row.VideoDownloadedNum = downloaded
	row.Type = kind
	row.DeletedAt = deleted
	return row
}

func draftKeys(drafts []schedule.Draft) []string {
	keys := make([]string, 0, len(drafts))
	for _, draft := range drafts {
		keys = append(keys, draft.UniqueKey)
	}
	return keys
}
