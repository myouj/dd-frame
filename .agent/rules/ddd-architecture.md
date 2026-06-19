---
description: DDD 六边形架构分层约束。编写或修改 internal/ 和 example/ 下的 Go 代码时自动应用。
globs: ["internal/**/*.go", "example/**/*.go"]
---

# DDD 架构规则

## 依赖方向

```
api → service → biz → domain ← model
```

领域层是核心，所有外层依赖内层，内层不感知外层。

## 分层约束

### domain/（领域层）

- **只允许**：标准库导入、本包内定义的类型
- **禁止**：导入任何第三方包（gin/gorm/redis 等）、导入其他模块、导入 model/biz/service/api 层
- 聚合根必须包含业务行为方法（状态流转、校验、计算），禁止贫血模型
- 值对象不可变，操作方法返回新实例
- 枚举必须有 `IsValid()` 和 `String()` 方法
- 领域错误必须有 `Reason` 常量前缀（如 `ORDER_NOT_FOUND`）

### biz/（业务编排层）

- **只允许**：调用 domain 聚合根方法、调用端口接口、调用仓储接口
- **禁止**：直接写业务规则（if status == / status = 赋值）、导入 service/api 层、导入其他模块
- 端口接口（外部依赖）必须定义在 `biz/ports.go`
- 所有外部依赖通过接口注入，不直接实例化具体实现

### service/（应用边界层）

- **只允许**：HTTP/gRPC DTO ↔ 领域对象的转换、调用 biz 层
- **禁止**：包含任何业务判断、金额计算、状态校验

### api/（API 层）

- **只允许**：请求解析（ShouldBindJSON）、context 注入、调用 service 层、返回响应
- **禁止**：包含业务逻辑、直接操作 DB、引用 biz/domain 层

### model/（数据层）

- 仓储接口参数和返回值必须使用领域对象，禁止暴露 DB 模型
- DAO 实现必须有编译期校验：`var _ Repo = (*DAO)(nil)`
- converter 函数在 model 包内实现领域对象 ↔ DB 模型转换

## 模块隔离

- 模块位于 `internal/{module}/`，Go 编译器阻止外部导入
- 模块间**禁止直接引用**（order 不能 import product）
- 跨模块协作：需求方在 `biz/ports.go` 定义端口接口 → `app/wire.go` 中创建 ACL 适配器 → 注入

## IoC 装配

- 每个模块有 `wire.go` 负责本模块依赖装配，返回 API handler
- `app/wire.go` 是唯一的 Composition Root，知道所有模块并注册路由
- 新增模块必须在 `app/wire.go` 中注册
