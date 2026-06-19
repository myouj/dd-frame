package app

import (
	"github.com/example/dd-frame/pkg/log"
)

// InitLogger 初始化结构化日志
func InitLogger(cfg *LogConfig) {
	log.Init(cfg.Level, cfg.Format)
	log.Info("logger initialized", "level", cfg.Level, "format", cfg.Format)
}
