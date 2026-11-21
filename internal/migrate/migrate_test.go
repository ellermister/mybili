package migrate

import (
	"path/filepath"
	"testing"

	"mybili/internal/db"
)

func TestAutoCreatesTables(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "new.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := Auto(conn); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"videos", "video_parts", "audio_parts", "danmaku",
		"favorite_lists", "favorite_list_videos",
		"subscriptions", "subscription_videos", "upper",
		"covers", "coverables", "settings", "tasks",
	} {
		if !conn.Migrator().HasTable(name) {
			t.Fatalf("new database is missing table %s", name)
		}
	}
	if !conn.Migrator().HasColumn("covers", "thumbnail_generated_at") {
		t.Fatal("covers is missing thumbnail_generated_at")
	}
	if conn.Migrator().HasTable("jobs") {
		t.Fatal("should not create Laravel's jobs table")
	}
}

func TestAutoSettlesAfterTinyintColumns(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "old.sqlite"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Exec(`CREATE TABLE videos (
		id integer primary key autoincrement not null,
		link varchar not null,
		title varchar not null,
		intro text not null,
		cover varchar not null,
		bvid varchar not null,
		pubtime datetime not null,
		attr integer not null default 0,
		invalid tinyint(1) not null default 0,
		frozen tinyint(1) not null default 0,
		cache_image varchar not null,
		page integer not null default 1,
		fav_time datetime,
		created_at datetime,
		updated_at datetime,
		video_downloaded_at datetime,
		video_downloaded_num integer not null default 0,
		duration integer not null default 0,
		upper_id integer,
		deleted_at datetime,
		type integer not null default 2,
		audio_downloaded_num integer not null default 0
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Exec(`INSERT INTO videos (link, title, intro, cover, bvid, pubtime, cache_image, invalid, frozen) VALUES ('l','t','i','c','bv','2025-01-01 00:00:00','',1,0)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := Auto(conn); err != nil {
		t.Fatal(err)
	}
	if err := conn.Exec(`CREATE INDEX videos_keep_me ON videos(bvid)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := Auto(conn); err != nil {
		t.Fatal(err)
	}
	var kept, invalid, rows int
	if err := conn.Raw(`SELECT count(*) FROM sqlite_master WHERE type='index' AND name='videos_keep_me'`).Scan(&kept).Error; err != nil {
		t.Fatal(err)
	}
	if kept != 1 {
		t.Fatal("second startup rebuilt videos")
	}
	if err := conn.Raw(`SELECT invalid FROM videos`).Scan(&invalid).Error; err != nil {
		t.Fatal(err)
	}
	if invalid != 1 {
		t.Fatalf("invalid changed to %d", invalid)
	}
	if err := conn.Raw(`SELECT count(*) FROM pragma_table_info('videos') WHERE name='invalid' AND type='tinyint(1)'`).Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatal("invalid is still tinyint(1)")
	}
}
