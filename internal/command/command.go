package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	"mybili/internal/config"
	"mybili/internal/db"
	"mybili/internal/dbtime"
	"mybili/internal/download"
	"mybili/internal/feeder"
	"mybili/internal/httpapi"
	"mybili/internal/jobs"
	"mybili/internal/migrate"
	"mybili/internal/model"
	"mybili/internal/notify"
	"mybili/internal/readable"
	"mybili/internal/schedule"
	"mybili/internal/syncbili"
	"mybili/internal/taskrun"
	"mybili/internal/usage"
)

func Run(ctx context.Context, cfg config.Config, args []string) error {
	cmd := &cli.Command{
		Name:  "mybili",
		Usage: "bilibili 收藏夹工具",
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "启动网站和接口，并派发定时任务",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "no-schedule", Usage: "不派发定时任务"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return serve(ctx, cfg, cmd.Bool("no-schedule"))
				},
			},
			{
				Name:  "worker",
				Usage: "执行收藏夹同步任务",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return work(ctx, cfg)
				},
			},
			{
				Name:  "migrate",
				Usage: "按模型对齐数据库（与设置页按钮相同）",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					conn, err := db.Open(cfg.DBPath, cfg.SQLLogLevel)
					if err != nil {
						return err
					}
					if err := migrate.Auto(conn); err != nil {
						return err
					}
					fmt.Println("数据库已对齐")
					return nil
				},
			},
			{
				Name:  "sync-fav",
				Usage: "同步收藏夹列表（不拉视频，封面有变化时写入任务）",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					syncer, err := openSyncer(cfg, true)
					if err != nil {
						return err
					}
					count, err := syncer.Folders(ctx)
					if count > 0 {
						fmt.Printf("收藏夹列表已更新，共 %d 条\n", count)
					}
					printCoverJobs(syncer)
					return err
				},
			},
			{
				Name:  "sync-fav-videos",
				Usage: "同步收藏夹里的视频（不拉分P，封面有变化时写入任务）",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "fav", Usage: "只同步这个收藏夹，不传则全部"},
					&cli.IntFlag{Name: "page", Usage: "只拉这一页，不传则全部页"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("收藏夹 ID 请用 --fav")
					}
					favID := cmd.Int64("fav")
					page := cmd.Int("page")
					if cmd.IsSet("fav") && favID <= 0 {
						return fmt.Errorf("--fav 必须大于 0")
					}
					if cmd.IsSet("page") && page <= 0 {
						return fmt.Errorf("--page 必须大于 0")
					}
					syncer, err := openSyncer(cfg, true)
					if err != nil {
						return err
					}
					result, err := syncer.FavoriteVideos(ctx, favID, page)
					if result.Folders > 0 {
						fmt.Printf("收藏夹视频已更新，%d 个收藏夹，%d 条视频\n", result.Folders, result.Videos)
					}
					printCoverJobs(syncer)
					return err
				},
			},
			{
				Name:  "sync-video",
				Usage: "同步视频详情和分P（封面或头像有变化时写入任务）",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "id", Usage: "视频 ID，也是稿件 aid"},
					&cli.StringFlag{Name: "bvid", Usage: "BV 号"},
					&cli.BoolFlag{Name: "missing-parts", Usage: "只处理还没有分P的视频"},
					&cli.IntFlag{Name: "limit", Usage: "配合 --missing-parts，最多处理多少个"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("请用 --id、--bvid 或 --missing-parts")
					}
					videoID := cmd.Int64("id")
					bvid := strings.TrimSpace(cmd.String("bvid"))
					missing := cmd.Bool("missing-parts")
					limit := cmd.Int("limit")
					if cmd.IsSet("id") && videoID <= 0 {
						return fmt.Errorf("--id 必须大于 0")
					}
					if missing && (videoID > 0 || bvid != "") {
						return fmt.Errorf("--missing-parts 不能和 --id、--bvid 一起用")
					}
					if missing && limit <= 0 {
						return fmt.Errorf("--missing-parts 需要 --limit，且必须大于 0")
					}
					if !missing && videoID <= 0 && bvid == "" {
						return fmt.Errorf("请指定 --id 或 --bvid")
					}
					if videoID > 0 && bvid != "" {
						return fmt.Errorf("--id 和 --bvid 只填一个")
					}
					syncer, err := openSyncer(cfg, true)
					if err != nil {
						return err
					}
					if missing {
						count, err := syncer.MissingParts(ctx, limit)
						if count > 0 {
							fmt.Printf("视频详情和分P已更新，共 %d 个\n", count)
						}
						printCoverJobs(syncer)
						return err
					}
					if err := syncer.Video(ctx, videoID, bvid); err != nil {
						printCoverJobs(syncer)
						return err
					}
					fmt.Println("视频详情和分P已更新")
					printCoverJobs(syncer)
					return nil
				},
			},
			{
				Name:  "fix-invalid-fav",
				Usage: "用收藏夹接口修正失效稿件的标题和封面",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "fav", Usage: "只处理这个收藏夹，不传则全部"},
					&cli.IntFlag{Name: "page", Usage: "只处理这一页，不传则按收藏数量翻页"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("请用 --fav 或 --page")
					}
					favID := cmd.Int64("fav")
					page := cmd.Int("page")
					if cmd.IsSet("fav") && favID <= 0 {
						return fmt.Errorf("--fav 必须大于 0")
					}
					if cmd.IsSet("page") && page <= 0 {
						return fmt.Errorf("--page 必须大于 0")
					}
					syncer, err := openSyncer(cfg, true)
					if err != nil {
						return err
					}
					fixed, err := fixInvalidFavorites(ctx, syncer, favID, page)
					if fixed > 0 {
						fmt.Printf("失效稿件已修正 %d 条\n", fixed)
					}
					printCoverJobs(syncer)
					return err
				},
			},
			{
				Name:  "fix-freeze",
				Usage: "按已下载的分P修正冻结且无效的视频",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("这个命令没有参数")
					}
					syncer, err := openSyncer(cfg, false)
					if err != nil {
						return err
					}
					fixed, err := syncer.FixFreeze(ctx)
					if fixed > 0 {
						fmt.Printf("冻结状态已修正 %d 条\n", fixed)
					}
					return err
				},
			},
			{
				Name:  "readable-names",
				Usage: "清空可读目录，并按收藏夹和订阅重建硬链接",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("这个命令没有参数")
					}
					conn, err := db.Open(cfg.DBPath, cfg.SQLLogLevel)
					if err != nil {
						return err
					}
					n, err := readable.Generate(ctx, conn, cfg.DataDir, cfg.HumanReadableDir)
					if err != nil {
						return err
					}
					fmt.Printf("可读文件名已生成 %d 个链接\n", n)
					return nil
				},
			},
			{
				Name:  "sync-subscriptions",
				Usage: "更新订阅里的稿件（默认只拉第一页）",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "id", Usage: "只更新这个订阅，不传则全部启用中的订阅"},
					&cli.BoolFlag{Name: "pull-all", Usage: "拉完全部页"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("请用 --id 或 --pull-all")
					}
					subID := cmd.Int64("id")
					if cmd.IsSet("id") && subID <= 0 {
						return fmt.Errorf("--id 必须大于 0")
					}
					syncer, err := openSyncer(cfg, true)
					if err != nil {
						return err
					}
					syncer.QueueVideo = func(ctx context.Context, bvid string) error {
						return syncbili.EnqueueVideoDetail(ctx, syncer.DB, bvid, dbtime.Now(), taskrun.New(cfg).Wake)
					}
					pullAll := cmd.Bool("pull-all")
					if subID > 0 {
						if err := syncer.UpdateSubscription(ctx, subID, pullAll); err != nil {
							return err
						}
						fmt.Println("订阅已更新")
						printCoverJobs(syncer)
						return nil
					}
					var rows []model.Subscription
					if err := syncer.DB.WithContext(ctx).Where("status = ?", 1).Order("id").Find(&rows).Error; err != nil {
						return err
					}
					for _, row := range rows {
						if err := syncer.UpdateSubscription(ctx, row.ID, pullAll); err != nil {
							return err
						}
						fmt.Printf("订阅 %d 已更新\n", row.ID)
					}
					printCoverJobs(syncer)
					return nil
				},
			},
			{
				Name:  "scan-downloads",
				Usage: "检查已经记过下载的视频和音频，文件不在时写入下载任务",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "video", Usage: "只检查这个视频"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("视频 ID 请用 --video")
					}
					videoID := cmd.Int64("video")
					if cmd.IsSet("video") && videoID <= 0 {
						return fmt.Errorf("--video 必须大于 0")
					}
					conn, err := db.Open(cfg.DBPath, cfg.SQLLogLevel)
					if err != nil {
						return err
					}
					tasks := taskrun.New(cfg)
					count, err := download.EnqueueLost(ctx, conn, cfg.DataDir, videoID, dbtime.Now(), tasks.Wake)
					if err != nil {
						return err
					}
					fmt.Printf("丢失的文件已入队 %d 条\n", count)
					return nil
				},
			},
			{
				Name:  "reconcile-downloads",
				Usage: "按磁盘上已有的视频和音频补下载路径，并重算已下载数量",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "video", Usage: "只检查这个视频"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("视频 ID 请用 --video")
					}
					videoID := cmd.Int64("video")
					if cmd.IsSet("video") && videoID <= 0 {
						return fmt.Errorf("--video 必须大于 0")
					}
					conn, err := db.Open(cfg.DBPath, cfg.SQLLogLevel)
					if err != nil {
						return err
					}
					result, err := download.ReconcileDownloads(ctx, conn, cfg.DataDir, videoID, dbtime.Now())
					if err != nil {
						return err
					}
					fmt.Printf("已有文件已记下 %d 条\n", result.Noted)
					fmt.Printf("下载数量已修正 %d 条\n", result.Recounted)
					return nil
				},
			},
			{
				Name:  "sync-covers",
				Usage: "检查还没有本地封面的记录并写入任务",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "type", Usage: "只检查 video、favorite、avatar、subscription 中的一种"},
					&cli.IntFlag{Name: "limit", Usage: "最多入队多少条，不传则不限制"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						return fmt.Errorf("请用 --type 或 --limit")
					}
					kind := strings.TrimSpace(cmd.String("type"))
					switch kind {
					case "", syncbili.CoverKindVideo, syncbili.CoverKindFavorite, syncbili.CoverKindAvatar, syncbili.CoverKindSubscription:
					default:
						return fmt.Errorf("--type 只能是 video、favorite、avatar、subscription")
					}
					limit := 0
					if cmd.IsSet("limit") {
						limit = cmd.Int("limit")
						if limit <= 0 {
							return fmt.Errorf("--limit 必须大于 0")
						}
					}
					syncer, err := openSyncer(cfg, false)
					if err != nil {
						return err
					}
					count, err := syncer.Covers(ctx, kind, limit)
					if err != nil {
						return err
					}
					fmt.Printf("封面任务已写入 %d 条\n", count)
					return nil
				},
			},
		},
	}
	return cmd.Run(ctx, args)
}

