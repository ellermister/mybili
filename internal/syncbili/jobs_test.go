package syncbili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"gorm.io/gorm"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/jobs"
	"mybili/internal/model"
	"mybili/internal/schedule"
	"mybili/internal/taskrun"
)

func TestFavTickSkipsWhenSwitchOffOrOffSlot(t *testing.T) {
	conn := openTestDB(t)
	runner := favRunner(conn)
	if err := runner.Tick(context.Background(), shanghai(2026, 10, 30, 10, 0)); err != nil {
		t.Fatal(err)
	}
	enableFavSync(t, conn)
	if err := runner.Tick(context.Background(), shanghai(2026, 10, 30, 10, 3)); err != nil {
		t.Fatal(err)
	}
	if n := countJobs(t, conn); n != 0 {
		t.Fatalf("tasks=%d", n)
	}
}

func TestFavTickEnqueuesListAndFirstPage(t *testing.T) {
	conn := openTestDB(t)
	enableFavSync(t, conn)
	insertVideoRow(t, conn, 1)
	now := shanghai(2026, 10, 30, 10, 0)
	runner := favRunner(conn)
	if err := runner.Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if err := runner.Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	var rows []model.Job
	if err := conn.Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].UniqueKey != "sync:fav-list" || rows[0].Queue != taskrun.QueueBiliFast || rows[0].Type != TypeFavList || rows[0].Status != jobs.StatusPending {
		t.Fatalf("list job = %+v", rows[0])
	}
	if rows[1].UniqueKey != "sync:fav-videos:1" || rows[1].Payload != `{"page":1}` {
		t.Fatalf("page job = %+v", rows[1])
	}
}

func TestFavTickAtFourUsesFullSync(t *testing.T) {
	conn := openTestDB(t)
	enableFavSync(t, conn)
	insertVideoRow(t, conn, 1)
	now := shanghai(2026, 10, 30, 4, 0)
	if err := favRunner(conn).Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := conn.Model(&model.Job{}).Order("unique_key").Pluck("unique_key", &keys).Error; err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != "sync:fav-list" || keys[1] != "sync:fav-videos:0" {
		t.Fatalf("keys = %v", keys)
	}
}

func TestFavTickEmptyLibraryUsesFullSync(t *testing.T) {
	conn := openTestDB(t)
	enableFavSync(t, conn)
	now := shanghai(2026, 10, 30, 11, 20)
	if err := favRunner(conn).Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	var job model.Job
	if err := conn.Where("unique_key = ?", "sync:fav-videos:0").First(&job).Error; err != nil {
		t.Fatal(err)
	}
	if job.Payload != `{"page":0}` {
		t.Fatalf("payload = %s", job.Payload)
	}
}

func TestFavTickReopensFinishedTask(t *testing.T) {
	conn := openTestDB(t)
	enableFavSync(t, conn)
	insertVideoRow(t, conn, 1)
	earlier := dbtime.DateTime{Time: shanghai(2026, 10, 30, 12, 0)}
	done := model.Job{
		Queue: taskrun.QueueBiliFast, Type: TypeFavList, UniqueKey: "sync:fav-list",
		Status: jobs.StatusDone, Payload: "{}", MaxRetries: 3, UpdatedAt: &earlier,
	}
	if err := conn.Create(&done).Error; err != nil {
		t.Fatal(err)
	}
	now := shanghai(2026, 10, 30, 12, 10)
	if err := favRunner(conn).Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	var job model.Job
	if err := conn.Where("unique_key = ?", "sync:fav-list").First(&job).Error; err != nil {
		t.Fatal(err)
	}
	if job.ID != done.ID || job.Status != jobs.StatusPending {
		t.Fatalf("job = %+v", job)
	}
}

