package syncbili

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func TestSyncVideosUpsertsAndDetaches(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 29, 4, 0, 0, 0, time.UTC)
	folder := model.FavoriteList{
		ID: 7, Title: "夹", Cover: "c", MediaCount: 2, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	deletedAt := dbtime.DateTime{Time: now}
	old := sampleVideo(11, "旧标题")
	old.CacheImage = "keep"
	old.VideoDownloadedNum = 3
	old.Frozen = 1
	skipped := sampleVideo(13, "已失效视频")
	skipped.Invalid = 1
	removed := sampleVideo(14, "要解开")
	deleted := sampleVideo(15, "已删除")
	deleted.DeletedAt = &deletedAt
	for _, row := range []model.Video{old, skipped, removed, deleted} {
		if err := conn.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, videoID := range []int64{11, 13, 14, 15} {
		link := model.FavoriteListVideo{FavoriteListID: 7, VideoID: videoID}
		if err := conn.Create(&link).Error; err != nil {
			t.Fatal(err)
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("pn") {
		case "1":
			_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + mediaJSON(11, "新标题", 0) + `],"has_more":true}}`))
		case "2":
			_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + mediaJSON(12, "新建", 0) + `,` + mediaJSON(13, "已失效视频", 0) + `,` + mediaJSON(15, "删了也别改", 0) + `],"has_more":false}}`))
		default:
			t.Errorf("page %s", r.URL.Query().Get("pn"))
		}
	}))
	defer server.Close()

	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	id := int64(7)
	result, err := newTestSync(conn, client, nil, now, 1).FavoriteVideos(context.Background(), id, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Folders != 1 || result.Videos != 2 {
		t.Fatalf("result = %+v", result)
	}

	var updated model.Video
	if err := conn.First(&updated, 11).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Title != "新标题" || updated.CacheImage != "keep" || updated.VideoDownloadedNum != 3 || updated.Frozen != 0 {
		t.Fatalf("updated = %+v", updated)
	}
	if updated.UpperID == nil || *updated.UpperID != 9 || updated.FavTime == nil || updated.FavTime.Unix() != 1743737512 {
		t.Fatalf("mapped = %+v", updated)
	}

	var created model.Video
	if err := conn.First(&created, 12).Error; err != nil {
		t.Fatal(err)
	}
	if created.Link != "https://www.bilibili.com/video/BV12" || created.CacheImage != "" {
		t.Fatalf("created = %+v", created)
	}

	var stillSkipped model.Video
	if err := conn.First(&stillSkipped, 13).Error; err != nil {
		t.Fatal(err)
	}
	if stillSkipped.Title != "已失效视频" {
		t.Fatalf("skipped title changed: %s", stillSkipped.Title)
	}
	var stillDeleted model.Video
	if err := conn.First(&stillDeleted, 15).Error; err != nil {
		t.Fatal(err)
	}
	if stillDeleted.Title != "已删除" {
		t.Fatalf("deleted title changed: %s", stillDeleted.Title)
	}

	var links []model.FavoriteListVideo
	if err := conn.Where("favorite_list_id = ?", 7).Order("video_id").Find(&links).Error; err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].VideoID != 11 || links[1].VideoID != 12 {
		t.Fatalf("links = %+v", links)
	}
}

