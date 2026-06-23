package cron

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"

	applog "github.com/example/dd-frame/pkg/log"
)

// --- 分布式锁接口 ---

// Locker 分布式锁抽象接口
//
// 用于多实例部署时防止同一任务被重复执行。
// Lock 返回 token（空字符串表示未获锁），Unlock 需传入该 token 以校验所有权。
type Locker interface {
	// Lock 尝试获取锁，返回 token 和 error。
	// token 为空字符串表示未获锁（被其他实例占用）。
	// key 为锁名称，ttl 为锁过期时间（防止死锁）。
	Lock(ctx context.Context, key string, ttl time.Duration) (token string, err error)

	// Unlock 释放锁，仅当 token 匹配时才删除（防止误删其他实例的锁）。
	Unlock(ctx context.Context, key string, token string) error

	// Close 释放锁实现内部资源（如后台清理 goroutine）。
	Close() error
}

// --- 内存锁实现（单实例） ---

type memoryLocker struct {
	mu    sync.Mutex
	locks map[string]memLockEntry // key -> {token, 过期时间}
	done  chan struct{}
}

type memLockEntry struct {
	token     string
	expiresAt time.Time
}

// NewMemoryLocker 创建内存锁实例（仅适用于单实例部署）
func NewMemoryLocker() Locker {
	ml := &memoryLocker{
		locks: make(map[string]memLockEntry),
		done:  make(chan struct{}),
	}
	// 定期清理过期锁
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ml.done:
				return
			case <-ticker.C:
				ml.mu.Lock()
				now := time.Now()
				for k, entry := range ml.locks {
					if now.After(entry.expiresAt) {
						delete(ml.locks, k)
					}
				}
				ml.mu.Unlock()
			}
		}
	}()
	return ml
}

func (m *memoryLocker) Close() error {
	close(m.done)
	return nil
}

func (m *memoryLocker) Lock(_ context.Context, key string, ttl time.Duration) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if entry, ok := m.locks[key]; ok && now.Before(entry.expiresAt) {
		return "", nil // 锁被占用
	}
	token := generateToken()
	m.locks[key] = memLockEntry{token: token, expiresAt: now.Add(ttl)}
	return token, nil
}

func (m *memoryLocker) Unlock(_ context.Context, key string, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.locks[key]; ok && entry.token == token {
		delete(m.locks, key)
	}
	return nil
}

// --- Redis 锁实现（多实例） ---

type redisLocker struct {
	client    *redis.Client
	keyPrefix string
}

// NewRedisLocker 创建基于 Redis 的分布式锁实例
func NewRedisLocker(client *redis.Client, keyPrefix string) Locker {
	if keyPrefix == "" {
		keyPrefix = "cron:lock:"
	}
	return &redisLocker{client: client, keyPrefix: keyPrefix}
}

func (r *redisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := generateToken()
	ok, err := r.client.SetNX(ctx, r.keyPrefix+key, token, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("redis lock failed: %w", err)
	}
	if !ok {
		return "", nil // 锁被占用
	}
	return token, nil
}

// unlockScript 原子校验 token 后删除，防止误删其他实例持有的锁
var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`)

func (r *redisLocker) Unlock(ctx context.Context, key string, token string) error {
	err := unlockScript.Run(ctx, r.client, []string{r.keyPrefix + key}, token).Err()
	if err != nil {
		return fmt.Errorf("redis unlock failed: %w", err)
	}
	return nil
}

func (r *redisLocker) Close() error { return nil }

// --- 调度器 ---

// Scheduler 定时任务调度器
//
// 封装 robfig/cron/v3，提供任务注册、分布式锁、生命周期管理。
type Scheduler struct {
	cron   *cron.Cron
	locker Locker
	jobs   []jobInfo
}

type jobInfo struct {
	Name     string
	Spec     string
	EntryID  cron.EntryID
}

// Config 调度器配置
type Config struct {
	Enabled   bool   `mapstructure:"enabled"`
	Locker    string `mapstructure:"locker"`    // memory / redis
	KeyPrefix string `mapstructure:"key_prefix"` // Redis 锁 key 前缀
}

// NewScheduler 创建定时任务调度器
func NewScheduler(cfg Config, redisClient *redis.Client) *Scheduler {
	c := cron.New(cron.WithSeconds()) // 支持秒级精度

	var locker Locker
	switch cfg.Locker {
	case "redis":
		if redisClient != nil {
			locker = NewRedisLocker(redisClient, cfg.KeyPrefix)
			applog.Info("cron: redis locker", "key_prefix", cfg.KeyPrefix)
		} else {
			locker = NewMemoryLocker()
			applog.Warn("cron: redis client nil, falling back to memory locker")
		}
	default:
		locker = NewMemoryLocker()
		applog.Info("cron: memory locker")
	}

	return &Scheduler{
		cron:   c,
		locker: locker,
	}
}

// AddJob 注册定时任务
//
// name: 任务唯一名称（用于锁 key 和日志）
// spec: cron 表达式（6 段，含秒），如 "0 */5 * * * *" 表示每 5 分钟
// ttl: 锁过期时间，通常设为任务最大执行时间
// fn: 任务函数
func (s *Scheduler) AddJob(name, spec string, ttl time.Duration, fn func(ctx context.Context)) error {
	lockKey := name
	entryID, err := s.cron.AddFunc(spec, func() {
		ctx := context.Background()

		// 尝试获取分布式锁
		token, err := s.locker.Lock(ctx, lockKey, ttl)
		if err != nil {
			applog.Error("cron: lock error", "job", name, "err", err)
			return
		}
		if token == "" {
			applog.Debug("cron: job skipped (locked by another instance)", "job", name)
			return
		}
		defer func() {
			if err := s.locker.Unlock(ctx, lockKey, token); err != nil {
				applog.Warn("cron: unlock error", "job", name, "err", err)
			}
		}()

		start := time.Now()
		applog.Info("cron: job started", "job", name)
		fn(ctx)
		applog.Info("cron: job finished", "job", name, "duration", time.Since(start).String())
	})
	if err != nil {
		return fmt.Errorf("add cron job '%s' failed: %w", name, err)
	}

	s.jobs = append(s.jobs, jobInfo{Name: name, Spec: spec, EntryID: entryID})
	applog.Info("cron: job registered", "job", name, "spec", spec)
	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
	applog.Info("cron: scheduler started", "jobs", len(s.jobs))
}

// Stop 停止调度器，等待正在执行的任务完成（最多 30 秒超时）
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		applog.Info("cron: scheduler stopped")
	case <-timer.C:
		applog.Warn("cron: scheduler stop timed out, some jobs may still be running")
	}
	// 释放锁实现内部资源
	if err := s.locker.Close(); err != nil {
		applog.Warn("cron: locker close error", "err", err)
	}
}

// Jobs 返回已注册的任务列表
func (s *Scheduler) Jobs() []jobInfo {
	return s.jobs
}

// generateToken 生成 16 字节随机 token（32 位十六进制字符串）
func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
