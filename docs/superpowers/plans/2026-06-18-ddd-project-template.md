# DDD 模块化单体项目框架 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将当前分层单体 DDD 骨架重构为模块化单体架构，集成 Gin + Connect gRPC + GORM + Redis + Viper + Zap 基础设施，形成生产可用的项目模板。

**Architecture:** 模块化单体，每个限界上下文位于 `internal/` 下拥有独立 DDD 分层（domain → model → biz → service → api），通过端口接口 + ACL 防腐层实现跨模块解耦，app/ 作为 Composition Root 装配所有依赖。

**Tech Stack:** Go 1.26, Gin, Connect RPC, GORM, go-redis, Viper, Zap

---

## File Structure

### 新建文件

| 文件 | 职责 |
|------|------|
| `.gitignore` | Go 项目忽略规则 |
| `config/config.yaml` | 默认配置 |
| `config/config.example.yaml` | 配置示例（不含敏感信息） |
| `app/config.go` | Viper 配置加载 |
| `app/logger.go` | Zap 日志初始化 |
| `app/database.go` | GORM 数据库初始化 |
| `app/cache.go` | Redis 初始化 |
| `app/server.go` | Gin HTTP + Connect gRPC 服务器 |
| `app/wire.go` | IoC 总装配 |
| `pkg/response/response.go` | 统一 HTTP 响应格式 |
| `pkg/errors/errors.go` | 统一错误处理（重写） |
| `pkg/pagination/pagination.go` | 分页工具 |
| `pkg/log/log.go` | 结构化日志封装（重写） |
| `middleware/recovery.go` | Panic 恢复中间件 |
| `middleware/cors.go` | CORS 中间件 |
| `middleware/request_id.go` | 请求 ID 中间件 |
| `middleware/logger.go` | 请求日志中间件 |
| `middleware/auth.go` | JWT 鉴权中间件 |
| `internal/_template/domain/entity.go` | 模块模板 - 聚合根 |
| `internal/_template/domain/value_object.go` | 模块模板 - 值对象 |
| `internal/_template/domain/enums.go` | 模块模板 - 枚举 |
| `internal/_template/domain/errors.go` | 模块模板 - 领域错误 |
| `internal/_template/domain/service.go` | 模块模板 - 领域服务 |
| `internal/_template/biz/service.go` | 模块模板 - 应用服务 |
| `internal/_template/biz/ports.go` | 模块模板 - 端口接口 |
| `internal/_template/service/app_service.go` | 模块模板 - 边界层 |
| `internal/_template/api/http_handler.go` | 模块模板 - HTTP handler |
| `internal/_template/api/grpc_handler.go` | 模块模板 - gRPC handler |
| `internal/_template/model/repo.go` | 模块模板 - 仓储接口 |
| `internal/_template/model/dao.go` | 模块模板 - 仓储实现 |
| `internal/_template/model/cache.go` | 模块模板 - 缓存实现 |
| `internal/_template/wire.go` | 模块模板 - IoC 装配 |
| `example/order/domain/entity.go` | 订单示例 - 聚合根 |
| `example/order/domain/value_object.go` | 订单示例 - Money 值对象 |
| `example/order/domain/enums.go` | 订单示例 - 枚举 |
| `example/order/domain/errors.go` | 订单示例 - 领域错误 |
| `example/order/domain/service.go` | 订单示例 - 领域服务接口 |
| `example/order/biz/service.go` | 订单示例 - 应用服务 |
| `example/order/biz/ports.go` | 订单示例 - 端口接口 |
| `example/order/service/app_service.go` | 订单示例 - 边界层 |
| `example/order/api/http_handler.go` | 订单示例 - HTTP handler |
| `example/order/model/repo.go` | 订单示例 - 仓储接口 |
| `example/order/model/dao.go` | 订单示例 - 仓储实现 |
| `example/order/model/cache.go` | 订单示例 - 缓存 |
| `example/order/wire.go` | 订单示例 - IoC |

### 删除文件/目录

