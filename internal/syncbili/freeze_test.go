package syncbili

import (
	"context"
	"testing"
	"time"

	"mybili/internal/dbtime"
	"mybili/internal/model"

	"gorm.io/gorm"
)

func TestFixFreezeRecountsOnlyZeroCounters(t *testing.T) {
	conn := openTestDB(t)
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.Local)
	downloaded := stamp(100)
	deleted := stamp(1)
	rows := []model.Video{
		freezeVideo(1, 1, 1, 0, 0, nil),
		freezeVideo(2, 1, 1, 0, 0, nil),
		freezeVideo(3, 1, 1, 2, 0, nil),
		freezeVideo(4, 0, 1, 0, 0, nil),
		freezeVideo(5, 1, 0, 0, 0, nil),
		freezeVideo(6, 1, 1, 0, 0, &deleted),
		freezeVideo(7, 1, 1, 0, 0, nil),
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	parts := []model.VideoPart{
		{VideoID: 1, Cid: 1, Page: 1, From: "v", Part: "P", Duration: 1, Vid: "", Weblink: "", FirstFrame: "", VideoDownloadedAt: &downloaded},
		{VideoID: 1, Cid: 2, Page: 2, From: "v", Part: "P2", Duration: 1, Vid: "", Weblink: "", FirstFrame: ""},
		{VideoID: 2, Cid: 3, Page: 1, From: "v", Part: "P", Duration: 1, Vid: "", Weblink: "", FirstFrame: ""},
	}
	if err := conn.Create(&parts).Error; err != nil {
		t.Fatal(err)
	}
	audio := model.AudioPart{VideoID: 7, Sid: 7, Duration: 1, AudioDownloadedAt: &downloaded}
	if err := conn.Create(&audio).Error; err != nil {
		t.Fatal(err)
	}
	syncer := newTestSync(conn, nil, nil, now, 40)
	fixed, err := syncer.FixFreeze(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if fixed != 3 {
		t.Fatalf("fixed = %d", fixed)
	}
	assertFreeze(t, conn, 1, 1, 1, 0)
	assertFreeze(t, conn, 2, 0, 0, 0)
	assertFreeze(t, conn, 3, 1, 2, 0)
	assertFreeze(t, conn, 4, 0, 0, 0)
	assertFreeze(t, conn, 5, 1, 0, 0)
	assertFreeze(t, conn, 6, 1, 0, 0)
	assertFreeze(t, conn, 7, 1, 0, 1)
}

func TestFixFreezeSpecsAtMidnight(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.Local)
	specs := FixFreezeSpecs()
	if !specs[0].Due(now) || specs[0].Due(now.Add(time.Minute)) || specs[0].Due(now.Add(4*time.Hour)) {
		t.Fatal("due window")
	}
	drafts, err := specs[0].Tasks(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(drafts) != 1 || drafts[0].UniqueKey != "sync:fix-freeze" || drafts[0].Type != TypeFixFreeze {
		t.Fatalf("draft = %+v", drafts)
	}
	if !drafts[0].ReopenAfter.Equal(now) {
		t.Fatalf("reopen = %s", drafts[0].ReopenAfter)
	}
}

func freezeVideo(id int64, frozen, invalid, videoNum, audioNum int, deleted *dbtime.DateTime) model.Video {
	row := sampleVideo(id, "冻结")
	row.Bvid = "BV"
	row.Frozen = frozen
	row.Invalid = invalid
	row.VideoDownloadedNum = videoNum
	row.AudioDownloadedNum = audioNum
	row.DeletedAt = deleted
	return row
}

func assertFreeze(t *testing.T, conn *gorm.DB, id int64, frozen int, videoNum, audioNum int) {
	t.Helper()
	var video model.Video
	if err := conn.First(&video, id).Error; err != nil {
		t.Fatal(err)
	}
	if video.Frozen != frozen || video.VideoDownloadedNum != videoNum || video.AudioDownloadedNum != audioNum {
		t.Fatalf("video %d = frozen %d video %d audio %d", id, video.Frozen, video.VideoDownloadedNum, video.AudioDownloadedNum)
	}
}
