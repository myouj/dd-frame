package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/example/dd-frame/pkg/auth"
	applog "github.com/example/dd-frame/pkg/log"
)

// Logger 请求日志中间件
//
// 自动记录请求方法、路径、状态码、延迟，
// 并携带 requestID、traceID、userID、clientIP。
//
// 在 c.Next() 前将上下文 logger 注入 gin.Context，
// 业务代码可通过 log.FromContext(c.Request.Context()) 获取带字段的 logger。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 提取 traceID（如果 otelgin 已注入 span）
		traceID := ""
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}

		requestID := c.GetString("request_id")
		clientIP := c.ClientIP()

		// 注入 context logger（业务代码可用 log.FromContext 获取）
		fields := []interface{}{
			"request_id", requestID,
			"trace_id", traceID,
			"client_ip", clientIP,
			"method", c.Request.Method,
			"path", path,
		}
		ctx := applog.WithLogger(c.Request.Context(), fields...)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// 提取 userID（如已认证）
		userID := ""
		if user, ok := auth.CurrentUser(c); ok {
			userID = user.Username
		}

		latency := time.Since(start)
		applog.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"request_id", requestID,
			"trace_id", traceID,
			"client_ip", clientIP,
			"user_id", userID,
		)
	}
}
