# dd-frame

基于 **DDD + 六边形架构**的 Go 语言**模块化单体**项目框架。

每个业务模块拥有独立的 DDD 分层，通过 Go `internal/` 实现编译器级别的模块隔离。

> 架构原理与设计详解 → [docs/architecture.md](docs/architecture.md)

## 快速开始

### 1. 克隆 & 配置

```bash
git clone https://github.com/myouj/dd-frame.git
cd dd-frame

# 复制配置模板，按需修改
cp config/config.example.yaml config/config.yaml
vi config/config.yaml
```

### 2. 编译 & 运行

```bash
make build    # 编译到 bin/dd-frame
make run      # 运行（HTTP :8080）
```

或直接使用 Go 命令：

```bash
go mod download
go build ./...
go run main.go
```

### 3. 测试接口（订单示例）

```bash
# 创建订单
curl -X POST http://localhost:8080/api/v1/order \
  -H "Content-Type: application/json" \
  -d '{
    "customerId": 1001,
    "items": [
      {"productId": 1, "quantity": 2, "unitPrice": 9900}
    ]
  }'

# 提交订单
curl -X POST http://localhost:8080/api/v1/order/ORD-xxx/submit

# 取消订单
curl -X POST http://localhost:8080/api/v1/order/ORD-xxx/cancel
```

## 目录结构

```
dd-frame/
├── main.go                      # 应用入口
├── Makefile                     # 构建管理
│
├── app/                         # 基础设施 & 全局装配
│   ├── config.go                #   配置加载（Viper）
│   ├── server.go                #   HTTP 服务器（Gin + 优雅关闭）
│   ├── database.go              #   数据库（GORM + MySQL，未配置时跳过）
│   ├── cache.go                 #   缓存（Redis，未配置时跳过）
│   ├── logger.go                #   日志（Zap）
│   └── wire.go                  #   Composition Root
│
├── config/
│   └── config.example.yaml      # 配置模板
│
├── middleware/                   # 中间件（recovery/cors/request_id/logger/auth）
│
├── internal/_template/          # 模块骨架模板（复制后使用）
│   ├── domain/                  #   领域层
│   ├── biz/                     #   业务编排层
│   ├── service/                 #   应用边界层
│   ├── api/                     #   API 层
│   ├── model/                   #   仓储 & 缓存
│   └── wire.go                  #   模块内 IoC
│
├── example/order/               # 订单完整示例（DDD 六边形架构）
│
├── pkg/                         # 共享工具包（errors/log/response/pagination）
│
├── proto/                       # Protobuf 定义（Connect RPC）
│   ├── buf.yaml
│   ├── buf.gen.yaml
│   └── order/v1/order.proto
│
└── docs/
    └── architecture.md          # 架构原理详解
```

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

### 参考订单示例

`example/order/` 包含完整的创建/提交/取消订单实现，展示了 DDD 各层如何协作。

### 关键约束

- 模块目录必须在 `internal/` 下（编译器级别隔离）
- 模块间不直接引用（通过端口接口 + ACL 防腐层）
- 所有模块在 `app/wire.go` 统一注册

## Makefile 命令

```bash
make help           # 查看所有可用命令
make build          # 编译项目
make run            # 运行项目
make test           # 运行测试
make vet            # 静态分析
make tidy           # 整理依赖
make clean          # 清理构建产物

# Proto 相关
make proto-deps     # 安装 proto 工具（protoc-gen-go, connect-go, buf）
make proto-gen      # 生成 Go 代码
make proto-buf      # 使用 buf 生成代码
make proto-lint     # 检查 proto 规范

# 组合命令
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
| Connect RPC | gRPC 兼容（预留） | `connectrpc.com/connect-go` |
| Buf | Proto 管理 | `github.com/bufbuild/buf` |

## 文档

| 文档 | 说明 |
|------|------|
| [架构原理](docs/architecture.md) | DDD 分层设计、模块隔离、用例流程、时序图 |
| [设计文档](docs/superpowers/specs/2026-06-18-ddd-project-template-design.md) | 项目框架设计规格 |
| [实施计划](docs/superpowers/plans/2026-06-18-ddd-project-template.md) | 10 个 Task 的实施记录 |
