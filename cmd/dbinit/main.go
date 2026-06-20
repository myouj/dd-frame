package main

import (
	"flag"
	"fmt"
	"os"

	"gorm.io/gorm"

	"github.com/example/dd-frame/app"
	authmodule "github.com/example/dd-frame/internal/auth"
	authmodel "github.com/example/dd-frame/internal/auth/model"
	"github.com/example/dd-frame/pkg/auth"
	applog "github.com/example/dd-frame/pkg/log"
)

func main() {
	seed := flag.Bool("seed", false, "初始化种子数据（admin 角色/用户 + 基础权限）")
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %s\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	app.InitLogger(&cfg.Log)
	defer applog.Sync()

	// 3. 连接数据库
	db, err := app.InitDatabase(&cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect database failed: %s\n", err)
		os.Exit(1)
	}
	if db == nil {
		fmt.Fprintln(os.Stderr, "database not configured, nothing to do")
		os.Exit(1)
	}

	// 4. 迁移表结构
	fmt.Println("==> migrating tables...")
	if err := migrate(db); err != nil {
		fmt.Fprintf(os.Stderr, "migrate failed: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("==> tables migrated successfully")

	// 5. 种子数据（可选）
	if *seed {
		fmt.Println("==> seeding data...")
		// 通过 auth 模块 Wire 触发种子数据（seedEnabled=true）
		// Wire 内部会调用 seedData，已存在的数据会跳过
		authmodule_seed(db, cfg)
		fmt.Println("==> seed data initialized")
	}

	fmt.Println("done.")
}

// migrate 执行所有模块的表结构迁移
func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// auth 模块
		&authmodel.UserModel{},
		&authmodel.RoleModel{},
		&authmodel.PermissionModel{},
		&authmodel.UserRoleModel{},
		&authmodel.RolePermissionModel{},
		// 新增模块的 model 在此追加
	)
}

// authmodule_seed 调用 auth 模块 Wire 触发种子数据
func authmodule_seed(db *gorm.DB, cfg *app.Config) {
	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresIn)
	// Wire 的 seedEnabled=true 会触发 seedData
	authmodule.Wire(db, jwtMgr, true)
}