| 路径 | 原因 |
|------|------|
| `api/` | 迁移到 `example/order/api/` |
| `biz/` | 迁移到 `example/order/biz/` |
| `domain/` | 迁移到 `example/order/domain/` |
| `service/` | 迁移到 `example/order/service/` |
| `model/` | 迁移到 `example/order/model/` |
| `ioc/` | 迁移到 `example/order/wire.go` |
| `router/` | 路由注册移入 `example/order/api/` |
| `core/` | 被 `app/` 替代 |
| `global/` | 被 `app/` 替代 |
| `component/` | 示例组件，不再需要 |
| `config/config.go` | 被 `app/config.go` 替代 |

---

### Task 1: 项目基础 — Git 初始化 & 清理旧结构

**Files:**
- Create: `.gitignore`
- Delete: `api/`, `biz/`, `domain/`, `service/`, `model/`, `ioc/`, `router/`, `core/`, `global/`, `component/`, `config/config.go`

- [ ] **Step 1: 创建 .gitignore**

```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
dd-frame

# Test binary
*.test

# Output of go coverage
*.out
*.prof

# Dependency directories
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Config (sensitive)
config/config.yaml

# Buf generated
proto/gen/

# Logs
*.log
logs/
```

- [ ] **Step 2: 删除旧目录结构**

```bash
cd /Users/mayujian/all_code/go_code/dd-frame
rm -rf api/ biz/ domain/ service/ model/ ioc/ router/ core/ global/ component/ config/config.go
```

- [ ] **Step 3: 创建新目录骨架**

```bash
mkdir -p app internal/_template/domain internal/_template/biz internal/_template/service internal/_template/api internal/_template/model pkg/response pkg/pagination pkg/errors pkg/log middleware proto example/order/domain example/order/biz example/order/service example/order/api example/order/model config migrations
```

- [ ] **Step 4: 初始化 Git**

```bash
git init
git add .gitignore
git commit -m "chore: initialize project with new modular monolith structure"
```

- [ ] **Step 5: 验证**

```bash
find . -type d | grep -v '.git' | sort
```

Expected: 新目录结构完整显示，旧目录已清除。

---

### Task 2: 共享工具包 — pkg/errors & pkg/log & pkg/response & pkg/pagination

**Files:**
- Create: `pkg/errors/errors.go`
- Create: `pkg/log/log.go`
- Create: `pkg/response/response.go`
- Create: `pkg/pagination/pagination.go`

- [ ] **Step 1: 创建 pkg/errors/errors.go**

```go
package errors

import (
	"fmt"
	"runtime"
)

// AppError 应用层统一错误结构
type AppError struct {
	Code    int    // 业务错误码
	Message string // 用户可见的错误消息
	Reason  string // 内部原因
	Stack   string // 调用栈
}

func (e *AppError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Reason)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建应用错误
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WithReason 附加内部原因
func (e *AppError) WithReason(reason string) *AppError {
	e.Reason = reason
	return e
}

// WithStack 附加调用栈
func (e *AppError) WithStack() *AppError {
	_, file, line, _ := runtime.Caller(1)
	e.Stack = fmt.Sprintf("%s:%d", file, line)
	return e
}

// Wrap 包装已有错误
func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Reason:  err.Error(),
	}
}
```

- [ ] **Step 2: 创建 pkg/log/log.go**

```go
package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局日志实例
var Logger *zap.SugaredLogger

// Init 初始化日志
func Init(level string, format string) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	var cfg zap.Config
	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := cfg.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	Logger = logger.Sugar()
}

// Debug 调试日志
func Debug(msg string, keysAndValues ...interface{}) {
	Logger.Debugw(msg, keysAndValues...)
}

// Info 信息日志
func Info(msg string, keysAndValues ...interface{}) {
	Logger.Infow(msg, keysAndValues...)
}

// Warn 警告日志
func Warn(msg string, keysAndValues ...interface{}) {
	Logger.Warnw(msg, keysAndValues...)
}

// Error 错误日志
func Error(msg string, keysAndValues ...interface{}) {
	Logger.Errorw(msg, keysAndValues...)
}

// WithContext 从 context 提取字段（预留扩展）
func WithContext(_ context.Context) *zap.SugaredLogger {
	return Logger
}

// Sync 刷新日志缓冲
func Sync() {
	_ = Logger.Sync()
}
```

- [ ] **Step 3: 创建 pkg/response/response.go**

