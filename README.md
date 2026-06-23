# dd-frame

基于 **DDD + 六边形架构**的 Go 语言**模块化单体**项目框架。

每个业务模块拥有独立的 DDD 分层，通过 Go `internal/` 实现编译器级别的模块隔离。

> 架构原理与设计详解 → [docs/architecture.md](docs/architecture.md)

## 快速开始

### 方式一：本地开发

```bash
git clone https://github.com/myouj/dd-frame.git
cd dd-frame

# 复制配置模板，按需修改
cp config/config.example.yaml config/config.yaml
vi config/config.yaml

# 编译 & 运行
make build    # 编译到 bin/dd-frame
make run      # 运行（HTTP :8080，自动检查端口占用）
```

### 方式二：Docker 一键启动（推荐）

```bash
# 启动完整开发环境（app + MySQL + Redis + Jaeger）
make docker-up

# 查看日志
make docker-logs

# 停止
make docker-down
```

启动后访问：

| 服务 | 地址 |
|------|------|
| 应用 API | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Jaeger UI | http://localhost:16686 |
| Prometheus 指标 | http://localhost:8080/metrics |
| 健康检查 | http://localhost:8080/health |

### 数据库初始化

```bash
make db-init    # 数据库迁移（创建/更新表结构）
make db-seed    # 初始化种子数据（admin 角色/用户/权限）
```

### 测试接口（订单示例）

```bash
# 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 创建订单
curl -X POST http://localhost:8080/api/v1/order \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"customerId": 1001, "items": [{"productId": 1, "quantity": 2, "unitPrice": 9900}]}'

# 提交订单
curl -X POST http://localhost:8080/api/v1/order/ORD-xxx/submit \
  -H "Authorization: Bearer $TOKEN"

# 取消订单
curl -X POST http://localhost:8080/api/v1/order/ORD-xxx/cancel \
  -H "Authorization: Bearer $TOKEN"
```

## 目录结构

```
dd-frame/
├── main.go                      # 应用入口
├── Makefile                     # 构建管理
├── Dockerfile                   # 多阶段构建
├── docker-compose.yaml          # 开发环境编排
│
├── app/                         # 基础设施 & 全局装配
│   ├── config.go                #   配置加载（Viper）
│   ├── server.go                #   HTTP 服务器（Gin + 优雅关闭）
│   ├── database.go              #   数据库（GORM + MySQL，未配置时跳过）
│   ├── cache.go                 #   缓存（Redis，未配置时跳过）
│   ├── logger.go                #   日志（Zap）
│   ├── health.go                #   健康检查（/health + /ready）
│   ├── tracing.go               #   分布式追踪（OpenTelemetry）
│   ├── metrics.go               #   Prometheus 指标端点
│   └── wire.go                  #   Composition Root
│
├── cmd/
│   └── dbinit/                  #   数据库初始化 CLI
│
├── config/
│   └── config.example.yaml      # 配置模板
│
├── middleware/                   # 中间件
│   ├── recovery.go              #   panic 恢复
│   ├── cors.go                  #   跨域
│   ├── request_id.go            #   请求 ID
│   ├── logger.go                #   请求日志（含 traceID + userID + context logger）
│   ├── auth.go                  #   JWT 认证
│   ├── rbac.go                  #   RBAC 权限
│   ├── ratelimit.go             #   请求限流（memory + Redis）
│   ├── upload.go                #   文件上传中间件
│   └── metrics.go               #   Prometheus 指标采集
│
├── internal/
│   ├── _template/               #   模块骨架模板
│   └── auth/                    #   RBAC 权限模块（用户→角色→权限）
│
├── example/order/               # 订单完整示例（DDD 六边形架构）
│
├── pkg/                         # 共享工具包
│   ├── auth/                    #   JWT + AuthUser + PermissionChecker
│   ├── cron/                    #   定时任务调度器 + 分布式锁
│   ├── errors/                  #   统一错误结构
│   ├── excel/                   #   Excel 解析 & 生成（泛型 + struct tag）
│   ├── log/                     #   Zap 日志封装 + 上下文注入 + 敏感数据脱敏
│   ├── pagination/              #   分页工具
│   ├── response/                #   统一响应格式
│   └── storage/                 #   文件存储抽象（本地 + 阿里云 OSS）
│
├── proto/                       # Protobuf 定义（Connect RPC）
│
└── docs/
    ├── architecture.md          # 架构原理详解
    └── feature-gap-analysis.md  # 功能差距分析
```

