package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/metrics"
)

// Metrics Prometheus HTTP 指标中间件。
//
// 自动采集每个请求的计数、延迟和并发数。
func Metrics() gin.HandlerFunc {
	return metrics.HTTPMiddleware()
}