```go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperr "github.com/example/dd-frame/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: msg,
	})
}

// FromError 从 AppError 生成响应
func FromError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperr.AppError); ok {
		httpStatus := mapCodeToHTTP(appErr.Code)
		c.JSON(httpStatus, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    50000,
		Message: "internal server error",
	})
}

// mapCodeToHTTP 业务错误码 → HTTP 状态码
func mapCodeToHTTP(code int) int {
	switch {
	case code >= 40000 && code < 50000:
		return http.StatusBadRequest
	case code >= 40100 && code < 40200:
		return http.StatusUnauthorized
	case code >= 40300 && code < 40400:
		return http.StatusForbidden
	case code >= 40400 && code < 40500:
		return http.StatusNotFound
	case code >= 40900 && code < 41000:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
```

- [ ] **Step 4: 创建 pkg/pagination/pagination.go**

```go
package pagination

// Request 分页请求参数
type Request struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	OrderBy  string `form:"orderBy" json:"orderBy"`
	Sort     string `form:"sort" json:"sort"` // asc / desc
}

// Response 分页响应结构
type Response struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Pages    int         `json:"pages"`
}

// Normalize 标准化分页参数（防越界）
func (r *Request) Normalize() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		r.PageSize = 20
	}
	if r.OrderBy == "" {
		r.OrderBy = "id"
	}
	if r.Sort == "" {
		r.Sort = "desc"
	}
}

// Offset 计算 SQL OFFSET
func (r *Request) Offset() int {
	return (r.Page - 1) * r.PageSize
}

// NewResponse 创建分页响应
func NewResponse(list interface{}, total int64, req *Request) *Response {
	pages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		pages++
	}
	return &Response{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Pages:    pages,
	}
}
```

- [ ] **Step 5: 提交**

```bash
git add pkg/
git commit -m "feat: add shared packages (errors, log, response, pagination)"
```

---

### Task 3: 配置管理 — config.yaml & app/config.go

**Files:**
- Create: `config/config.yaml`
- Create: `config/config.example.yaml`
- Create: `app/config.go`

- [ ] **Step 1: 创建 config/config.example.yaml**

```yaml
server:
  http_port: 8080
  grpc_port: 8081
  mode: debug  # debug / release / test

database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  dbname: ddframe
  user: root
  password: ""

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

jwt:
  secret: "change-me-in-production"
  expires_in: 24  # hours

log:
  level: debug    # debug / info / warn / error
  format: console # json / console
```

- [ ] **Step 2: 复制为 config/config.yaml**

```bash
cp config/config.example.yaml config/config.yaml
```

- [ ] **Step 3: 创建 app/config.go**

```go
package app

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用总配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTPPort int    `mapstructure:"http_port"`
	GRPCPort int    `mapstructure:"grpc_port"`
	Mode     string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DBName   string `mapstructure:"dbname"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig JWT 鉴权配置
type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// GlobalConfig 全局配置实例（启动时加载一次）
var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 支持环境变量覆盖（前缀 DD_）
	viper.SetEnvPrefix("DD")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}
```

- [ ] **Step 4: 提交**

```bash
git add config/ app/config.go
git commit -m "feat: add config management with Viper"
```

---

### Task 4: 基础设施 — app/logger.go & app/database.go & app/cache.go

**Files:**
- Create: `app/logger.go`
- Create: `app/database.go`
- Create: `app/cache.go`

- [ ] **Step 1: 创建 app/logger.go**

```go
package app

import (
	"github.com/example/dd-frame/pkg/log"
)

// InitLogger 初始化结构化日志
func InitLogger(cfg *LogConfig) {
	log.Init(cfg.Level, cfg.Format)
	log.Info("logger initialized", "level", cfg.Level, "format", cfg.Format)
}
```

- [ ] **Step 2: 创建 app/database.go**

```go
package app

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	applog "github.com/example/dd-frame/pkg/log"
)

// GlobalDB 全局数据库实例
var GlobalDB *gorm.DB

// InitDatabase 初始化 GORM 数据库连接
func InitDatabase(cfg *DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var gormLogLevel logger.LogLevel
	if GlobalConfig.Server.Mode == "debug" {
		gormLogLevel = logger.Info
	} else {
		gormLogLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB failed: %w", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)

	GlobalDB = db
	applog.Info("database connected", "driver", cfg.Driver, "host", cfg.Host, "dbname", cfg.DBName)
	return db, nil
}
```

- [ ] **Step 3: 创建 app/cache.go**

```go
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
```

- [ ] **Step 4: 提交**

```bash
git add app/logger.go app/database.go app/cache.go
git commit -m "feat: add logger (Zap), database (GORM), cache (Redis) initialization"
```

---

### Task 5: 服务器 — app/server.go & main.go

**Files:**
- Create: `app/server.go`
- Modify: `main.go`

- [ ] **Step 1: 创建 app/server.go**

```go
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/middleware"
	applog "github.com/example/dd-frame/pkg/log"
)

