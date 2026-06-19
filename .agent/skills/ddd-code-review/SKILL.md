---
name: ddd-code-review
description: Review code changes against DDD hexagonal architecture principles. Use after completing feature development, before merging code, or when the user asks to review/audit code for DDD compliance. Strictly checks layer dependency violations, anemic domain model anti-patterns, missing port interfaces, ACL corruption, and module isolation breaches.
---

# DDD 代码审查

对本项目（模块化单体 + 六边形架构）的代码变更进行严格的 DDD 设计合规性审查。

## 审查原则

1. **只报真实违规**，不报风格偏好
2. **每条问题引用具体规则编号**，便于定位
3. **提供修复建议**，而不只是指出问题
4. **按严重等级排序**：P0 架构违规 > P1 分层违规 > P2 规范偏离

## 审查范围

审查以下目录中的代码变更：

```
internal/{module}/   ← 业务模块（重点审查）
example/{module}/    ← 示例模块
app/                 ← 全局装配
```

以下目录不在审查范围：

```
pkg/                 ← 共享工具包，不适用 DDD 规则
middleware/          ← 横切关注点，不适用 DDD 规则
```

## 审查规则

---

### 规则 D-1：领域层零外部依赖（P0）

**domain/ 下的代码禁止导入任何非标准库包。**

检查方式：扫描 `internal/{module}/domain/` 和 `example/{module}/domain/` 下所有 Go 文件的 import 语句。

**违规示例：**

```go
// domain/entity.go — 违规！导入了 GORM
import "gorm.io/gorm"

type Order struct {
    gorm.Model  // 违规！领域对象不应依赖 ORM
    OrderNo string
}
```

```go
// domain/service.go — 违规！导入了 model 层
import "github.com/example/dd-frame/internal/order/model"
```

**正确做法：** 领域层只使用标准库 + 本包内定义的类型。

---

### 规则 D-2：聚合根必须包含业务行为（P0）

聚合根（entity.go 中的主结构体）**必须有业务方法**，不能是纯数据载体（贫血模型）。

**违规示例：**

```go
// domain/entity.go — 贫血模型反模式
type Order struct {
    ID         int64
    OrderNo    string
    Status     OrderStatus
    Amount     Money
}
// 没有任何方法！所有逻辑都在 biz 层 → 贫血模型
```

**正确做法：**

```go
type Order struct { ... }

func (o *Order) Submit() error {
    if o.Status != OrderStatusDraft { return ErrOrderStatusInvalid(...) }
    o.Status = OrderStatusSubmitted
    return nil
}
```

**判断标准：**
- 聚合根至少有 1 个业务行为方法（状态流转、计算、校验）
- 如果所有 if/switch 业务判断都在 biz 层 → 贫血模型违规

---

### 规则 D-3：biz 层禁止包含业务规则（P0）

biz 层只做**用例编排**（调用聚合根方法 + 调用端口接口），禁止写业务校验逻辑。

**违规示例：**

```go
// biz/service.go — 违规！业务规则泄漏到 biz 层
func (s *orderService) SubmitOrder(ctx context.Context, orderNo string) error {
    order, _ := s.repo.QueryByOrderNo(ctx, orderNo)

    // 违规！状态校验应该在聚合根方法中
    if order.Status != OrderStatusDraft {
        return fmt.Errorf("cannot submit order in status %s", order.Status)
    }
    if len(order.Items) == 0 {
        return fmt.Errorf("order is empty")
    }
    order.Status = OrderStatusSubmitted  // 违规！直接修改状态，应调用 order.Submit()

    return s.repo.UpdateStatus(ctx, order.ID, order.Status)
}
```

**正确做法：**

```go
func (s *orderService) SubmitOrder(ctx context.Context, orderNo string) error {
    order, _ := s.repo.QueryByOrderNo(ctx, orderNo)
    if err := order.Submit(); err != nil {  // 业务规则在聚合根中
        return err
    }
    return s.repo.UpdateStatus(ctx, order.ID, order.Status)
}
```

**判断标准：** biz 方法中出现 `if order.Status ==` 或 `order.Status =` 直接赋值 → 违规。

---

### 规则 D-4：依赖方向违规（P0）

合法的依赖方向：`api → service → biz → domain ← model`