func TestSyncVideosKeepsDownloadedInvalid(t *testing.T) {
	conn := openTestDB(t)
	folder := model.FavoriteList{
		ID: 7, Title: "夹", Cover: "c", MediaCount: 1, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	row := sampleVideo(11, "本地标题")
	row.VideoDownloadedNum = 1
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: 7, VideoID: 11}).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + mediaJSON(11, "已失效视频", 1) + `],"has_more":false}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	id := int64(7)
	result, err := newTestSync(conn, client, nil, time.Now(), 20).FavoriteVideos(context.Background(), id, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Videos != 0 {
		t.Fatalf("result = %+v", result)
	}
	var saved model.Video
	if err := conn.First(&saved, 11).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Title != "本地标题" || saved.VideoDownloadedNum != 1 {
		t.Fatalf("saved = %+v", saved)
	}
	var links int64
	if err := conn.Model(&model.FavoriteListVideo{}).Where("favorite_list_id = ?", 7).Count(&links).Error; err != nil {
		t.Fatal(err)
	}
	if links != 1 {
		t.Fatalf("links = %d", links)
	}
}

func TestSyncVideosStopsOnRiskControl(t *testing.T) {
	conn := openTestDB(t)
	for _, id := range []int64{1, 2} {
		folder := model.FavoriteList{
			ID: id, Title: "夹", Cover: "c", MediaCount: 0, CacheImage: "",
			Ctime: stamp(1), Mtime: stamp(2),
		}
		if err := conn.Create(&folder).Error; err != nil {
			t.Fatal(err)
		}
	}
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Query().Get("media_id") == "1" {
			_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + mediaJSON(21, "甲", 0) + `],"has_more":false}}`))
			return
		}
		w.WriteHeader(http.StatusPreconditionFailed)
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	result, err := newTestSync(conn, client, nil, time.Now(), 20).FavoriteVideos(context.Background(), 0, 0)
	if _, ok := err.(*SuspendedError); !ok {
		t.Fatalf("err = %v", err)
	}
	if result.Folders != 1 || result.Videos != 1 || hits.Load() != 2 {
		t.Fatalf("result=%+v hits=%d", result, hits.Load())
	}
	var rows int64
	if err := conn.Model(&model.Video{}).Count(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("videos = %d", rows)
	}
}

func TestSyncVideosOnePageDoesNotDetach(t *testing.T) {
	conn := openTestDB(t)
	folder := model.FavoriteList{
		ID: 7, Title: "夹", Cover: "c", MediaCount: 1, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	old := sampleVideo(14, "留着")
	if err := conn.Create(&old).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: 7, VideoID: 14}).Error; err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Query().Get("pn") != "2" {
			t.Errorf("pn = %s", r.URL.Query().Get("pn"))
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + mediaJSON(11, "第二页", 0) + `],"has_more":true}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	result, err := newTestSync(conn, client, nil, time.Now(), 20).FavoriteVideos(context.Background(), 7, 2)
	if err != nil {
		t.Fatal(err)
	}
	if result.Videos != 1 || hits.Load() != 1 {
		t.Fatalf("result=%+v hits=%d", result, hits.Load())
	}
	var links []model.FavoriteListVideo
	if err := conn.Where("favorite_list_id = ?", 7).Order("video_id").Find(&links).Error; err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].VideoID != 11 || links[1].VideoID != 14 {
		t.Fatalf("links = %+v", links)
	}
}

func TestSyncVideosRequiresLocalFolder(t *testing.T) {
	conn := openTestDB(t)
	id := int64(404)
	_, err := newTestSync(conn, nil, nil, time.Now(), 20).FavoriteVideos(context.Background(), id, 0)
	if err == nil {
		t.Fatal("expected missing folder")
	}
}