// RunServer 启动 Gin HTTP 服务器（gRPC 预留）
func RunServer(cfg *Config, router *gin.Engine) {
	// 应用全局中间件
	router.Use(
		middleware.Recovery(),
		middleware.Cors(),
		middleware.RequestID(),
		middleware.Logger(),
	)

	httpAddr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	// 优雅关闭
	go func() {
		applog.Info("HTTP server starting", "addr", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			applog.Error("HTTP server error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	applog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		applog.Error("server forced to shutdown", "err", err)
	}
	applog.Info("server exited")
}
```

- [ ] **Step 2: 重写 main.go**

```go
package main

import (
	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/app"
	applog "github.com/example/dd-frame/pkg/log"
)

func main() {
	// 1. 加载配置
	cfg, err := app.LoadConfig("config/config.yaml")
	if err != nil {
		panic("load config failed: " + err.Error())
	}

	// 2. 初始化日志
	app.InitLogger(&cfg.Log)
	defer applog.Sync()

	// 3. 初始化数据库（无 DB 时跳过）
	// db, err := app.InitDatabase(&cfg.Database)

	// 4. 初始化 Redis（无 Redis 时跳过）
	// rdb, err := app.InitRedis(&cfg.Redis)

	// 5. 装配模块
	router := app.Wire(cfg)

	// 6. 启动服务器
	app.RunServer(cfg, router)
}

// initRouter 创建空的 Gin 引擎（Wire 中注册路由）
func initRouter() *gin.Engine {
	if app.GlobalConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}
```

- [ ] **Step 3: 提交**

```bash
git add app/server.go main.go
git commit -m "feat: add HTTP server with graceful shutdown and main entry"
```

---

### Task 6: 中间件 — middleware/

**Files:**
- Create: `middleware/recovery.go`
- Create: `middleware/cors.go`
- Create: `middleware/request_id.go`
- Create: `middleware/logger.go`
- Create: `middleware/auth.go`

- [ ] **Step 1: 创建 middleware/recovery.go**

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/log"
	"github.com/example/dd-frame/pkg/response"
)

// Recovery Panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", "err", r, "path", c.Request.URL.Path)
				response.Error(c, http.StatusInternalServerError, 50000, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

- [ ] **Step 2: 创建 middleware/cors.go**

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors CORS 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: 创建 middleware/request_id.go**

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// RequestID 请求 ID 中间件（链路追踪）
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(requestIDHeader)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("request_id", rid)
		c.Header(requestIDHeader, rid)
		c.Next()
	}
}
```

- [ ] **Step 4: 创建 middleware/logger.go**

```go
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
```

- [ ] **Step 5: 创建 middleware/auth.go**

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/response"
)

// Auth JWT 鉴权中间件
//
// 从 Authorization header 提取 Bearer token 并验证。
// 验证成功后将 userID 和 companyID 注入 gin.Context。
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, 40100, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, 40101, "invalid authorization format")
			c.Abort()
			return
		}

		token := parts[1]
		// TODO: 实际项目中替换为 JWT 解析逻辑
		// claims, err := jwt.ParseToken(token, secret)
		// c.Set("userID", claims.UserID)
		// c.Set("companyID", claims.CompanyID)
		_ = token
		_ = secret

		c.Next()
	}
}
```

- [ ] **Step 6: 提交**

```bash
git add middleware/
git commit -m "feat: add middleware (recovery, cors, request_id, logger, auth)"
```

---

### Task 7: 模块模板 — internal/_template/

**Files:**
- Create: `internal/_template/domain/entity.go`
- Create: `internal/_template/domain/value_object.go`
- Create: `internal/_template/domain/enums.go`
- Create: `internal/_template/domain/errors.go`
- Create: `internal/_template/domain/service.go`
- Create: `internal/_template/biz/service.go`
- Create: `internal/_template/biz/ports.go`
- Create: `internal/_template/service/app_service.go`
- Create: `internal/_template/api/http_handler.go`
- Create: `internal/_template/api/grpc_handler.go`
- Create: `internal/_template/model/repo.go`
- Create: `internal/_template/model/dao.go`
- Create: `internal/_template/model/cache.go`
- Create: `internal/_template/wire.go`

- [ ] **Step 1: 创建 domain 层模板文件**

创建 `internal/_template/domain/entity.go`:
```go
// Package domain 包含聚合根、实体定义。
// 领域层零外部依赖，仅使用标准库。
//
// 使用方式：复制 _template/ 为 internal/{module}/，替换 EntityName。
package domain