**严格禁止的导入关系：**

| 禁止方向 | 说明 |
|----------|------|
| domain → biz/service/api/model | 领域层不能依赖任何外层 |
| domain → 其他模块 | 模块间领域层禁止交叉引用 |
| biz → service/api | biz 不能依赖上层 |
| service → api | service 不能依赖 API 层 |

**检查方式：** 扫描各层 import 路径，检查是否违反依赖方向。

**违规示例：**

```go
// domain/service.go — 违规！domain 导入了 biz
import "github.com/example/dd-frame/internal/order/biz"
```

```go
// biz/service.go — 违规！biz 导入了 service
import "github.com/example/dd-frame/internal/order/service"
```

---

### 规则 D-5：模块隔离违规（P0）

`internal/{moduleA}/` 禁止直接导入 `internal/{moduleB}/`。

**违规示例：**

```go
// internal/order/biz/service.go — 违规！直接引用了 product 模块
import "github.com/example/dd-frame/internal/product/domain"
```

**正确做法：** 通过端口接口 + ACL 解耦，在 `app/wire.go` 中装配。

```go
// internal/order/biz/ports.go — 定义端口接口
type InventoryClient interface {
    QueryStock(ctx context.Context, productID int64) (int, error)
}

// app/wire.go — ACL 适配器在 Composition Root 中实现
type productInventoryAdapter struct { productSvc *product.Service }
func (a *productInventoryAdapter) QueryStock(...) { a.productSvc.QueryStock(...) }
```

---

### 规则 M-1：仓储接口使用领域对象（P1）

model/repo.go 中的仓储接口参数和返回值**必须使用领域对象**，禁止使用 DB 模型或原始类型。

**违规示例：**

```go
// model/repo.go — 违规！暴露了 DB 模型
type OrderRepo interface {
    Create(ctx context.Context, model *OrderModel) error  // 应用 *domain.Order
    FindByOrderNo(ctx context.Context, no string) (*OrderModel, error)  // 应返回 *domain.Order
}
```

**违规示例：**

```go
// model/repo.go — 违规！方法名暴露 DB 语义
type OrderRepo interface {
    FindAll(ctx context.Context) ([]*domain.Order, error)  // 应使用领域语义：QueryByCustomerID
    SelectBySQL(ctx context.Context, query string) (...)    // 禁止暴露 SQL
}
```

---

### 规则 M-2：编译期接口校验缺失（P1）

每个接口的实现必须有编译期校验：

```go
var _ InterfaceName = (*ImplName)(nil)
```

**检查范围：**
- `model/dao.go`：`var _ {Entity}Repo = (*{Entity}DAO)(nil)`
- `biz/acl/`：`var _ biz.External{System}Client = (*{System}Adapter)(nil)`

---

### 规则 M-3：端口接口位置错误（P1）

出站端口接口**必须定义在 biz/ports.go**，不能放在其他位置。

**违规示例：**

```go
// domain/service.go — 违规！外部端口不应在 domain 层
type PaymentGateway interface {
    Pay(ctx context.Context, amount Money) error
}
```

```go
// model/repo.go — 违规！外部端口不应在 model 层
type NotificationSender interface {
    Send(ctx context.Context, msg string) error
}
```

**正确做法：** 所有外部依赖端口定义在 `biz/ports.go`。

---

### 规则 M-4：service 层包含业务逻辑（P1）

service 层（应用边界层）**只做 DTO 转换**，禁止包含业务判断。

**违规示例：**

```go
// service/app_service.go — 违规！包含业务逻辑
func (s *OrderAppService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error) {
    // 违规！金额校验是业务规则，应在 domain 或 biz 层
    if input.TotalAmount <= 0 {
        return nil, fmt.Errorf("amount must be positive")
    }
    // 违规！折扣计算是业务逻辑
    if input.CustomerVIP {
        input.TotalAmount *= 0.9
    }
    ...
}
```

**正确做法：** service 层只做 `Input DTO → biz DTO` 和 `domain Object → Output DTO` 转换。

---

### 规则 M-5：api 层包含业务逻辑（P1）

api 层**只做请求解析和响应序列化**，禁止包含业务判断。

**违规示例：**