func openSyncer(cfg config.Config, requireCookie bool) (syncbili.Sync, error) {
	conn, err := db.Open(cfg.DBPath, cfg.SQLLogLevel)
	if err != nil {
		return syncbili.Sync{}, err
	}
	syncer, err := syncbili.New(conn, cfg.FavPageSize)
	if err != nil {
		if requireCookie {
			return syncbili.Sync{}, err
		}
		syncer = syncbili.WithoutClient(conn)
	}
	syncer.DataDir = cfg.DataDir
	tasks := taskrun.New(cfg)
	syncer.CoverPush = syncbili.CoverRedis(tasks)
	syncer.Wake = tasks.Wake
	syncer.UseNotify(tasks)
	return syncer, nil
}

func fixInvalidFavorites(ctx context.Context, syncer syncbili.Sync, favID int64, page int) (int, error) {
	var folders []model.FavoriteList
	query := syncer.DB.WithContext(ctx).Order("id")
	if favID > 0 {
		query = query.Where("id = ?", favID)
	}
	if err := query.Find(&folders).Error; err != nil {
		return 0, err
	}
	if favID > 0 && len(folders) == 0 {
		return 0, fmt.Errorf("favorite %d is not in the local list, run sync-fav first", favID)
	}
	total := 0
	for _, folder := range folders {
		pages := []int{page}
		if page <= 0 {
			pages = folderPages(folder.MediaCount, syncer.PageSize)
		}
		for _, current := range pages {
			fixed, err := syncer.FixInvalidFavorite(ctx, folder.ID, current)
			total += fixed
			if err != nil {
				return total, err
			}
		}
	}
	return total, nil
}

