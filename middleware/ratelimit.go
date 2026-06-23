package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"

	applog "github.com/example/dd-frame/pkg/log"
)

// skipRateLimitPaths 不参与限流的路径
var skipRateLimitPaths = map[string]bool{
	"/health":  true,
	"/ready":   true,
	"/metrics": true,
}

// --- 内存后端 ---

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type memoryBackend struct {
	mu       sync.RWMutex
	limiters map[string]*ipLimiter
	rate     rate.Limit
	burst    int
}

func newMemoryBackend(r float64, burst int) *memoryBackend {
	mb := &memoryBackend{
		limiters: make(map[string]*ipLimiter),
		rate:     rate.Limit(r),
		burst:    burst,
	}
	// 定期清理过期 limiter（5 分钟无活动）
	go mb.cleanup()
	return mb
}

func (m *memoryBackend) getLimiter(ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.limiters[ip]; ok {
		v.lastSeen = time.Now()
		return v.limiter
	}
	l := rate.NewLimiter(m.rate, m.burst)
	m.limiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

func (m *memoryBackend) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		for ip, v := range m.limiters {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(m.limiters, ip)
			}
		}
		m.mu.Unlock()
	}
}

func (m *memoryBackend) allow(ip string) bool {
	return m.getLimiter(ip).Allow()
}

// --- Redis 后端 ---

type redisBackend struct {
	client    *redis.Client
	keyPrefix string
	rate      int // 每秒请求数上限
}

func newRedisBackend(client *redis.Client, keyPrefix string, r float64) *redisBackend {
	if keyPrefix == "" {
		keyPrefix = "rl:"
	}
	return &redisBackend{
		client:    client,
		keyPrefix: keyPrefix,
		rate:      int(r),
	}
}

func (r *redisBackend) allow(ip string) bool {
	ctx := context.Background()
	key := r.keyPrefix + ip

	// 滑动窗口：INCR + EXPIRE 1s
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		applog.Warn("ratelimit redis error, falling back to allow", "err", err)
		return true
	}
	if count == 1 {
		r.client.Expire(ctx, key, time.Second)
	}
	return count <= int64(r.rate)
}

// --- 中间件 ---

// RateLimiter 请求限流中间件
//
// 支持 memory（令牌桶）和 redis（滑动窗口）两种后端。
// 按客户端 IP 限流，超限返回 429。
func RateLimiter(enabled bool, backend string, r float64, burst int, redisClient *redis.Client, keyPrefix string) gin.HandlerFunc {
	if !enabled || r <= 0 {
		// 未启用或速率无效，返回空中间件
		return func(c *gin.Context) { c.Next() }
	}

	if burst <= 0 {
		burst = int(r)
	}

	var memBE *memoryBackend
	var redisBE *redisBackend

	switch backend {
	case "redis":
		if redisClient != nil {
			redisBE = newRedisBackend(redisClient, keyPrefix, r)
			applog.Info("ratelimit: redis backend", "rate", r, "key_prefix", keyPrefix)
		} else {
			// Redis 不可用，回退到内存
			memBE = newMemoryBackend(r, burst)
			applog.Warn("ratelimit: redis client nil, falling back to memory backend")
		}
	default:
		memBE = newMemoryBackend(r, burst)
		applog.Info("ratelimit: memory backend", "rate", r, "burst", burst)
	}

	return func(c *gin.Context) {
		if skipRateLimitPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		ip := c.ClientIP()
		var allowed bool
		if memBE != nil {
			allowed = memBE.allow(ip)
		} else {
			allowed = redisBE.allow(ip)
		}

		if !allowed {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    42900,
				"message": fmt.Sprintf("rate limit exceeded (max %v req/s)", r),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