```go
// api/http_handler.go — 违规！包含业务逻辑
func (a *OrderAPI) CreateHandler(c *gin.Context) {
    var input service.CreateOrderInput
    c.ShouldBindJSON(&input)

    // 违规！库存检查是业务逻辑
    for _, item := range input.Items {
        if item.Quantity > 100 {
            response.Error(c, 400, 40001, "quantity too large")
            return
        }
    }
    ...
}
```

---

### 规则 S-1：枚举缺少必要方法（P2）

每个枚举类型必须有 `IsValid() bool` 和 `String() string` 方法。

**违规示例：**

```go
// domain/enums.go — 违规！缺少 IsValid 和 String
type OrderStatus int

const (
    OrderStatusDraft OrderStatus = 0
    OrderStatusSubmitted OrderStatus = 1
)
// 没有 IsValid() 和 String() 方法
```

---

### 规则 S-2：领域错误缺少 Reason 常量（P2）

每个领域错误必须有对应的 `Reason` 常量，用于错误码传递。

**违规示例：**

```go
// domain/errors.go — 违规！直接用 fmt.Errorf，没有 Reason 常量
func ErrOrderNotFound(no string) error {
    return fmt.Errorf("order %s not found", no)  // 缺少 Reason 前缀
}
```

**正确做法：**

```go
const ReasonOrderNotFound = "ORDER_NOT_FOUND"

func ErrOrderNotFound(no string) error {
    return fmt.Errorf("[%s] order %s not found", ReasonOrderNotFound, no)
}
```

---

### 规则 S-3：wire.go 装配遗漏（P2）

模块 `wire.go` 必须创建并返回完整的依赖链，不能有遗漏的端口实现。

**检查方式：**
- wire 函数返回的 API handler 所依赖的所有 service/biz/model 都已创建
- 所有 biz 层需要的端口接口都有对应的实现注入

---

### 规则 S-4：app/wire.go 未注册模块（P2）

新增模块后必须在 `app/wire.go` 中注册路由。

**检查方式：** 如果 `internal/{module}/wire.go` 存在，但 `app/wire.go` 中没有调用 `{module}.Wire()` → 违规。

---

## 审查输出格式

审查结果按以下格式输出：

```markdown
## DDD 代码审查报告

### 概要

| 等级 | 数量 |
|------|------|
| P0 架构违规 | N |
| P1 分层违规 | N |
| P2 规范偏离 | N |

### 问题列表

#### [P0] 规则 D-3：biz 层包含业务规则
- **文件**: `internal/order/biz/service.go:45-52`
- **问题**: `SubmitOrder` 方法中直接判断 `order.Status != OrderStatusDraft`，应调用 `order.Submit()`
- **修复**: 将状态校验移入聚合根方法，biz 层改为 `if err := order.Submit(); err != nil { return err }`

#### [P1] 规则 M-1：仓储接口暴露 DB 模型
- **文件**: `internal/product/model/repo.go:8`
- **问题**: `Create` 方法参数类型为 `*ProductModel`，应为 `*domain.Product`
- **修复**: 修改接口参数类型为领域对象

### 通过项
- [x] 规则 D-1：领域层零外部依赖
- [x] 规则 D-4：依赖方向正确
- [x] 规则 D-5：模块隔离完整
- ...
```

## 快速审查命令

```bash
# 1. 检查领域层是否有非标准库导入
grep -rn "import" internal/*/domain/ example/*/domain/ | grep -v "^\s*//"

# 2. 检查是否有跨模块导入
grep -rn "internal/" internal/*/ --include="*.go" | grep -v "_template" | grep -v "wire.go"

# 3. 检查 biz 层是否有状态直接赋值
grep -rn "\.Status\s*=" internal/*/biz/ example/*/biz/ --include="*.go"

# 4. 检查编译期校验
grep -rn "var _ " internal/*/model/ example/*/model/ --include="*.go"

# 5. 编译验证
go build ./... && go vet ./...
```

## 审查流程

1. **获取变更文件列表**（git diff 或用户指定范围）
2. **按 P0 → P1 → P2 顺序逐项检查**
3. **对每个问题提供：文件路径 + 行号 + 问题描述 + 修复建议**
4. **最后列出所有通过的规则**
5. **如果存在 P0 问题，明确标注"必须修复后才能合并"**
