package log

import (
	"context"

	"go.uber.org/zap"
)

type ctxLoggerKey struct{}

// WithLogger 将带字段的 SugaredLogger 注入 context
//
// 后续通过 FromContext 获取时自动携带这些字段。
func WithLogger(ctx context.Context, fields ...interface{}) context.Context {
	l := Logger.With(fields...)
	return context.WithValue(ctx, ctxLoggerKey{}, l)
}

// FromContext 从 context 获取请求级 logger
//
// 如果 context 中没有注入 logger，返回全局 Logger。
// 业务代码推荐使用此方法获取 logger，自动携带请求上下文。
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if l, ok := ctx.Value(ctxLoggerKey{}).(*zap.SugaredLogger); ok {
		return l
	}
	return Logger
}