## 可观测性

框架内置三项可观测性能力，通过配置开关控制：

### 健康检查

| 端点 | 用途 | 说明 |
|------|------|------|
| `GET /health` | K8s livenessProbe | 始终返回 200，表示进程存活 |
| `GET /ready` | K8s readinessProbe | 检测 DB + Redis 连通性，降级时返回 503 |

### Prometheus 指标

配置 `metrics.enabled: true` 后暴露 `/metrics` 端点，自动采集：

| 指标 | 类型 | 标签 |
|------|------|------|
| `http_requests_total` | Counter | method, path, status |
| `http_request_duration_seconds` | Histogram | method, path |
| `http_requests_in_flight` | Gauge | — |

### 分布式追踪

配置 `tracing.enabled: true` 后启用 OpenTelemetry，支持 OTLP 导出（Jaeger / Tempo）：

- 自动为每个 HTTP 请求创建 Span
- W3C TraceContext 头传播
- traceID 自动注入请求日志

## 新增业务模块

### 从模板创建（推荐）

```bash
# 1. 复制模板
cp -r internal/_template internal/product

# 2. 替换占位名
find internal/product -type f -name "*.go" \
  -exec sed -i '' 's/EntityName/Product/g' {} +

# 3. 修改 import 路径
find internal/product -type f -name "*.go" \
  -exec sed -i '' 's|internal/_template|internal/product|g' {} +

# 4. 在 app/wire.go 中注册
#    import "github.com/example/dd-frame/internal/product"
#    productAPI := product.Wire()
#    productAPI.RegisterRoutes(v1)
```

### 关键约束

- 模块目录必须在 `internal/` 下（编译器级别隔离）
- 模块间不直接引用（通过端口接口 + ACL 防腐层）
- 所有模块在 `app/wire.go` 统一注册

## Makefile 命令

```bash
make help           # 查看所有可用命令

# 开发
make build          # 编译项目
make run            # 运行（含端口检查）
make test           # 运行测试
make vet            # 静态分析
make tidy           # 整理依赖
make clean          # 清理构建产物

# 数据库
make db-init        # 数据库迁移
make db-seed        # 初始化种子数据

# Swagger
make swagger        # 生成 API 文档
make swagger-deps   # 安装 swag CLI

# Proto
make proto-deps     # 安装 proto 工具
make proto-gen      # 生成 Go 代码

# Docker
make docker-build   # 构建镜像
make docker-up      # 启动开发环境
make docker-down    # 停止开发环境
make docker-logs    # 查看容器日志

# 组合
make check          # vet + test
make all            # tidy + proto-gen + build + test
```

## 技术栈

| 组件 | 用途 | 包 |
|------|------|----|
| Go 1.26 | 编程语言 | — |
| Gin | HTTP 路由框架 | `github.com/gin-gonic/gin` |
| GORM | ORM 框架 | `gorm.io/gorm` + `gorm.io/driver/mysql` |
| Redis | 分布式锁 & 缓存 | `github.com/redis/go-redis/v9` |
| Viper | 配置解析 | `github.com/spf13/viper` |
| Zap | 结构化日志 | `go.uber.org/zap` |
| OpenTelemetry | 分布式追踪 | `go.opentelemetry.io/otel` + `otelgin` |
| Prometheus | 指标采集 | `github.com/prometheus/client_golang` |
| JWT | 认证鉴权 | `github.com/golang-jwt/jwt/v5` |
| Swagger | API 文档 | `swaggo/swag` + `gin-swagger` |
| Connect RPC | gRPC 兼容（预留） | `connectrpc.com/connect-go` |
| Buf | Proto 管理 | `github.com/bufbuild/buf` |
| excelize | Excel 解析 & 生成 | `github.com/xuri/excelize/v2` |
| robfig/cron | 定时任务 | `github.com/robfig/cron/v3` |
| aliyun-oss-go-sdk | 阿里云 OSS | `github.com/aliyun/aliyun-oss-go-sdk/oss` |

## 文档

| 文档 | 说明 |
|------|------|
| [架构原理](docs/architecture.md) | DDD 分层设计、模块隔离、中间件链、启动流程 |
| [功能差距分析](docs/feature-gap-analysis.md) | 框架已实现 vs 待实现功能清单 |
| [Swagger API](http://localhost:8080/swagger/index.html) | 在线 API 文档（启动后访问） |
