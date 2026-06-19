---
name: ddd-feature-dev
description: Guide feature development following DDD hexagonal architecture principles. Use when creating new business features, adding domain modules, designing use cases, or when the user asks to implement a new feature in the DDD project. Covers domain modeling, port/adapter design, layered code generation, ACL anti-corruption, and IoC assembly.
---

# DDD 功能开发设计

基于六边形架构（端口与适配器）的 DDD 功能开发流程指导。

适用于本项目**模块化单体（Modular Monolith）** 架构，每个限界上下文在 `internal/` 下拥有独立 DDD 分层。

## 项目架构速览

```
dd-frame/
├── app/                    # 基础设施 & 全局 Composition Root
│   ├── config.go           #   Viper 配置
│   ├── server.go           #   Gin HTTP 服务器
│   ├── database.go         #   GORM 初始化
│   ├── cache.go            #   Redis 初始化
│   ├── logger.go           #   Zap 日志
│   └── wire.go             #   Composition Root：注册所有模块
├── middleware/             # 横切关注点（recovery/cors/auth/logger/request_id）
├── pkg/                    # 共享工具包（errors/log/response/pagination）
├── internal/{module}/      # 业务模块（Go internal/ 编译器隔离）
│   ├── domain/             #   领域层
│   ├── biz/                #   业务编排层
│   ├── service/            #   应用边界层
│   ├── api/                #   API 层
│   ├── model/              #   仓储 & 缓存
│   └── wire.go             #   模块内 IoC 装配
├── example/order/          # 完整订单示例（参考用）
└── internal/_template/     # 干净模块骨架模板
```

**依赖方向**：`api → service → biz → domain ← model`，领域层零外部依赖。

## 开发流程总览

按以下顺序自底向上开发，确保依赖方向正确：

```
1. 需求分析 → 识别限界上下文与用例
2. 创建模块 → 从 internal/_template/ 复制骨架
3. domain/  → 领域建模（聚合根、值对象、枚举、错误、领域服务接口）
4. model/repo.go → 仓储接口定义
5. biz/ports.go → 端口接口定义（外部依赖）
6. biz/service.go → 应用服务编排（用例实现）
7. model/dao.go → 仓储 GORM 实现 + converter
8. model/cache.go → 缓存接口
9. service/ → 应用边界层（DTO 转换）
10. api/    → HTTP Handler + 路由注册
11. wire.go → 模块内 IoC 装配
12. app/wire.go → 全局注册模块路由
```

## Phase 1: 需求分析与限界上下文

### 1.1 识别限界上下文

从需求中提取核心业务概念，确定所属限界上下文：

```
问自己：
- 这个功能的核心业务概念是什么？（如：订单、支付、库存）
- 它和哪些现有上下文有交互？
- 它的边界在哪里？（哪些属于这个上下文，哪些不属于）
```

### 1.2 梳理用例清单

列出该上下文的所有用例，区分入站和出站：

```markdown
## [上下文名] 用例清单

### 入站用例（驱动型）
- UC-1: 创建 XX — 用户通过 HTTP 触发
- UC-2: 提交 XX — 用户通过 HTTP 触发
- UC-3: 查询 XX 列表 — 用户通过 HTTP 触发

### 出站依赖（被驱动型）
- 持久化：XX 仓储（CRUD）
- 外部系统：支付平台 / 库存系统 / 通知服务
```

### 1.3 创建模块

```bash
# 从模板复制骨架到 internal/ 下
cp -r internal/_template internal/{module}

# 全局替换 EntityName → 实际领域名
find internal/{module} -type f -name "*.go" \
  -exec sed -i '' 's/EntityName/{Entity}/g' {} +

# 修改 import 路径
find internal/{module} -type f -name "*.go" \
  -exec sed -i '' 's|internal/_template|internal/{module}|g' {} +
```

## Phase 2: 领域建模（domain/）

**路径**: `internal/{module}/domain/`

### 2.1 聚合根（Entity）

聚合根是外部访问聚合的唯一入口，必须包含业务行为方法：

```go
// internal/{module}/domain/entity.go
package domain

type {Entity} struct {
    ID         int64
    // ... 业务字段
    Status     {Entity}Status
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// 业务行为方法 — 规则内聚在聚合根中
func (e *{Entity}) {Action}() error {
    // 业务校验
    if e.Status != {ValidStatus} {
        return Err{Invalid}(e.Status, "{action}")
    }
    // 状态流转
    e.Status = {NextStatus}
    return nil
}
```

**规则**：
- 聚合根方法包含所有业务校验和状态变更逻辑
- 禁止在 biz/service 层写业务规则（贫血模型反模式）
- 一个聚合根对应一个仓储接口