import "time"

// EntityName 聚合根（替换为实际业务名称，如 Order、Payment）
//
// 聚合根是外部访问聚合内部对象的唯一入口。
// 所有对聚合内部状态的修改必须通过聚合根方法完成。
type EntityName struct {
	ID        int64     // 主键 ID
	Status    int       // 业务状态（使用枚举常量）
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// EntityNameItem 聚合内实体（替换为实际名称，如 OrderItem）
type EntityNameItem struct {
	ID int64 // 主键 ID
}

// DoAction 聚合根行为方法（替换为实际业务操作）
//
// 业务规则内聚在聚合根方法中，不要在 biz 层写业务逻辑。
func (e *EntityName) DoAction() error {
	// 1. 业务校验
	// if e.Status != ValidStatus { return ErrInvalidStatus }
	// 2. 状态流转
	// e.Status = NextStatus
	return nil
}
```

创建 `internal/_template/domain/value_object.go`:
```go
package domain

// Money 金额值对象示例（替换为实际值对象）
//
// 值对象特征：不可变，通过属性值判等，操作方法返回新实例。
type Money int64

// NewMoney 创建金额
func NewMoney(cents int64) Money {
	return Money(cents)
}

// Add 金额相加（返回新值对象）
func (m Money) Add(other Money) Money {
	return Money(int64(m) + int64(other))
}
```

创建 `internal/_template/domain/enums.go`:
```go
package domain

// StatusEnum 状态枚举示例（替换为实际枚举）
//
// 每个枚举必须包含 IsValid() 和 String() 方法。
type StatusEnum int

const (
	StatusA StatusEnum = 0 // 状态A
	StatusB StatusEnum = 1 // 状态B
)

// IsValid 校验枚举值是否合法
func (s StatusEnum) IsValid() bool {
	return s >= StatusA && s <= StatusB
}

// String 返回描述
func (s StatusEnum) String() string {
	switch s {
	case StatusA:
		return "状态A"
	case StatusB:
		return "状态B"
	default:
		return "未知"
	}
}
```

创建 `internal/_template/domain/errors.go`:
```go
package domain

import "fmt"

// 领域错误 reason 常量
const (
	ReasonNotFound      = "ENTITY_NOT_FOUND"
	ReasonStatusInvalid = "ENTITY_STATUS_INVALID"
)

// ErrNotFound 实体不存在
func ErrNotFound(id string) error {
	return fmt.Errorf("[%s] entity not found: %s", ReasonNotFound, id)
}

// ErrStatusInvalid 状态不合法
func ErrStatusInvalid(current int, action string) error {
	return fmt.Errorf("[%s] cannot %s in status %d", ReasonStatusInvalid, action, current)
}
```

创建 `internal/_template/domain/service.go`:
```go
package domain

// DomainServiceInterface 领域服务接口示例
//
// 当逻辑不适合放在单个聚合根上时（跨聚合操作），使用领域服务。
type DomainServiceInterface interface {
	// DoSomething 执行跨聚合操作
	DoSomething(param string) (string, error)
}
```

- [ ] **Step 2: 创建 biz 层模板文件**

创建 `internal/_template/biz/service.go`:
```go
package biz

import (
	"context"
	"fmt"

	"github.com/example/dd-frame/internal/_template/domain"
	"github.com/example/dd-frame/internal/_template/model"
)

// EntityService 应用服务接口（替换为实际名称）
type EntityService interface {
	CreateEntity(ctx context.Context, req *CreateRequest) (*domain.EntityName, error)
}

// CreateRequest 创建请求 DTO
type CreateRequest struct {
	// 业务字段
}

// entityService 应用服务实现
type entityService struct {
	repo model.EntityRepo
	// 其他端口接口...
}

// NewEntityService 创建应用服务
func NewEntityService(repo model.EntityRepo) EntityService {
	return &entityService{repo: repo}
}

// CreateEntity 创建实体用例编排
func (s *entityService) CreateEntity(ctx context.Context, req *CreateRequest) (*domain.EntityName, error) {
	// 1. 构建聚合根
	entity := &domain.EntityName{}

	// 2. 调用聚合根方法（业务规则在领域层）
	if err := entity.DoAction(); err != nil {
		return nil, err
	}

	// 3. 持久化
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("save entity failed: %w", err)
	}

	return entity, nil
}
```

创建 `internal/_template/biz/ports.go`:
```go
package biz

