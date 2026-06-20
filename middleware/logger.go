package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	applog "github.com/example/dd-frame/pkg/log"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		applog.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"request_id", c.GetString("request_id"),
		)
	}
}
