package main

import (
	"context"
	"time"

	"github.com/example/dd-frame/app"
	appcron "github.com/example/dd-frame/pkg/cron"
	applog "github.com/example/dd-frame/pkg/log"

	_ "github.com/example/dd-frame/docs" // swagger 生成文档
)

//	@title		dd-frame API
//	@version	1.0
//	@description	DDD 模块化单体项目 API 文档
//	@BasePath	/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description			Bearer JWT Token

func main() {
	// 1. 加载配置
	cfg, err := app.LoadConfig("config/config.yaml")
	if err != nil {
		panic("load config failed: " + err.Error())
	}

	// 2. 初始化日志
	app.InitLogger(&cfg.Log)
	defer applog.Sync()

	// 3. 初始化分布式追踪（返回 shutdown 函数）
	shutdownTracing := app.InitTracing(&cfg.Tracing)
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			applog.Error("tracing shutdown error", "err", err)
		}
	}()

	// 4. 初始化数据库（未配置时自动跳过）
	_, err = app.InitDatabase(&cfg.Database)
	if err != nil {
		panic("init database failed: " + err.Error())
	}

	// 5. 初始化 Redis（未配置时自动跳过）
	_, err = app.InitRedis(&cfg.Redis)
	if err != nil {
		panic("init redis failed: " + err.Error())
	}

	// 6. 装配模块
	router := app.Wire(cfg)

	// 7. 初始化定时任务
	scheduler := initCron(&cfg.Cron)
	if scheduler != nil {
		defer scheduler.Stop()
	}

	// 8. 启动服务器
	app.RunServer(cfg, router)
}

// initCron 初始化定时任务调度器
func initCron(cfg *app.CronConfig) *appcron.Scheduler {
	if !cfg.Enabled {
		return nil
	}

	s := appcron.NewScheduler(appcron.Config{
		Enabled:   cfg.Enabled,
		Locker:    cfg.Locker,
		KeyPrefix: cfg.KeyPrefix,
	}, app.GlobalRedis)

	// ---------- 注册定时任务 ----------
	// 示例：每 5 分钟执行一次清理任务
	if err := s.AddJob("example:cleanup", "0 */5 * * * *", 2*time.Minute, func(ctx context.Context) {
		applog.Info("example cleanup job running")
	}); err != nil {
		applog.Error("cron: failed to register job", "job", "example:cleanup", "err", err)
	}

	// 新增任务在此追加，例如：
	// if err := s.AddJob("report:daily", "0 0 2 * * *", 10*time.Minute, dailyReport); err != nil {
	// 	applog.Error("cron: failed to register job", "job", "report:daily", "err", err)
	// }

	s.Start()
	return s
}
