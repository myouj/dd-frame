package model

import "context"

// OrderCache 订单缓存接口
type OrderCache interface {
	SetOrderLock(ctx context.Context, orderNo string, ttl int) (bool, error)
	ReleaseOrderLock(ctx context.Context, orderNo string) error
	GetOrderCache(ctx context.Context, orderNo string) (string, error)
	SetOrderCache(ctx context.Context, orderNo string, data string, ttl int) error
}
