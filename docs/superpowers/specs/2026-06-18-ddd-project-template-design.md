# DDD 通用项目框架模板 — 设计文档

> 日期: 2026-06-18
> 状态: 已确认

## 1. 概述

基于 DDD 六边形架构（端口与适配器）构建的 Go 语言生产可用项目模板。采用模块化单体架构，每个限界上下文拥有独立的 DDD 分层，同时支持 Gin (HTTP) 和 Connect RPC (gRPC 兼容) 双协议。

### 1.1 设计目标

- **生产可用**：集成 Gin、GORM、Redis、Viper、Zap 等真实基础设施
- **模块化**：每个限界上下文自包含完整 DDD 分层，模块间通过端口接口解耦
- **示例分离**：架构骨架与订单示例代码分离，example/ 目录可删除
- **双协议**：同时支持 RESTful HTTP 和 Connect RPC

### 1.2 技术栈

| 组件 | 库 | 用途 |
|------|-----|------|
| HTTP | `gin-gonic/gin` | RESTful API 路由与中间件 |
| gRPC | `connectrpc/connect-go` | Connect RPC 服务（兼容 gRPC 和 HTTP/1.1） |
| ORM | `gorm.io/gorm` + MySQL driver | 数据库操作 |
| 缓存 | `redis/go-redis/v9` | 分布式锁 & 缓存 |
| 配置 | `spf13/viper` | YAML/ENV 配置解析 |
| 日志 | `uber-go/zap` | 结构化日志 |
| 校验 | `go-playground/validator/v10` | 请求参数校验（Gin 内置） |
| Proto | `bufbuild/buf` + `connectrpc` | RPC 接口定义 & 代码生成 |

## 2. 目录结构

```
dd-frame/
├── main.go                           # 应用入口
├── go.mod
├── go.sum
│
├── app/                              # 应用启动 & 装配层
│   ├── server.go                     # Gin HTTP + Connect gRPC 服务器启动
│   ├── config.go                     # Viper 配置加载
│   ├── logger.go                     # Zap 日志初始化
│   ├── database.go                   # GORM 数据库初始化
│   ├── cache.go                      # Redis 初始化
│   └── wire.go                       # IoC 总装配（Composition Root）
│
├── internal/                         # 业务模块（Go internal 约定，不可外部导入）
│   ├── _template/                    # 模块模板（新建模块时复制此目录）
│   │   ├── domain/
│   │   │   ├── entity.go             #     聚合根 + 实体
│   │   │   ├── value_object.go       #     值对象
│   │   │   ├── enums.go              #     枚举
│   │   │   ├── errors.go             #     领域错误
│   │   │   └── service.go            #     领域服务接口
│   │   ├── biz/
│   │   │   ├── service.go            #     应用服务（用例编排）
│   │   │   └── ports.go              #     出站端口接口定义
│   │   ├── service/
│   │   │   └── app_service.go        #     应用边界层（DTO 转换）
│   │   ├── api/
│   │   │   ├── http_handler.go       #     Gin handler
│   │   │   └── grpc_handler.go       #     Connect RPC handler
│   │   ├── model/
│   │   │   ├── repo.go              #     仓储接口
│   │   │   ├── dao.go               #     仓储实现（GORM）
│   │   │   └── cache.go             #     缓存实现
│   │   └── wire.go                   #   模块内 IoC 装配
│   │
│   └── (业务模块如 order、payment 等从此模板复制)
│
├── pkg/                              # 共享工具包（可外部导入）
│   ├── errors/                       #   统一错误处理
│   │   └── errors.go
│   ├── response/                     #   统一 HTTP 响应格式
│   │   └── response.go
│   ├── pagination/                   #   分页工具
│   │   └── pagination.go
│   └── log/                          #   结构化日志封装
│       └── log.go
│
├── proto/                            # Connect RPC 接口定义
│   ├── buf.yaml
│   ├── buf.gen.yaml
│   ├── order/v1/
│   │   └── order.proto
│   └── payment/v1/
│       └── payment.proto
│
├── config/                           # 配置文件
│   ├── config.yaml
│   └── config.example.yaml
│
├── middleware/                        # 共享中间件
│   ├── auth.go                       #   JWT 鉴权
│   ├── cors.go                       #   CORS 跨域
│   ├── recovery.go                   #   Panic 恢复
│   ├── request_id.go                 #   请求 ID（链路追踪）
│   └── logger.go                     #   请求日志
│
├── migrations/                       # 数据库迁移文件
│
├── example/                          # 订单完整示例（参考用，可删除）
│   └── order/
│       ├── domain/
│       ├── biz/
│       ├── service/
│       ├── api/
│       ├── model/
│       └── wire.go
│
└── .gitignore
```

## 3. 架构设计

### 3.1 模块化单体（Modular Monolith）

每个限界上下文（bounded context）是一个独立的 Go package，位于 `internal/` 下，拥有完整的 DDD 分层：

```
internal/{context}/
├── domain/     ← 最内层：零外部依赖（聚合根、值对象、枚举、错误）
├── model/      ← 数据层：仓储接口 + 仓储实现 + 缓存
├── biz/        ← 编排层：应用服务 + 出站端口接口
├── service/    ← 边界层：DTO 转换（HTTP/gRPC DTO ↔ 领域对象）
├── api/        ← 入站层：HTTP handler + gRPC handler
└── wire.go     ← 模块 IoC 装配
```

**依赖方向：api → service → biz → domain ← model**

### 3.2 服务器双协议架构

