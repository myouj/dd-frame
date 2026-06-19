---
description: 项目模块创建与文件组织规范。创建新模块或修改项目结构时自动应用。
globs: ["internal/_template/**/*.go", "internal/**/wire.go", "app/wire.go"]
---

# 模块与项目结构规范

## 项目目录职责

| 目录 | 职责 | 可删除 |
|------|------|--------|
| `app/` | 基础设施初始化 + 全局 Composition Root | 否 |
| `config/` | YAML 配置文件（config.yaml 被 .gitignore） | 否 |
| `middleware/` | Gin 中间件（横切关注点） | 否 |
| `pkg/` | 共享工具包（errors/log/response/pagination） | 否 |
| `internal/_template/` | 干净 DDD 模块骨架模板 | 否 |
| `internal/{module}/` | 业务模块实现 | 按需 |
| `example/` | 订单完整示例（参考用） | 可删除 |
| `proto/` | Connect RPC proto 定义（预留） | 按需 |

## 新建模块流程

1. `cp -r internal/_template internal/{module}`
2. 全局替换 `EntityName` → 实际领域名
3. 修改 import 路径中的 `_template` → `{module}`
4. 按 DDD 顺序填充：domain → model → biz → service → api → wire
5. 在 `app/wire.go` 注册模块路由

## 模块文件结构

```
internal/{module}/
├── domain/
│   ├── entity.go          # 聚合根 + 实体
│   ├── value_object.go    # 值对象
│   ├── enums.go           # 枚举
│   ├── errors.go          # 领域错误
│   └── service.go         # 领域服务接口
├── biz/
│   ├── service.go         # 应用服务（用例编排）
│   ├── ports.go           # 出站端口接口
│   └── acl/               # 防腐层（按需创建）
├── service/
│   └── app_service.go     # 应用边界层（DTO 转换）
├── api/
│   ├── http_handler.go    # Gin HTTP handler + RegisterRoutes
│   └── grpc_handler.go    # Connect RPC handler（按需）
├── model/
│   ├── repo.go            # 仓储接口
│   ├── dao.go             # 仓储 GORM 实现 + converter
│   └── cache.go           # 缓存接口
└── wire.go                # 模块内 IoC 装配，返回 *api.{Entity}API
```

## 文件命名约定

- 文件名用 `snake_case`（如 `http_handler.go`、`app_service.go`）
- 聚合根文件用 `entity.go`（非聚合名）
- 仓储接口文件用 `repo.go`，实现用 `dao.go`
- 端口接口文件用 `ports.go`
- 应用服务文件用 `app_service.go`
- IoC 装配文件用 `wire.go`

## 配置管理

- `config/config.example.yaml` 提交到仓库（配置模板）
- `config/config.yaml` 本地使用，被 `.gitignore` 排除
- 环境变量覆盖用 `DD_` 前缀（如 `DD_DATABASE_HOST`）

## 构建验证

每次修改后必须通过：

```bash
go build ./...   # 编译通过
go vet ./...     # 静态分析通过
```