### 2.2 值对象（Value Object）

```go
// internal/{module}/domain/value_object.go
type {ValueObject} int64  // 或其他不可变类型

func (v {ValueObject}) Add(other {ValueObject}) {ValueObject} {
    return {ValueObject}(int64(v) + int64(other))
}
```

**规则**：不可变，通过属性值判等，操作方法返回新实例。

### 2.3 枚举

```go
// internal/{module}/domain/enums.go
type {Entity}Status int

const (
    {Entity}StatusDraft {Entity}Status = 0
    {Entity}StatusActive {Entity}Status = 1
)

func (s {Entity}Status) IsValid() bool    { /* 范围校验 */ }
func (s {Entity}Status) IsFinal() bool    { /* 终态判断 */ }
func (s {Entity}Status) String() string   { /* 中文描述 */ }
```

**规则**：每个枚举必须包含 `IsValid()`、`String()`，按需添加业务判断方法。

### 2.4 领域错误

```go
// internal/{module}/domain/errors.go
const Reason{Error} = "{ERROR_CODE}"

func Err{Error}(...) error {
    return fmt.Errorf("[%s] ...", Reason{Error}, ...)
}
```

### 2.5 领域服务接口

当逻辑不适合放在单个聚合根上时（跨聚合操作）：

```go
// internal/{module}/domain/service.go
type {DomainService} interface {
    {Method}(param Type) (Result, error)
}
```

## Phase 3: 端口与仓储接口

### 3.1 仓储接口（model/repo.go）

```go
// internal/{module}/model/repo.go
package model

type {Entity}Repo interface {
    Create(ctx context.Context, entity *domain.{Entity}) error
    QueryByID(ctx context.Context, id int64) (*domain.{Entity}, error)
    UpdateStatus(ctx context.Context, id int64, status domain.{Entity}Status) error
    // 使用领域语义命名，不暴露 DB 细节
}
```

**规则**：
- 接口参数和返回值使用领域对象，不使用 DB 模型
- 方法命名体现领域语义：`QueryByOrderNo` 而非 `FindByField`

### 3.2 出站端口接口（biz/ports.go）

```go
// internal/{module}/biz/ports.go
type External{System}Client interface {
    {Action}(ctx context.Context, params...) (Result, error)
}
```

## Phase 4: 应用服务编排（biz/）

**路径**: `internal/{module}/biz/service.go`

```go
// internal/{module}/biz/service.go
type {Entity}Service interface {
    {UseCase}(ctx context.Context, req *Request) (*domain.{Entity}, error)
}

type {entity}Service struct {
    repo       model.{Entity}Repo
    extClient  External{System}Client
    // ... 其他端口
}

func (s *{entity}Service) {UseCase}(ctx context.Context, ...) (*domain.{Entity}, error) {
    // 1. 参数校验
    // 2. 构建/加载聚合根
    // 3. 调用聚合根方法（业务规则在领域层）
    // 4. 调用外部端口（库存、支付等）
    // 5. 持久化
    // 6. 异步通知
    return entity, nil
}
```

**规则**：
- 只做编排，不写业务规则
- 业务校验交给聚合根方法
- 所有外部依赖通过端口接口注入

## Phase 5: 防腐层（biz/acl/，按需创建）

当需要对接外部 SDK 时，在模块内创建 `biz/acl/` 目录：

```go
// internal/{module}/biz/acl/{system}_adapter.go
package acl

var _ biz.External{System}Client = (*{System}Adapter)(nil)

type {System}Adapter struct { /* 注入外部 SDK */ }

func (a *{System}Adapter) {Action}(ctx context.Context, ...) (Result, error) {
    // 1. 领域参数 → SDK 请求格式
    // 2. 调用外部 SDK
    // 3. SDK 响应 → 领域结果
    // 4. SDK 错误 → 领域错误
    return result, nil
}
```

**规则**：领域层永远不引用外部 SDK 的 DTO，所有转换在 ACL 完成。

## Phase 6: 仓储 & 缓存实现（model/）

### 6.1 DAO 实现

```go
// internal/{module}/model/dao.go
var _ {Entity}Repo = (*{Entity}DAO)(nil)

type {Entity}Model struct { /* DB 表结构 + GORM tag */ }
func ({Entity}Model) TableName() string { return "t_{table}" }

type {Entity}DAO struct { db *gorm.DB }

func (d *{Entity}DAO) Create(ctx context.Context, entity *domain.{Entity}) error {
    m := entityToModel(entity)
    d.db.WithContext(ctx).Create(&m)
    entity.ID = m.ID
    return nil
}

// converter: 领域对象 ↔ DB 模型
func entityToModel(e *domain.{Entity}) *{Entity}Model { /* ... */ }
func modelToEntity(m *{Entity}Model) *domain.{Entity} { /* ... */ }
```

