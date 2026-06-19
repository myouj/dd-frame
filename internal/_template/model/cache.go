package model

import "context"

// EntityCache 缓存接口
type EntityCache interface {
	SetLock(ctx context.Context, key string, ttl int) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, data string, ttl int) error
}