func folderPages(count, pageSize int) []int {
	if pageSize < 1 {
		pageSize = 40
	}
	if count <= 0 {
		return nil
	}
	n := (count + pageSize - 1) / pageSize
	pages := make([]int, 0, n)
	for page := 1; page <= n; page++ {
		pages = append(pages, page)
	}
	return pages
}

func printCoverJobs(syncer syncbili.Sync) {
	if n := syncer.CoverJobs(); n > 0 {
		fmt.Printf("封面任务已写入 %d 条\n", n)
	}
}

func serve(ctx context.Context, cfg config.Config, noSchedule bool) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	conn, err := db.Open(cfg.DBPath, cfg.SQLLogLevel)
	if err != nil {
		return err
	}
	if cfg.AutoMigrate {
		if err := migrate.Auto(conn); err != nil {
			return err
		}
	}
	reporter := usage.Start(conn)
	go reporter.Loop(ctx)
	tasks := taskrun.NewBlocking(cfg)
	if noSchedule {
		log.Printf("定时任务已关闭")
	} else {
		specs := syncbili.FavSpecs(conn)
		specs = append(specs, syncbili.SubSpecs(conn)...)
		specs = append(specs, syncbili.FixInvalidSpecs(conn, cfg.FavPageSize)...)
		specs = append(specs, syncbili.MissingPartSpecs(conn)...)
		specs = append(specs, syncbili.FixFreezeSpecs()...)
		go schedule.Runner{
			Specs: specs,
			Store: jobs.Store{DB: conn},
			Wake:  tasks.Wake,
		}.Loop(ctx)
		log.Printf("定时任务已开启")
	}
	go feeder.Feeder{
		DB:     conn,
		Tasks:  tasks,
		Queues: taskrun.FeedQueues(),
	}.Loop(ctx)
	log.Printf("任务投喂已开启")
	app := &httpapi.App{
		DB: conn, Config: cfg, Tasks: tasks,
		Output: func(ctx context.Context, id int64) ([]string, int, int, string, error) {
			lines, progress, err := tasks.TaskOutput(ctx, id)
			if lines == nil {
				lines = []string{}
			}
			return lines, progress.Done, progress.Total, progress.Text, err
		},
	}
	dist := filepath.Join("web", "dist")
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: app.Handler(dist),
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("mybili 监听 %s", cfg.HTTPAddr)
	err = server.ListenAndServe()
	reporter.End()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func work(ctx context.Context, cfg config.Config) error {
	conn, err := db.Open(cfg.DBPath, cfg.SQLLogLevel)
	if err != nil {
		return err
	}
	if cfg.AutoMigrate {
		if err := migrate.Auto(conn); err != nil {
			return err
		}
	}
	log.Printf("worker 已启动")
	tasks := taskrun.NewBlocking(cfg)
	syncbili.Register(tasks, conn, cfg.FavPageSize, cfg.DataDir, syncbili.Hooks{
		AfterLinked: func(ctx context.Context, ids []int64) {
			_ = readable.EnqueueVideos(ctx, conn, tasks.Wake, ids)
		},
		AfterUnlinked: func(ctx context.Context, folder string, ids []int64) {
			_ = readable.UnlinkFolder(ctx, conn, cfg.DataDir, cfg.HumanReadableDir, folder, ids)
		},
	})
	download.Register(tasks, conn, cfg.DataDir, download.Hooks{
		PartReady: func(ctx context.Context, ready download.PartReady) {
			if ready.Fresh {
				_ = notify.EnqueueDownloaded(ctx, tasks, ready.PartID)
			}
			_ = readable.EnqueuePart(ctx, conn, tasks.Wake, ready.PartID)
		},
	})
	notify.Register(tasks, conn, cfg.DataDir)
	readable.Register(tasks, conn, cfg.DataDir, cfg.HumanReadableDir)
	return tasks.Loop(ctx, taskrun.AllQueues())
}
