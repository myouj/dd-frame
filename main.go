package main

import (
	"github.com/example/dd-frame/app"
	applog "github.com/example/dd-frame/pkg/log"
)

func main() {
	// 1. 加载配置
	cfg, err := app.LoadConfig("config/config.yaml")
	if err != nil {
		panic("load config failed: " + err.Error())
	}

	// 2. 初始化日志
	app.InitLogger(&cfg.Log)
	defer applog.Sync()

	// 3. 初始化数据库（未配置时自动跳过）
	_, err = app.InitDatabase(&cfg.Database)
	if err != nil {
		panic("init database failed: " + err.Error())
	}

	// 4. 初始化 Redis（未配置时自动跳过）
	_, err = app.InitRedis(&cfg.Redis)
	if err != nil {
		panic("init redis failed: " + err.Error())
	}

	// 5. 装配模块
	router := app.Wire(cfg)

	// 6. 启动服务器
	app.RunServer(cfg, router)
}
