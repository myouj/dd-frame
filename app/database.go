package app

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	applog "github.com/example/dd-frame/pkg/log"
)

// GlobalDB 全局数据库实例
var GlobalDB *gorm.DB

// InitDatabase 初始化 GORM 数据库连接
//
// 如果 Host 或 DBName 为空，跳过初始化（未配置数据库时不报错）。
func InitDatabase(cfg *DatabaseConfig) (*gorm.DB, error) {
	if cfg.Host == "" || cfg.DBName == "" {
		applog.Info("database skipped: no host or dbname configured")
		return nil, nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var gormLogLevel logger.LogLevel
	if GlobalConfig.Server.Mode == "debug" {
		gormLogLevel = logger.Info
	} else {
		gormLogLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB failed: %w", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)

	GlobalDB = db
	applog.Info("database connected", "driver", cfg.Driver, "host", cfg.Host, "dbname", cfg.DBName)
	return db, nil
}