### 6.2 缓存接口

```go
// internal/{module}/model/cache.go
type {Entity}Cache interface {
    SetLock(ctx context.Context, key string, ttl int) (bool, error)
    ReleaseLock(ctx context.Context, key string) error
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, data string, ttl int) error
}
```

## Phase 7: 应用边界层（service/）

```go
// internal/{module}/service/app_service.go
type {Entity}AppService struct {
    usecase biz.{Entity}Service
}

// Input/Output DTO — 对应 HTTP 请求/响应格式
type Create{Entity}Input struct { /* HTTP 请求字段 + json tag */ }
type Create{Entity}Output struct { /* HTTP 响应字段 + json tag */ }

func (s *{Entity}AppService) Create{Entity}(ctx context.Context, input *Input) (*Output, error) {
    // 1. HTTP DTO → 业务 DTO
    // 2. 调用 biz 层
    // 3. 领域对象 → HTTP DTO
    return output, nil
}
```

## Phase 8: API Handler（api/）

```go
// internal/{module}/api/http_handler.go
type {Entity}API struct {
    svc *service.{Entity}AppService
}

func New{Entity}API(svc *service.{Entity}AppService) *{Entity}API {
    return &{Entity}API{svc: svc}
}

// RegisterRoutes 注册模块路由
func (a *{Entity}API) RegisterRoutes(rg *gin.RouterGroup) {
    group := rg.Group("/{entity}")
    group.POST("", a.CreateHandler)
    group.POST("/:id/submit", a.SubmitHandler)
    group.GET("/:id", a.GetHandler)
}

func (a *{Entity}API) CreateHandler(c *gin.Context) {
    var input service.Create{Entity}Input
    if err := c.ShouldBindJSON(&input); err != nil {
        response.Error(c, http.StatusBadRequest, 40000, "invalid request")
        return
    }
    output, err := a.svc.Create{Entity}(c.Request.Context(), &input)
    if err != nil {
        response.FromError(c, err)
        return
    }
    response.Success(c, output)
}
```

## Phase 9: 模块内 IoC 装配（wire.go）

```go
// internal/{module}/wire.go
package {module}

func Wire() *api.{Entity}API {
    // 1. 数据层
    repo := model.New{Entity}DAO()

    // 2. 端口适配器
    extClient := acl.New{System}Adapter()

    // 3. 业务编排层
    svc := biz.New{Entity}Service(repo, extClient)

    // 4. 应用边界层
    appSvc := service.New{Entity}AppService(svc)

    // 5. API 层
    return api.New{Entity}API(appSvc)
}
```

## Phase 10: 全局注册（app/wire.go）

在 Composition Root 中注册新模块：

```go
// app/wire.go — 新增一行
import "github.com/example/dd-frame/internal/{module}"

func Wire(cfg *Config) *gin.Engine {
    r := gin.New()
    v1 := r.Group("/api/v1")

    // 已有模块
    orderAPI := order.Wire()
    orderAPI.RegisterRoutes(v1)

    // 新增模块 ← 在这里追加
    {module}API := {module}.Wire()
    {module}API.RegisterRoutes(v1)

    return r
}
```

## 跨模块协作

模块间**不直接引用**，通过端口接口 + ACL 防腐层解耦：

```
模块 A 需要调用模块 B 的能力：
1. 在模块 A 的 biz/ports.go 中定义端口接口
2. 在 app/wire.go 中创建 ACL 适配器，实现模块 A 的端口接口
3. ACL 适配器内部调用模块 B 的 service 层（app/wire.go 是唯一知道所有模块的地方）
4. 将适配器注入模块 A 的 biz 层
```

## 设计检查清单

完成开发后，逐项检查：

- [ ] 聚合根方法包含业务规则，biz 层无业务判断
- [ ] 所有枚举有 `IsValid()` 和 `String()` 方法
- [ ] 仓储接口使用领域对象，不暴露 DB 模型
- [ ] 端口接口在 `biz/ports.go` 定义，实现在 ACL 或 stub
- [ ] 防腐层隔离了外部 SDK DTO，领域层不引用外部包
- [ ] service 层只做 DTO 转换，不含业务逻辑
- [ ] api 层只做请求解析，不含业务逻辑
- [ ] 模块 `wire.go` 装配完整依赖链
- [ ] `app/wire.go` 已注册新模块路由
- [ ] 编译期校验 `var _ Interface = (*Impl)(nil)` 已添加
- [ ] 新增用例有对应的错误码定义
- [ ] `go build ./...` 和 `go vet ./...` 零错误