import "context"

// ExternalServiceClient 出站端口接口示例
//
// 定义对外部系统的依赖，由 ACL 防腐层或 infrastructure 实现。
type ExternalServiceClient interface {
	// CallExternal 调用外部服务
	CallExternal(ctx context.Context, param string) (string, error)
}
```

- [ ] **Step 3: 创建 service 层模板文件**

创建 `internal/_template/service/app_service.go`:
```go
package service

import (
	"context"

	"github.com/example/dd-frame/internal/_template/biz"
)

// EntityAppService 应用边界层（DTO 转换）
//
// 负责 HTTP/gRPC DTO ↔ 业务 DTO 的转换，调用 biz 层。
type EntityAppService struct {
	usecase biz.EntityService
}

// NewEntityAppService 创建应用边界服务
func NewEntityAppService(usecase biz.EntityService) *EntityAppService {
	return &EntityAppService{usecase: usecase}
}

// CreateInput HTTP 入参 DTO
type CreateInput struct {
	// HTTP 请求字段
}

// CreateOutput HTTP 出参 DTO
type CreateOutput struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// CreateEntity 创建实体（HTTP 入口）
func (s *EntityAppService) CreateEntity(ctx context.Context, input *CreateInput) (*CreateOutput, error) {
	// 1. HTTP DTO → 业务 DTO
	bizReq := &biz.CreateRequest{}

	// 2. 调用 biz 层
	entity, err := s.usecase.CreateEntity(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	// 3. 领域对象 → HTTP DTO
	return &CreateOutput{
		ID: entity.ID,
	}, nil
}
```

- [ ] **Step 4: 创建 api 层模板文件**

创建 `internal/_template/api/http_handler.go`:
```go
package api

import (
	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/internal/_template/service"
	"github.com/example/dd-frame/pkg/response"
)

// EntityAPI HTTP Handler（替换为实际名称）
type EntityAPI struct {
	svc *service.EntityAppService
}

// NewEntityAPI 创建 API handler
func NewEntityAPI(svc *service.EntityAppService) *EntityAPI {
	return &EntityAPI{svc: svc}
}

// RegisterRoutes 注册模块路由
func (a *EntityAPI) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/entity")
	{
		group.POST("", a.CreateHandler)
		// group.GET("/:id", a.GetHandler)
		// group.POST("/:id/submit", a.SubmitHandler)
	}
}

// CreateHandler 创建实体 handler
func (a *EntityAPI) CreateHandler(c *gin.Context) {
	var input service.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, 40000, "invalid request body")
		return
	}

	output, err := a.svc.CreateEntity(c.Request.Context(), &input)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, output)
}
```

创建 `internal/_template/api/grpc_handler.go`:
```go
package api

// GRPCHandler Connect RPC handler 模板
//
// 实际使用时实现 proto 生成的接口：
//
// type EntityGRPCHandler struct {
//     svc *service.EntityAppService
// }
//
// func (h *EntityGRPCHandler) CreateEntity(
//     ctx context.Context,
//     req *connect.Request[pb.CreateEntityRequest],
// ) (*connect.Response[pb.CreateEntityResponse], error) {
//     // 1. proto DTO → service DTO
//     // 2. 调用 svc
//     // 3. service DTO → proto DTO
// }
```

- [ ] **Step 5: 创建 model 层模板文件**

创建 `internal/_template/model/repo.go`:
```go
package model

import (
	"context"

	"github.com/example/dd-frame/internal/_template/domain"
)

