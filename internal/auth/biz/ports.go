package biz

import "context"

// CacheInvalidator 权限缓存失效端口
type CacheInvalidator interface {
	// InvalidateUserPermissions 清除指定用户的权限缓存
	InvalidateUserPermissions(ctx context.Context, userID int64) error
}

// AuditLogger 审计日志端口
type AuditLogger interface {
	LogLogin(ctx context.Context, userID int64, username string, success bool)
	LogPermissionChange(ctx context.Context, operatorID int64, action string, detail string)
}

// StubCacheInvalidator 无缓存时的空实现
type StubCacheInvalidator struct{}

func (s *StubCacheInvalidator) InvalidateUserPermissions(_ context.Context, _ int64) error {
	return nil
}

// StubAuditLogger 无审计时的空实现
type StubAuditLogger struct{}

func (s *StubAuditLogger) LogLogin(_ context.Context, _ int64, _ string, _ bool)              {}
func (s *StubAuditLogger) LogPermissionChange(_ context.Context, _ int64, _ string, _ string) {}
