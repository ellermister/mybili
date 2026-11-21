package readable

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/migrate"
	"mybili/internal/model"
)

func TestSanitizeAndEpisode(t *testing.T) {
	if sanitize(`a/b:c*?"<>|#%&{}`+"`$!@+=") != "a b c" {
		t.Fatalf("sanitize = %q", sanitize(`a/b:c*?"<>|#%&{}`+"$!@+=`"))
	}
	if sanitize("   ") != "untitled" {
		t.Fatal(sanitize("   "))
	}
	long := ""
	for i := 0; i < 90; i++ {
		long += "字"
	}
	if got := []rune(sanitize(long)); len(got) != 80 {
		t.Fatalf("len = %d", len(got))
	}
	video := model.Video{Title: "正片", Bvid: "BV1", Pubtime: dbtime.DateTime{Time: time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)}}
	if episodeFileName(video, model.VideoPart{Part: "第一集", Page: 3}) != "S01E03 - 第一集.mp4" {
		t.Fatal(episodeFileName(video, model.VideoPart{Part: "第一集", Page: 3}))
	}
	if episodeFileName(video, model.VideoPart{Part: "正片", Page: 3}) != "S01E03.mp4" {
		t.Fatal(episodeFileName(video, model.VideoPart{Part: "正片", Page: 3}))
	}
	if seriesFolder(video) != "正片 (2026) [BV1]" {
		t.Fatal(seriesFolder(video))
	}
	if !matchesEpisode("S01E01 - 标题.mp4", "S01E01") || matchesEpisode("S01E010.mp4", "S01E01") {
		t.Fatal("episode match")
	}
}

func TestGenerateLinksMoviesAndClearsOldDirs(t *testing.T) {
	conn := openReadableDB(t)
	data := t.TempDir()
	root := filepath.Join(data, "public", "medias")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "旧目录"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "旧目录", "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "keep.txt"), []byte("k"), 0o644); err != nil {
		t.Fatal(err)
	}
	stamp := dbtime.DateTime{Time: time.Date(2026, 10, 31, 12, 0, 0, 0, dbtime.Location())}
	video := model.Video{
		Link: "l", Title: "标题/一", Intro: "", Cover: "c", Bvid: "BV1",
		Pubtime: stamp, Type: 2, VideoDownloadedNum: 1, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	stored := "/storage/videos/" + itoa(video.ID) + ".8.mp4"
	part := model.VideoPart{
		VideoID: video.ID, Cid: 8, Page: 1, From: "v", Part: "P1", Duration: 3,
		Vid: "", Weblink: "", FirstFrame: "", VideoDownloadPath: &stored,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(data, "public", "videos", itoa(video.ID)+".8.mp4")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	coverRel := "images/cover.jpg"
	coverAbs := filepath.Join(data, "public", coverRel)
	if err := os.MkdirAll(filepath.Dir(coverAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(coverAbs, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	cover := model.Cover{Type: "video", URL: "u", Filename: "cover.jpg", Path: coverRel, MimeType: "image/jpeg"}
	if err := conn.Create(&cover).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.Coverable{CoverID: cover.ID, CoverableID: video.ID, CoverableType: model.CoverableVideo}).Error; err != nil {
		t.Fatal(err)
	}
	fav := model.FavoriteList{Title: "收藏/夹", Cover: "c", Ctime: stamp, Mtime: stamp}
	if err := conn.Create(&fav).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: fav.ID, VideoID: video.ID}).Error; err != nil {
		t.Fatal(err)
	}
	sub := model.Subscription{Type: "up", Mid: "1", Name: "订阅甲"}
	if err := conn.Create(&sub).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.SubscriptionVideo{SubscriptionID: sub.ID, VideoID: video.ID, Bvid: "BV1"}).Error; err != nil {
		t.Fatal(err)
	}
	skipped := model.Video{
		Link: "l", Title: "没下", Intro: "", Cover: "c", Bvid: "BV2",
		Pubtime: stamp, Type: 2, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&skipped).Error; err != nil {
		t.Fatal(err)
	}

	n, err := Generate(context.Background(), conn, data, "")
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("links = %d", n)
	}
	if _, err := os.Stat(filepath.Join(root, "旧目录")); !os.IsNotExist(err) {
		t.Fatal("old dir remains")
	}
	if _, err := os.Stat(filepath.Join(root, "keep.txt")); err != nil {
		t.Fatal(err)
	}
	series := filepath.Join(root, "收藏 夹", "标题 一 (2026) [BV1]")
	movie := filepath.Join(series, "Season 01", "S01E01 - P1.mp4")
	if !os.SameFile(mustStat(t, source), mustStat(t, movie)) {
		t.Fatal("movie is not a hard link")
	}
	if _, err := os.Stat(filepath.Join(series, "poster.jpg")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "订阅甲", "标题 一 (2026) [BV1]", "poster.jpg")); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateSeriesAndRefusesMediaRoot(t *testing.T) {
	conn := openReadableDB(t)
	data := t.TempDir()
	root := filepath.Join(data, "public", "medias")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	stamp := dbtime.DateTime{Time: time.Date(2026, 10, 31, 12, 0, 0, 0, dbtime.Location())}
	video := model.Video{
		Link: "l", Title: "剧集", Intro: "", Cover: "c", Bvid: "BV1",
		Pubtime: stamp, Type: 2, VideoDownloadedNum: 2, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	writePart(t, conn, data, video.ID, 1, 11, "第一集")
	writePart(t, conn, data, video.ID, 2, 12, "第二集")
	fav := model.FavoriteList{Title: "片单", Cover: "c", Ctime: stamp, Mtime: stamp}
	if err := conn.Create(&fav).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: fav.ID, VideoID: video.ID}).Error; err != nil {
		t.Fatal(err)
	}
	n, err := Generate(context.Background(), conn, data, root)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("links = %d", n)
	}
	season := filepath.Join(root, "片单", "剧集 (2026) [BV1]", "Season 01")
	if _, err := os.Stat(filepath.Join(season, "S01E01 - 第一集.mp4")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(season, "S01E02 - 第二集.mp4")); err != nil {
		t.Fatal(err)
	}

	videos := filepath.Join(data, "public")
	if _, err := Generate(context.Background(), conn, data, videos); err == nil {
		t.Fatal("public dir was accepted")
	}
	if _, err := Generate(context.Background(), conn, data, filepath.Join(data, "missing")); err == nil {
		t.Fatal("missing dir was accepted")
	}
}

