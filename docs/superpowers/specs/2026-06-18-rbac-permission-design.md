# RBAC 权限管理系统设计

## 概述

为 dd-frame 引入基于 DDD 的 RBAC 权限管理机制，包含完整的用户认证、角色管理、权限控制和请求拦截能力。

## 设计决策

| 维度 | 决策 |
|------|------|
| 范围 | 完整管理模块（登录/用户/角色/权限 CRUD） |
| 权限模型 | 资源:操作粒度（如 `order:create`） |
| 用户模型 | 内置基础用户模型 + 端口扩展点 |
| 权限拦截 | 路由声明式 `RequirePermission("xxx")` |
| 模块位置 | `middleware/`（拦截）+ `pkg/auth/`（JWT 工具）+ `internal/auth/`（DDD 业务模块） |

## 领域模型

### 实体关系

```
User ──M:N── Role ──M:N── Permission
```

- **User**：聚合根，持有加密密码、基础信息、状态
- **Role**：聚合根，唯一 code（如 `admin`、`operator`）
- **Permission**：聚合根，code = `resource:action`（如 `order:create`）
- **UserRole**：关联实体，user_id ↔ role_id
- **RolePermission**：关联实体，role_id ↔ permission_id

### 领域服务

| 接口 | 说明 |
|------|------|
| `PasswordHasher` | 密码加密/校验（bcrypt），端口接口 |
| `TokenGenerator` | JWT Token 签发/校验，由 `pkg/auth/` 提供实现 |

### 端口扩展点

| 端口 | 说明 |
|------|------|
| `UserExtension` | 允许业务方扩展用户属性（如 company_id、department） |
| `AuditLogger` | 审计日志端口，记录权限变更、登录事件 |
| `CacheInvalidator` | 权限缓存失效端口，角色/权限变更时清除相关缓存 |

## API 设计

### Auth API（`/api/v1/auth`）— 认证（公开路由）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/auth/login` | 无 | 用户名+密码登录，返回 JWT Token |
| POST | `/auth/refresh` | JWT | 刷新 Token |
| POST | `/auth/logout` | JWT | 注销（客户端清除 Token） |
| GET | `/auth/me` | JWT | 获取当前用户信息 + 权限列表 |
| PUT | `/auth/password` | JWT | 修改当前用户密码 |

### User API（`/api/v1/user`）— 用户管理

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/user` | `user:list` | 分页查询用户列表 |
| GET | `/user/:id` | `user:read` | 获取用户详情（含角色） |
| POST | `/user` | `user:create` | 创建用户 |
| PUT | `/user/:id` | `user:update` | 更新用户信息 |
| DELETE | `/user/:id` | `user:delete` | 禁用用户（软删除） |
| POST | `/user/:id/roles` | `user:assign_role` | 分配角色 |
| DELETE | `/user/:id/roles/:roleCode` | `user:revoke_role` | 移除角色 |

### Role API（`/api/v1/role`）— 角色管理

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/role` | `role:list` | 查询角色列表 |
| GET | `/role/:code` | `role:read` | 获取角色详情（含权限） |
| POST | `/role` | `role:create` | 创建角色 |
| PUT | `/role/:code` | `role:update` | 更新角色 |
| DELETE | `/role/:code` | `role:delete` | 删除角色 |
| POST | `/role/:code/permissions` | `role:assign_perm` | 分配权限 |
| DELETE | `/role/:code/permissions/:permCode` | `role:revoke_perm` | 移除权限 |

