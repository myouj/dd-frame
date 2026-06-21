package app

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	applog "github.com/example/dd-frame/pkg/log"
)

// RegisterMetricsRoute 注册 Prometheus /metrics 端点
//
// 仅在配置启用时注册。使用 promhttp 原生 handler。
func RegisterMetricsRoute(r *gin.Engine, cfg *MetricsConfig) {
	if !cfg.Enabled {
		applog.Info("metrics disabled")
		return
	}

	path := cfg.Path
	if path == "" {
		path = "/metrics"
	}

	r.GET(path, gin.WrapH(promhttp.Handler()))
	applog.Info("metrics endpoint registered", "path", path)
}
