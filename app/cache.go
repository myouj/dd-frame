package app

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	applog "github.com/example/dd-frame/pkg/log"
)

// GlobalRedis 全局 Redis 客户端实例
var GlobalRedis *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis failed: %w", err)
	}

	GlobalRedis = rdb
	applog.Info("redis connected", "addr", cfg.Addr, "db", cfg.DB)
	return rdb, nil
}