func TestSyncPageEnqueuesVideosWithoutParts(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC)
	folder := model.FavoriteList{
		ID: 7, Title: "夹", Cover: "c", MediaCount: 2, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	ready := sampleVideo(11, "已有分P")
	if err := conn.Create(&ready).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: ready.ID, Cid: 8, Page: 1, From: "v", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", FirstFrame: "",
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + mediaJSON(11, "已有分P", 0) + `,` + mediaJSON(12, "新建", 0) + `],"has_more":false}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	syncer := newTestSync(conn, client, nil, now, 20)
	if _, err := syncer.FavoriteVideos(context.Background(), folder.ID, 1); err != nil {
		t.Fatal(err)
	}
	var jobs []model.Job
	if err := conn.Where("type = ?", TypeSyncVideo).Order("id").Find(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].UniqueKey != "sync:video:BV12" || jobs[0].Status != "pending" {
		t.Fatalf("jobs = %+v", jobs)
	}
	if _, err := syncer.FavoriteVideos(context.Background(), folder.ID, 1); err != nil {
		t.Fatal(err)
	}
	var again int64
	if err := conn.Model(&model.Job{}).Where("type = ?", TypeSyncVideo).Count(&again).Error; err != nil {
		t.Fatal(err)
	}
	if again != 1 {
		t.Fatalf("jobs = %d", again)
	}
	if err := conn.Where("1 = 1").Delete(&model.Job{}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := syncer.FavoriteVideos(context.Background(), folder.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := conn.Model(&model.Job{}).Where("unique_key = ?", "sync:video:BV12").Count(&again).Error; err != nil {
		t.Fatal(err)
	}
	if again != 1 {
		t.Fatalf("after clear = %d", again)
	}
}

func TestSyncVideosStoresDecodedAid(t *testing.T) {
	conn := openTestDB(t)
	folder := model.FavoriteList{
		ID: 7, Title: "动画", Cover: "c", MediaCount: 1, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	aid, ok := bilibili.BvToAv("BV1134y1g7Vq")
	if !ok || aid != 832246077 {
		t.Fatalf("aid = %d ok=%v", aid, ok)
	}
	upperID := int64(42)
	ready := sampleVideo(aid, "铃芽之旅")
	ready.Bvid = "BV1134y1g7Vq"
	ready.Page = 1
	ready.UpperID = &upperID
	ready.VideoDownloadedNum = 1
	if err := conn.Create(&ready).Error; err != nil {
		t.Fatal(err)
	}
	part := model.VideoPart{
		VideoID: aid, Cid: 8, Page: 1, From: "v", Part: "正片", Duration: 3,
		Vid: "", Weblink: "", FirstFrame: "",
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	stub := sampleVideo(780019, "空壳")
	stub.Bvid = "BV1134y1g7Vq"
	stub.Type = 24
	if err := conn.Create(&stub).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: 7, VideoID: stub.ID}).Error; err != nil {
		t.Fatal(err)
	}

	body := `{"id":780019,"type":24,"title":"铃芽之旅","cover":"c","intro":"原版","page":0,"duration":7283,"upper":{"mid":0},"attr":0,"link":"https://www.bilibili.com/bangumi/play/ep780019","bvid":"BV1134y1g7Vq","pubtime":1697853602,"fav_time":1795367267}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + body + `],"has_more":false}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	if _, err := newTestSync(conn, client, nil, time.Now(), 20).FavoriteVideos(context.Background(), 7, 0); err != nil {
		t.Fatal(err)
	}

	var saved model.Video
	if err := conn.First(&saved, aid).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Title != "铃芽之旅" || saved.Page != 1 || saved.UpperID == nil || *saved.UpperID != 42 || saved.VideoDownloadedNum != 1 {
		t.Fatalf("saved = %+v", saved)
	}
	var stillStub model.Video
	if err := conn.First(&stillStub, 780019).Error; err != nil {
		t.Fatal(err)
	}
	var links []model.FavoriteListVideo
	if err := conn.Where("favorite_list_id = ?", 7).Find(&links).Error; err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].VideoID != aid {
		t.Fatalf("links = %+v", links)
	}
	var jobs int64
	if err := conn.Model(&model.Job{}).Where("type = ?", TypeSyncVideo).Count(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if jobs != 0 {
		t.Fatalf("detail jobs = %d", jobs)
	}
}

func TestSyncVideosSkipsAudioWhenAidTaken(t *testing.T) {
	conn := openTestDB(t)
	folder := model.FavoriteList{
		ID: 7, Title: "夹", Cover: "c", MediaCount: 2, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	aid, ok := bilibili.BvToAv("BV19x411F7m7")
	if !ok || aid != 457094 {
		t.Fatalf("aid = %d ok=%v", aid, ok)
	}
	video := sampleVideo(aid, "《车姬》")
	video.Bvid = "BV19x411F7m7"
	video.VideoDownloadedNum = 1
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	videoJSON := `{"id":457094,"type":2,"title":"《车姬》","cover":"c","intro":"i","page":1,"duration":175,"upper":{"mid":9},"attr":0,"bvid":"BV19x411F7m7","pubtime":1359640344,"fav_time":1795370533}`
	audioJSON := `{"id":457094,"type":12,"title":"Kawaii Time","cover":"a","intro":"歌","page":1,"duration":239,"upper":{"mid":8},"attr":0,"bvid":"BV19x411F7m7","pubtime":1534218133,"fav_time":1795370509}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + audioJSON + `,` + videoJSON + `],"has_more":false}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	result, err := newTestSync(conn, client, nil, time.Now(), 20).FavoriteVideos(context.Background(), 7, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Videos != 1 {
		t.Fatalf("result = %+v", result)
	}
	var saved model.Video
	if err := conn.First(&saved, aid).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Title != "《车姬》" || saved.Type != 2 || saved.VideoDownloadedNum != 1 {
		t.Fatalf("saved = %+v", saved)
	}
	var audioJobs int64
	if err := conn.Model(&model.Job{}).Where("type = ?", TypeSyncAudio).Count(&audioJobs).Error; err != nil {
		t.Fatal(err)
	}
	if audioJobs != 0 {
		t.Fatalf("audio jobs = %d", audioJobs)
	}
}

func TestSyncVideosKeepsEarlierFavoriteWhenAidCollides(t *testing.T) {
	conn := openTestDB(t)
	folder := model.FavoriteList{
		ID: 7, Title: "夹", Cover: "c", MediaCount: 2, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	videoJSON := `{"id":457094,"type":2,"title":"《车姬》","cover":"c","intro":"i","page":1,"duration":175,"upper":{"mid":9},"attr":0,"bvid":"BV19x411F7m7","pubtime":1359640344,"fav_time":1795370533}`
	audioJSON := `{"id":457094,"type":12,"title":"Kawaii Time","cover":"a","intro":"歌","page":1,"duration":239,"upper":{"mid":8},"attr":0,"bvid":"BV19x411F7m7","pubtime":1534218133,"fav_time":1795370509}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[` + videoJSON + `,` + audioJSON + `],"has_more":false}}`))
	}))
	defer server.Close()
	client := &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()}
	result, err := newTestSync(conn, client, nil, time.Now(), 20).FavoriteVideos(context.Background(), 7, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Videos != 1 {
		t.Fatalf("result = %+v", result)
	}
	var saved model.Video
	if err := conn.First(&saved, 457094).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Title != "Kawaii Time" || saved.Type != 12 {
		t.Fatalf("saved = %+v", saved)
	}
	var videoJobs int64
	if err := conn.Model(&model.Job{}).Where("type = ?", TypeSyncVideo).Count(&videoJobs).Error; err != nil {
		t.Fatal(err)
	}
	if videoJobs != 0 {
		t.Fatalf("video jobs = %d", videoJobs)
	}
}

func sampleVideo(id int64, title string) model.Video {
	return model.Video{
		ID: id, Link: "old", Title: title, Intro: "i", Cover: "c", Bvid: "BV",
		Pubtime: stamp(1), CacheImage: "", Page: 1, Type: 2,
	}
}

func mediaJSON(id int64, title string, attr int) string {
	return fmt.Sprintf(`{"id":%d,"type":2,"title":"%s","cover":"c","intro":"i","page":1,"duration":3,"upper":{"mid":9},"attr":%d,"bvid":"BV%d","pubtime":1499900654,"fav_time":1743737512}`, id, title, attr, id)
}
