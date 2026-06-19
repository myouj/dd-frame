package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/log"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		rid, _ := c.Get("request_id")

		log.Info("request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
			"request_id", rid,
		)
	}
}
