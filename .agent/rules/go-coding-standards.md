---
description: Go 编码规范约束。编写或修改任何 .go 文件时自动应用。
globs: ["**/*.go"]
---

# Go 编码规范

## 错误处理

- 禁止用 `_` 忽略 error 返回值（除非有明确注释说明原因）
- 错误包装必须用 `%w` 并附带上下文：`fmt.Errorf("xxx failed: %w", err)`
- 错误比较用 `errors.Is()` / `errors.As()`，禁止 `==`
- 业务逻辑禁止 `panic`，仅允许在程序启动配置校验和 `init()` 中使用
- 导出错误变量命名：`Err{Xxx}`，Reason 常量命名：`Reason{Xxx}`

## 命名规范

- 包名全小写，无下划线（`userservice` 而非 `user_service`）
- 接口命名不以 `I` 开头，用行为描述（`UserFinder` 而非 `IUserService`）
- Receiver 用类型首字母小写，同类型保持一致（`(u *User)` 不用 `(self *User)`）
- 导出的函数、类型、常量必须有 godoc 注释
- 常量用 `CamelCase`，不用 `ALL_CAPS`（Go 风格）

## 并发安全

- 函数第一个参数为 `context.Context`（涉及 I/O 操作时）
- 所有 goroutine 必须有明确退出机制（context.Done / channel close）
- 共享状态必须有同步保护（sync.Mutex / atomic / channel）
- `sync.WaitGroup.Add()` 必须在 goroutine 启动前调用

## 性能

- 循环中禁止 `defer`（提取为独立函数）
- 循环拼接字符串用 `strings.Builder`
- 已知长度的 slice 用 `make([]T, 0, cap)` 预分配
- 禁止在 `range` 遍历 map 时修改该 map

## 安全

- 禁止字符串拼接构造 SQL，必须用参数化查询
- 禁止硬编码密码、密钥、Token，从环境变量读取
- 外部输入必须校验边界条件

## API 设计

- 函数返回具体类型，不返回接口
- 函数参数接收接口，不接收具体类型
- 配置项较多时用 struct 传参，不过度使用 options pattern

## 测试

- 多场景用表驱动测试（`t.Run` 子测试）
- 测试必须覆盖正常路径和错误路径
- 子测试中用 `t.Error` + `return`，不用 `t.Fatal`

## 代码风格

- 魔法数字必须定义为常量
- 空 switch case 必须注释说明原因
- 函数体不超过 80 行，超过则拆分