```
┌──────────────────────────────────────┐
│        Gin HTTP Server :8080         │
│  Middleware → REST Routes            │
│  /api/v1/order  → order handlers    │
│  /api/v1/payment → payment handlers │
└──────────────────────────────────────┘

┌──────────────────────────────────────┐
│     Connect gRPC Server :8081        │
│  Connect RPC Services                │
│  /order.v1.OrderService/            │
│  /payment.v1.PaymentService/        │
└──────────────────────────────────────┘
```

- Gin 和 Connect gRPC 分别在独立端口启动
- 共享同一套 service/biz/domain 层代码
- Connect RPC handler 实现 proto 生成的接口，HTTP handler 直接调用 service 层

### 3.3 跨模块协作

模块间**不直接导入**，通过端口接口 + ACL 防腐层解耦：

1. 需求方模块在 `biz/ports.go` 定义出站端口接口
2. 提供方模块通过 ACL 适配器实现该接口
3. 在 `app/wire.go` 中装配注入

```go
// app/wire.go
func Wire() {
    // payment 模块先初始化（被依赖方）
    paymentSvc := payment.Wire(db, redis)

    // order 模块注入 payment 适配器
    paymentAdapter := order_acl.NewPaymentAdapter(paymentSvc)
    orderSvc := order.Wire(db, redis, paymentAdapter)
}
```

**依赖规则：**

| 规则 | 说明 |
|------|------|
| 模块间不直接导入 | order 不能 `import "internal/payment/..."` |
| 通过端口接口解耦 | 需求方定义接口，提供方实现 |
| ACL 翻译隔离 | 跨模块调用必须经过防腐层 |
| app/wire.go 装配 | 唯一"知道所有模块"的地方 |

## 4. 基础设施集成

### 4.1 应用启动流程（main.go → app/）

```
main.go
  │
  ├── app.LoadConfig()         ← Viper 加载 config/config.yaml
  ├── app.InitLogger()         ← Zap 结构化日志
  ├── app.InitDatabase()       ← GORM + MySQL
  ├── app.InitRedis()          ← go-redis
  ├── app.Wire()               ← IoC 装配所有模块
  └── app.RunServer()          ← 启动 Gin :8080 + Connect gRPC :8081
```

### 4.2 配置结构

```yaml
server:
  http_port: 8080
  grpc_port: 8081
  mode: debug

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
  secret: ""
  expires_in: 24

log:
  level: debug
  format: json
```

### 4.3 统一响应格式

```go
// pkg/response/response.go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{})
func Error(c *gin.Context, code int, msg string)
func FromError(c *gin.Context, err error)
```

### 4.4 中间件

| 中间件 | 功能 | 应用范围 |
|--------|------|---------|
| `recovery` | Panic 恢复，返回 500 | 全局 |
| `cors` | CORS 跨域 | 全局 |
| `request_id` | 生成 X-Request-ID | 全局 |
| `logger` | 请求日志（method/path/status/耗时） | 全局 |
| `auth` | JWT Token 鉴权 | 按路由组 |

## 5. 模块模板规范

### 5.1 新建模块流程

1. 复制 `internal/_template/` 为 `internal/{module_name}/`
2. 按 DDD 顺序填充代码：domain → model → biz → service → api → wire
3. 在 `app/wire.go` 中注册该模块
4. 在 `proto/` 下添加 Connect RPC 定义（如需要 gRPC）
5. 在 `middleware/` 中按需配置路由组中间件

### 5.2 分层编码规范

| 层 | 允许 | 禁止 |
|----|------|------|
| `domain/` | 纯业务逻辑、标准库 | 导入任何外部包 |
| `biz/` | 调用 domain 方法、调用端口接口 | 写业务规则、直接导入其他模块 |
| `service/` | DTO 转换、调用 biz | 写业务逻辑 |
| `api/` | 解析请求、注入 context | 写业务逻辑、直接操作 DB |
| `model/` | GORM 操作、DB 模型转换 | 暴露 DB 模型到外部 |

## 6. 示例分离策略

```
dd-frame/
├── internal/_template/    # 干净骨架（必须保留）
└── example/order/         # 完整示例（可删除）
```

- `example/` 包含订单领域的完整实现（创建/提交/取消），附带详细注释
- 开发者可参考 example 理解 DDD 实践，然后基于 `_template/` 创建自己的模块
- 不需要示例时，直接删除 `example/` 目录

## 7. 实施范围

### 包含（P0 = 必须，P1 = 重要）

| 模块 | 优先级 |
|------|--------|
| 项目基础（go.mod、main.go、目录结构、.gitignore） | P0 |
| 基础设施（Gin、Connect gRPC、GORM、Redis、Viper、Zap） | P0 |
| 共享工具（pkg/response、pkg/errors、pkg/pagination、pkg/log） | P0 |
| 配置管理（config.yaml + config.example.yaml） | P0 |
| 模块模板（internal/_template/ 完整 DDD 分层骨架） | P0 |
| 中间件（auth、cors、recovery、request_id、logger） | P1 |
| 订单示例（example/order/ 完整实现） | P1 |
| Proto 定义（buf.yaml + Connect RPC proto + 代码生成） | P1 |

### 不包含

- CI/CD pipeline
- Docker / K8s 部署配置
- 单元测试框架
- 数据库迁移工具集成

## 8. 设计决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| 架构模式 | 模块化单体 | 模块边界清晰，易于演进到微服务 |
| HTTP 框架 | Gin + Connect RPC | Gin 生态成熟，Connect 兼容 gRPC 和 HTTP/1.1 |
| 模块隔离 | Go `internal/` | 编译器级别的导入保护 |
| 跨模块协作 | 端口接口 + ACL | 避免模块间直接耦合 |
| 示例代码 | 独立 example/ 目录 | 骨架干净，示例可选参考 |