### Permission API（`/api/v1/permission`）— 权限管理

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/permission` | `permission:list` | 查询权限列表 |
| POST | `/permission` | `permission:create` | 创建权限定义 |
| PUT | `/permission/:code` | `permission:update` | 更新权限 |
| DELETE | `/permission/:code` | `permission:delete` | 删除权限 |

## JWT Token 设计

### Token 结构

```json
{
  "sub": 1001,
  "username": "admin",
  "roles": ["admin", "operator"],
  "exp": 1719000000,
  "iat": 1718913600
}
```

Token 仅存储 `user_id` + `username` + `roles`，不存权限列表（权限实时查询，保证变更后立即生效）。

### Context 注入

JWT 中间件校验通过后，将 `*AuthUser` 结构体注入 `gin.Context`：

```go
type AuthUser struct {
    UserID   int64
    Username string
    Roles    []string
}
```

下游通过 `auth.CurrentUser(c)` 或 `auth.MustCurrentUser(c)` 获取。

## 中间件设计

### JWT 认证中间件（`middleware/auth.go`）

- 从 `Authorization: Bearer <token>` 提取并校验 Token
- 有效时注入 `*AuthUser` 到 Context
- 提供 `RequireAuth()`（强制认证）和 `OptionalAuth()`（可选认证）

### RBAC 权限中间件（`middleware/rbac.go`）

| 函数 | 说明 |
|------|------|
| `RequirePermission(codes ...string)` | 要求持有任一指定权限码 |
| `RequireAllPermissions(codes ...string)` | 要求持有所有指定权限码 |
| `RequireRole(roles ...string)` | 要求持有任一指定角色 |

**权限校验流程：**

1. 从 Context 获取 `AuthUser`（JWT 中间件已注入）
2. 超级管理员角色（配置 `admin_role`）直接放行
3. 查询用户关联的所有权限码（优先读 Redis 缓存）
4. 匹配请求所需权限码，命中放行，否则 403

**缓存策略：**

- 首次查询后缓存到 Redis（key: `rbac:perms:{userID}`，TTL 可配置）
- 角色/权限变更时通过 `CacheInvalidator` 端口清除缓存
- 无 Redis 时降级为每次实时查询

### 中间件注册顺序

```
Recovery → CORS → RequestID → Logger → RequireAuth → 路由（RequirePermission）
```

## 数据库设计

### users 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | 用户 ID |
| username | VARCHAR(64) UNIQUE | 用户名 |
| password | VARCHAR(255) | bcrypt 哈希 |
| nickname | VARCHAR(64) | 昵称 |
| email | VARCHAR(128) | 邮箱 |
| phone | VARCHAR(20) | 手机号 |
| status | TINYINT DEFAULT 1 | 1=启用 0=禁用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### roles 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INT PK AUTO_INCREMENT | 角色 ID |
| code | VARCHAR(32) UNIQUE | 角色编码 |
| name | VARCHAR(64) | 角色名称 |
| description | VARCHAR(255) | 描述 |
| status | TINYINT DEFAULT 1 | 状态 |
| created_at | DATETIME | 创建时间 |

### permissions 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INT PK AUTO_INCREMENT | 权限 ID |
| code | VARCHAR(64) UNIQUE | 权限码（resource:action） |
| resource | VARCHAR(32) | 资源名 |
| action | VARCHAR(32) | 操作名 |
| name | VARCHAR(64) | 权限名称 |
| description | VARCHAR(255) | 描述 |
| created_at | DATETIME | 创建时间 |

### user_roles 表

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | BIGINT | 用户 ID |
| role_id | INT | 角色 ID |
| created_at | DATETIME | 创建时间 |
| PK | (user_id, role_id) | 联合主键 |

### role_permissions 表

| 字段 | 类型 | 说明 |
|------|------|------|
| role_id | INT | 角色 ID |
| permission_id | INT | 权限 ID |
| created_at | DATETIME | 创建时间 |
| PK | (role_id, permission_id) | 联合主键 |

### 设计要点

- users 表用 `status=0` 软删除，不做物理删除
- 不使用物理外键，应用层保证一致性
- 高频查询字段（status、resource、关联表外键）建索引

### 种子数据

应用首次启动时自动插入：

- 用户：`admin` / `admin123`
- 角色：`admin`、`operator`、`viewer`
- 权限：`order:*` / `user:*` / `role:*` / `permission:*` 全套 CRUD 权限
- 角色-权限：admin → 全部，operator → 业务操作，viewer → 只读

## 文件结构

```
middleware/                             框架级中间件
├── recovery.go                         Panic 恢复
├── cors.go                             CORS 跨域
├── request_id.go                       X-Request-ID
├── logger.go                           请求日志
├── auth.go                             JWT 认证中间件
└── rbac.go                             RBAC 权限中间件

pkg/auth/                               认证共享工具
├── jwt.go                              Token 签发/解析/刷新
└── context.go                          AuthUser 实体 + CurrentUser()

internal/auth/                          DDD 分层 auth 模块
├── domain/
│   ├── entity.go                       User / Role / Permission 聚合根
│   ├── enums.go                        UserStatus / RoleStatus
│   ├── errors.go                       领域错误
│   └── service.go                      PasswordHasher 端口接口
├── biz/
│   ├── service.go                      AuthBizService 用例编排
│   └── ports.go                        CacheInvalidator / AuditLogger 端口
├── service/
│   └── app_service.go                  AuthAppService DTO 转换
├── api/
│   └── http_handler.go                 HTTP Handler + 路由注册
├── model/
│   ├── repo.go                         仓储接口
│   └── dao.go                          GORM 实现 + converter
└── wire.go                             模块内 IoC 装配
```

## 集成方式

### app/wire.go 变更

```go
// 公开路由（无需 JWT）
public := r.Group("/api/v1")

// 认证路由（需 JWT）
v1 := r.Group("/api/v1")
v1.Use(middleware.RequireAuth())

// auth 模块
authAPI, permissionChecker := auth.Wire()
authAPI.RegisterPublicRoutes(public)   // /auth/login
authAPI.RegisterRoutes(v1)             // /user/*, /role/*, /permission/*

// 业务模块路由自行声明 RequirePermission
orderAPI := order.Wire()
orderAPI.RegisterRoutes(v1)
```

### config.example.yaml 新增

```yaml
rbac:
  admin_role: admin              # 超级管理员角色 code
  permission_cache_ttl: 300      # 权限缓存有效期（秒），0=不缓存
  seed_enabled: true             # 是否自动初始化种子数据
```

## PermissionChecker 工具方法

### 接口定义（`pkg/auth/checker.go`）

```go
type PermissionChecker interface {
    HasPermission(ctx context.Context, userID int64, code string) (bool, error)
    HasAllPermissions(ctx context.Context, userID int64, codes []string) (bool, error)
    HasRole(ctx context.Context, userID int64, role string) (bool, error)
}
```

### 使用方式

| 场景 | 说明 |
|------|------|
| Handler 条件逻辑 | 根据权限决定返回不同数据（如管理员看全部，普通用户看自己） |
| BizService 编程式校验 | 在业务编排中判断权限，不依赖中间件 |
| 中间件内部 | RBAC 中间件同样调用 PermissionChecker.HasPermission |

### 依赖注入

- `internal/auth/wire.go` 创建 PermissionChecker 实现
- `auth.Wire()` 返回 `(API, PermissionChecker)`
- `app/wire.go` 将 PermissionChecker 传给 middleware 和业务模块
- 业务模块通过构造函数注入 PermissionChecker

## 请求鉴权流程

```
Client → Recovery → CORS → RequestID → Logger
  → RequireAuth() [JWT 校验 → AuthUser 注入]
    → 路由匹配 → RequirePermission("order:create") [RBAC 校验]
      → Handler [auth.CurrentUser(c) 获取用户信息]
        → Response
```