func TestFavTickKeepsFinishedInsideSlot(t *testing.T) {
	conn := openTestDB(t)
	enableFavSync(t, conn)
	insertVideoRow(t, conn, 1)
	now := shanghai(2026, 10, 30, 12, 10)
	stamp := dbtime.DateTime{Time: now}
	done := model.Job{
		Queue: taskrun.QueueBiliFast, Type: TypeFavList, UniqueKey: "sync:fav-list",
		Status: jobs.StatusDone, Payload: "{}", MaxRetries: 3, UpdatedAt: &stamp,
	}
	if err := conn.Create(&done).Error; err != nil {
		t.Fatal(err)
	}
	if err := favRunner(conn).Tick(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	var job model.Job
	if err := conn.Where("unique_key = ?", "sync:fav-list").First(&job).Error; err != nil {
		t.Fatal(err)
	}
	if job.Status != jobs.StatusDone {
		t.Fatalf("job = %+v", job)
	}
}

func TestFavHandlerRunsList(t *testing.T) {
	conn := openTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"list":[{"id":11,"title":"新标题","cover":"http://img.example/a.jpg","ctime":1499900654,"mtime":1523352284,"media_count":8}],"has_more":false}}`))
	}))
	defer server.Close()
	job := insertPending(t, conn, TypeFavList, "sync:fav-list", "{}")
	handlers := taskHandlers(nil, func() (Sync, error) {
		return Sync{
			DB:           conn,
			API:          &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()},
			Cookies:      []bilibili.Cookie{{Name: "DedeUserID", Value: "7"}},
			PageInterval: 0,
			Now:          func() time.Time { return time.Now() },
			CoverPush:    &memCoverPush{},
		}, nil
	})
	err := handlers[TypeFavList](context.Background(), taskrun.Message{ID: job.ID, Type: TypeFavList, Payload: job.Payload})
	if err != nil {
		t.Fatal(err)
	}
	var folder model.FavoriteList
	if err := conn.First(&folder, 11).Error; err != nil {
		t.Fatal(err)
	}
	if folder.Title != "新标题" {
		t.Fatalf("title = %s", folder.Title)
	}
	saved := reloadJob(t, conn, job.ID)
	if saved.Status != jobs.StatusDone {
		t.Fatalf("status = %s", saved.Status)
	}
}

func TestFavHandlerRunsVideosPage(t *testing.T) {
	conn := openTestDB(t)
	folder := model.FavoriteList{
		ID: 7, Title: "夹", Cover: "c", MediaCount: 1, CacheImage: "",
		Ctime: stamp(1), Mtime: stamp(2),
	}
	if err := conn.Create(&folder).Error; err != nil {
		t.Fatal(err)
	}
	var page string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page = r.URL.Query().Get("pn")
		_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[{"id":21,"type":2,"title":"一条","cover":"http://img.example/v.jpg","intro":"i","page":1,"duration":3,"upper":{"mid":9},"attr":0,"bvid":"BV21","pubtime":1499900654,"fav_time":1743737512}],"has_more":false}}`))
	}))
	defer server.Close()
	job := insertPending(t, conn, TypeFavVideos, "sync:fav-videos:1", `{"page":1}`)
	handlers := taskHandlers(nil, func() (Sync, error) {
		return Sync{
			DB:           conn,
			API:          &bilibili.Client{BaseURL: server.URL, HTTP: server.Client()},
			Cookies:      []bilibili.Cookie{{Name: "DedeUserID", Value: "7"}},
			PageSize:     40,
			PageInterval: 0,
			Now:          func() time.Time { return time.Now() },
			CoverPush:    &memCoverPush{},
		}, nil
	})
	err := handlers[TypeFavVideos](context.Background(), taskrun.Message{ID: job.ID, Type: TypeFavVideos, Payload: job.Payload})
	if err != nil {
		t.Fatal(err)
	}
	if page != "1" {
		t.Fatalf("page = %s", page)
	}
	var video model.Video
	if err := conn.First(&video, 21).Error; err != nil {
		t.Fatal(err)
	}
	if video.Title != "一条" {
		t.Fatalf("title = %s", video.Title)
	}
}

func TestFavHandlerSuspendMarksFailed(t *testing.T) {
	conn := openTestDB(t)
	until := time.Now().Add(time.Hour).Unix()
	text := fmtInt(until)
	row := model.Setting{Name: "bilibili_suspend_until", Value: &text}
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	job := insertPending(t, conn, TypeFavList, "sync:fav-list", "{}")
	handlers := taskHandlers(nil, func() (Sync, error) {
		return Sync{DB: conn, Now: func() time.Time { return time.Now() }}, nil
	})
	err := handlers[TypeFavList](context.Background(), taskrun.Message{ID: job.ID, Type: TypeFavList, Payload: "{}"})
	if err != nil {
		t.Fatal(err)
	}
	saved := reloadJob(t, conn, job.ID)
	if saved.Status != jobs.StatusFailed {
		t.Fatalf("status = %s", saved.Status)
	}
}

func favRunner(conn *gorm.DB) schedule.Runner {
	return schedule.Runner{Specs: FavSpecs(conn), Store: jobs.Store{DB: conn}}
}

func enableFavSync(t *testing.T, conn *gorm.DB) {
	t.Helper()
	text := `"on"`
	row := model.Setting{Name: settingFavoriteSync, Value: &text}
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
}

func insertVideoRow(t *testing.T, conn *gorm.DB, id int64) {
	t.Helper()
	row := model.Video{
		ID: id, Link: "l", Title: "t", Intro: "i", Cover: "c", Bvid: "BV",
		Pubtime: dbtime.DateTime{Time: shanghai(2026, 1, 1, 0, 0)}, CacheImage: "", Page: 1, Type: 2,
	}
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
}

func insertPending(t *testing.T, conn *gorm.DB, taskType, key, payload string) model.Job {
	t.Helper()
	row := model.Job{
		Queue: taskrun.QueueBiliFast, Type: taskType, UniqueKey: key,
		Status: jobs.StatusPending, Payload: payload, MaxRetries: 3,
	}
	if err := conn.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	return row
}

func reloadJob(t *testing.T, conn *gorm.DB, id int64) model.Job {
	t.Helper()
	var job model.Job
	if err := conn.First(&job, id).Error; err != nil {
		t.Fatal(err)
	}
	return job
}

func countJobs(t *testing.T, conn *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := conn.Model(&model.Job{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func shanghai(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, dbtime.Location())
}

func fmtInt(n int64) string {
	return strconv.FormatInt(n, 10)
}