// EntityRepo 仓储接口
//
// 使用领域对象，不暴露 DB 模型。方法命名体现领域语义。
type EntityRepo interface {
	Create(ctx context.Context, entity *domain.EntityName) error
	QueryByID(ctx context.Context, id int64) (*domain.EntityName, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
}
```

创建 `internal/_template/model/dao.go`:
```go
package model

import (
	"context"
	"time"

	"github.com/example/dd-frame/internal/_template/domain"
)

// 编译期校验：确保 DAO 实现了 Repo 接口
var _ EntityRepo = (*EntityDAO)(nil)

// EntityModel DB 表模型
type EntityModel struct {
	ID        int64     `gorm:"primary_key;auto_increment"`
	Status    int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName 表名
func (EntityModel) TableName() string {
	return "t_entity" // 替换为实际表名
}

// EntityDAO 仓储 GORM 实现
type EntityDAO struct {
	// db *gorm.DB // 注入 GORM 实例
}

// NewEntityDAO 创建 DAO
func NewEntityDAO() *EntityDAO {
	return &EntityDAO{}
}

func (d *EntityDAO) Create(_ context.Context, entity *domain.EntityName) error {
	// model := entityToModel(entity)
	// d.db.Create(&model)
	// entity.ID = model.ID
	return nil
}

func (d *EntityDAO) QueryByID(_ context.Context, _ int64) (*domain.EntityName, error) {
	// var model EntityModel
	// d.db.First(&model, id)
	// return modelToEntity(&model), nil
	return nil, nil
}

func (d *EntityDAO) UpdateStatus(_ context.Context, _ int64, _ int) error {
	return nil
}

// entityToModel 领域对象 → DB 模型
func entityToModel(e *domain.EntityName) *EntityModel {
	return &EntityModel{ID: e.ID, Status: e.Status}
}

// modelToEntity DB 模型 → 领域对象
func modelToEntity(m *EntityModel) *domain.EntityName {
	if m == nil {
		return nil
	}
	return &domain.EntityName{ID: m.ID, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}
```

创建 `internal/_template/model/cache.go`:
```go
package model

import "context"

// EntityCache 缓存接口
type EntityCache interface {
	SetLock(ctx context.Context, key string, ttl int) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, data string, ttl int) error
}
```

- [ ] **Step 6: 创建 wire.go 模板**

创建 `internal/_template/wire.go`:
```go
package _template

import (
	"github.com/example/dd-frame/internal/_template/api"
	"github.com/example/dd-frame/internal/_template/biz"
	"github.com/example/dd-frame/internal/_template/model"
	"github.com/example/dd-frame/internal/_template/service"
)

// Wire 模块内 IoC 装配
//
// 创建本模块所有依赖并返回 API handler。
// 在 app/wire.go 中调用此函数注册模块。
func Wire() *api.EntityAPI {
	// 1. 数据层
	repo := model.NewEntityDAO()

	// 2. 业务编排层
	svc := biz.NewEntityService(repo)

	// 3. 应用边界层
	appSvc := service.NewEntityAppService(svc)

	// 4. API 层
	return api.NewEntityAPI(appSvc)
}
```

- [ ] **Step 7: 提交**

```bash
git add internal/_template/
git commit -m "feat: add module template with complete DDD layered skeleton"
```

---

### Task 8: 订单示例 — example/order/

**Files:**
- Create: `example/order/domain/*.go` (5 files)
- Create: `example/order/biz/*.go` (2 files)
- Create: `example/order/service/app_service.go`
- Create: `example/order/api/http_handler.go`
- Create: `example/order/model/*.go` (3 files)
- Create: `example/order/wire.go`

- [ ] **Step 1: 创建 example/order/domain/ 层**

迁移当前 `domain/order/` 的代码到新路径，保持原有逻辑不变，文件重命名为：
- `entity.go` ← 原 `order.go`（聚合根 + OrderItem）
- `value_object.go` ← 原 `money.go`（Money + OptionalMoney）
- `enums.go` ← 原 `enums.go`（OrderStatus + PaymentType + OrderType）
- `errors.go` ← 原 `errors.go`（领域错误）
- `service.go` ← 原 `service.go`（领域服务接口）

package 改为 `package order`（保持不变）。

- [ ] **Step 2: 创建 example/order/model/ 层**

迁移并适配：
- `repo.go` ← 原 `model/order/repo/order_repo.go`（仓储接口，package 改为 `package model`）
- `dao.go` ← 原 `model/order/data/dao/order_dao.go`（仓储实现，更新 import 路径）
- `cache.go` ← 原 `model/order/data/cache/order_cache.go`（缓存接口）

所有 import 路径更新为 `github.com/example/dd-frame/example/order/domain`。

- [ ] **Step 3: 创建 example/order/biz/ 层**

迁移并适配：
- `service.go` ← 原 `biz/order/orderservice/order_service.go`（应用服务，package 改为 `package biz`）
- `ports.go` ← 原 `biz/order/common/deps/ports.go`（端口接口）

更新 import 路径指向 `example/order/domain` 和 `example/order/model`。

- [ ] **Step 4: 创建 example/order/service/ & api/ 层**

- `service/app_service.go` ← 原 `service/order/order_app_service.go`（更新 import）
- `api/http_handler.go` ← 原 `api/v1/order/order_api.go`（改为 Gin handler，使用 pkg/response）

- [ ] **Step 5: 创建 example/order/wire.go**

```go
package order

import (
	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/example/order/api"
	"github.com/example/dd-frame/example/order/biz"
	"github.com/example/dd-frame/example/order/model"
	"github.com/example/dd-frame/example/order/service"
)

// Wire 订单模块 IoC 装配
func Wire() *api.OrderAPI {
	repo := model.NewOrderDAO()
	svc := biz.NewOrderService(repo, nil, nil, nil, nil) // mock 依赖
	appSvc := service.NewOrderAppService(svc)
	return api.NewOrderAPI(appSvc)
}

// RegisterRoutes 注册订单路由
func RegisterRoutes(rg *gin.RouterGroup, orderAPI *api.OrderAPI) {
	orderAPI.RegisterRoutes(rg)
}
```

- [ ] **Step 6: 编译验证**

```bash
go build ./...
```

Expected: 编译通过，无错误。

- [ ] **Step 7: 提交**

```bash
git add example/
git commit -m "feat: add order example with complete DDD implementation"
```

---

### Task 9: IoC 总装配 — app/wire.go

**Files:**
- Create: `app/wire.go`

- [ ] **Step 1: 创建 app/wire.go**

```go
package app

import (
	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/middleware"
)

// Wire IoC 总装配（Composition Root）
//
// 装配所有模块依赖，注册路由，返回 Gin Engine。
// 这是整个应用唯一"知道所有模块"的地方。
func Wire(cfg *Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// === 模块装配 ===
	// 示例：取消注释以启用订单模块
	// orderAPI := order_example.Wire()
	// order_example.RegisterRoutes(router.Group("/api/v1"), orderAPI)

	// === 健康检查 ===
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// === 鉴权路由组示例 ===
	// authGroup := router.Group("/api/v1")
	// authGroup.Use(middleware.Auth(cfg.JWT.Secret))
	// {
	//     order_example.RegisterRoutes(authGroup, orderAPI)
	// }

	_ = middleware.Auth // 确保 middleware 包被引用

	return router
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./...
```

Expected: 编译通过。

- [ ] **Step 3: 提交**

```bash
git add app/wire.go
git commit -m "feat: add IoC composition root (app/wire.go)"
```

---

### Task 10: 依赖安装 & 最终验证

- [ ] **Step 1: 安装所有 Go 依赖**

```bash
cd /Users/mayujian/all_code/go_code/dd-frame
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/spf13/viper
go get go.uber.org/zap
go get github.com/google/uuid
go get connectrpc.com/connect-go
go mod tidy
```

- [ ] **Step 2: 编译验证**

```bash
go build ./...
```

Expected: 零错误。

- [ ] **Step 3: 静态检查**

```bash
go vet ./...
```

Expected: 零警告。

- [ ] **Step 4: 启动验证**

```bash
go run main.go &
sleep 2
curl http://localhost:8080/health
kill %1
```

Expected: `{"status":"ok"}`

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "feat: complete DDD modular monolith project template"
```

- [ ] **Step 6: 验证最终目录结构**

```bash
find . -type f -name "*.go" | grep -v vendor | sort
```

Expected: 所有新建文件均存在，旧目录结构已清除。