func TestLinkPartOrderAndUnlink(t *testing.T) {
	conn := openReadableDB(t)
	data := t.TempDir()
	root := filepath.Join(data, "public", "medias")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	stamp := dbtime.DateTime{Time: time.Date(2026, 10, 31, 12, 0, 0, 0, dbtime.Location())}
	video := model.Video{
		Link: "l", Title: "剧集", Intro: "", Cover: "c", Bvid: "BV1",
		Pubtime: stamp, Type: 2, VideoDownloadedNum: 1, CacheImage: "", CreatedAt: &stamp,
	}
	if err := conn.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	later := writePart(t, conn, data, video.ID, 3, 13, "第三集")
	first := writePart(t, conn, data, video.ID, 1, 11, "第一集")
	fav := model.FavoriteList{Title: "片单", Cover: "c", Ctime: stamp, Mtime: stamp}
	if err := conn.Create(&fav).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.FavoriteListVideo{FavoriteListID: fav.ID, VideoID: video.ID}).Error; err != nil {
		t.Fatal(err)
	}
	sub := model.Subscription{Type: "up", Mid: "1", Name: "订阅甲"}
	if err := conn.Create(&sub).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Create(&model.SubscriptionVideo{SubscriptionID: sub.ID, VideoID: video.ID, Bvid: "BV1"}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := LinkPart(context.Background(), conn, data, root, later); err != nil {
		t.Fatal(err)
	}
	third := filepath.Join(root, "片单", "剧集 (2026) [BV1]", "Season 01", "S01E03 - 第三集.mp4")
	if _, err := os.Stat(third); err != nil {
		t.Fatal(err)
	}
	if _, err := LinkPart(context.Background(), conn, data, root, later); err != nil {
		t.Fatal(err)
	}
	if _, err := LinkPart(context.Background(), conn, data, root, first); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "片单", "剧集 (2026) [BV1]", "Season 01", "S01E01 - 第一集.mp4")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(third); err != nil {
		t.Fatal("page 3 was renamed")
	}
	value := `"on"`
	if err := conn.Create(&model.Setting{Name: "human_readable_name_enabled", Value: &value}).Error; err != nil {
		t.Fatal(err)
	}
	if err := UnlinkFolder(context.Background(), conn, data, root, "片单", []int64{video.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "片单")); !os.IsNotExist(err) {
		t.Fatal("favorite dir remains")
	}
	subFile := filepath.Join(root, "订阅甲", "剧集 (2026) [BV1]", "Season 01", "S01E03 - 第三集.mp4")
	if _, err := os.Stat(subFile); err != nil {
		t.Fatal(err)
	}
	if err := UnlinkVideos(context.Background(), conn, data, root, []int64{video.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "订阅甲")); !os.IsNotExist(err) {
		t.Fatal("subscription dir remains")
	}
}

func TestEnqueuePartHonorsSwitch(t *testing.T) {
	conn := openReadableDB(t)
	if err := EnqueuePart(context.Background(), conn, nil, 4); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := conn.Model(&model.Job{}).Where("type = ?", TypeLink).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("jobs = %d", n)
	}
	value := `"on"`
	if err := conn.Create(&model.Setting{Name: "human_readable_name_enabled", Value: &value}).Error; err != nil {
		t.Fatal(err)
	}
	if err := EnqueuePart(context.Background(), conn, nil, 4); err != nil {
		t.Fatal(err)
	}
	if err := EnqueuePart(context.Background(), conn, nil, 4); err != nil {
		t.Fatal(err)
	}
	if err := conn.Model(&model.Job{}).Where("type = ?", TypeLink).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("jobs = %d", n)
	}
}

func writePart(t *testing.T, conn *gorm.DB, data string, videoID int64, page int, cid int64, name string) int64 {
	t.Helper()
	stored := "/storage/videos/" + itoa(videoID) + "." + itoa(cid) + ".mp4"
	part := model.VideoPart{
		VideoID: videoID, Cid: cid, Page: page, From: "v", Part: name, Duration: 3,
		Vid: "", Weblink: "", FirstFrame: "", VideoDownloadPath: &stored,
	}
	if err := conn.Create(&part).Error; err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(data, "public", "videos", itoa(videoID)+"."+itoa(cid)+".mp4")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("v"), 0o644); err != nil {
		t.Fatal(err)
	}
	return part.ID
}

func mustStat(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func openReadableDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "readable.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Auto(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}
