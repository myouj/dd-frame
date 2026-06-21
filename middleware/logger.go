package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	applog "github.com/example/dd-frame/pkg/log"
)

// Logger 请求日志中间件
//
// 自动记录请求方法、路径、状态码、延迟，
// 并携带 requestID 和 traceID（如已启用分布式追踪）。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)

		// 提取 traceID（如果 otelgin 已注入 span）
		traceID := ""
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}

		applog.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"request_id", c.GetString("request_id"),
			"trace_id", traceID,
		)
	}
}
