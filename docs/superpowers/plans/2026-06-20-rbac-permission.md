# RBAC 权限管理系统实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 dd-frame 构建完整的 RBAC 权限管理系统，包含 JWT 认证、角色权限 CRUD、路由声明式拦截和 PermissionChecker 工具方法。

**Architecture:** 三层架构 — `pkg/auth/`（JWT 工具 + AuthUser 实体 + PermissionChecker 接口）、`middleware/`（JWT/RBAC 中间件）、`internal/auth/`（DDD 分层 auth 业务模块）。auth 模块通过 `auth.Wire()` 返回 API handler 和 PermissionChecker，由 `app/wire.go` 统一装配。

**Tech Stack:** Go 1.26, Gin, GORM, golang-jwt/jwt/v5, golang.org/x/crypto/bcrypt, Redis (optional)

**Spec:** `docs/superpowers/specs/2026-06-18-rbac-permission-design.md`

---

## 文件结构总览

### 新增文件

| 文件 | 职责 |
|------|------|
| `pkg/auth/jwt.go` | JWT Token 签发/解析/刷新 |
| `pkg/auth/context.go` | AuthUser 实体 + CurrentUser() 辅助函数 |
| `pkg/auth/checker.go` | PermissionChecker 接口 |
| `middleware/recovery.go` | Panic 恢复中间件 |
| `middleware/cors.go` | CORS 跨域中间件 |
| `middleware/request_id.go` | X-Request-ID 中间件 |
| `middleware/logger.go` | 请求日志中间件 |
| `middleware/auth.go` | JWT 认证中间件 |
| `middleware/rbac.go` | RBAC 权限中间件 |
| `internal/auth/domain/entity.go` | User / Role / Permission 聚合根 |
| `internal/auth/domain/enums.go` | UserStatus 枚举 |
| `internal/auth/domain/errors.go` | 领域错误定义 |
| `internal/auth/domain/service.go` | PasswordHasher 端口接口 |
| `internal/auth/model/repo.go` | UserRepo / RoleRepo / PermissionRepo 仓储接口 |
| `internal/auth/model/dao.go` | GORM 仓储实现 + DB 模型 + converter |
| `internal/auth/biz/ports.go` | CacheInvalidator / AuditLogger 端口 |
| `internal/auth/biz/service.go` | AuthBizService 用例编排 |
| `internal/auth/service/app_service.go` | AuthAppService DTO 转换 |
| `internal/auth/api/http_handler.go` | HTTP Handler + 路由注册 |
| `internal/auth/wire.go` | 模块内 IoC 装配 + 种子数据 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `app/config.go` | 新增 RBACConfig 结构体 |
| `app/wire.go` | 集成 auth 模块 + 中间件 |
| `config/config.example.yaml` | 新增 rbac 配置段 |
| `go.mod` | 新增 golang-jwt/jwt/v5 依赖 |

---

### Task 1: 安装依赖 + JWT 工具（pkg/auth）

**Files:**
- Create: `pkg/auth/jwt.go`
- Create: `pkg/auth/context.go`
- Create: `pkg/auth/checker.go`
- Modify: `go.mod`

- [ ] **Step 1: 安装 golang-jwt/jwt/v5**

```bash
cd /Users/mayujian/all_code/go_code/dd-frame
go get github.com/golang-jwt/jwt/v5
```

注意：`golang.org/x/crypto` 已在 go.mod 中（间接依赖），bcrypt 可直接使用。

- [ ] **Step 2: 创建 `pkg/auth/jwt.go`**

```go
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 自定义声明
type Claims struct {
	UserID   int64    `json:"sub"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret    []byte
	expiresIn time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secret string, expiresHours int) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		expiresIn: time.Duration(expiresHours) * time.Hour,
	}
}

// Generate 签发 Token
func (m *JWTManager) Generate(userID int64, username string, roles []string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse 解析并校验 Token
func (m *JWTManager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
```

- [ ] **Step 3: 创建 `pkg/auth/context.go`**

```go
package auth

import "github.com/gin-gonic/gin"

// AuthUser 认证用户实体（从 JWT Claims 解析）
type AuthUser struct {
	UserID   int64    `json:"userId"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// IsAdmin 是否为超级管理员
func (a *AuthUser) IsAdmin() bool {
	for _, r := range a.Roles {
		if r == "admin" {
			return true
		}
	}
	return false
}

// HasRole 是否持有指定角色
func (a *AuthUser) HasRole(role string) bool {
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}
	return false
}

const contextKeyAuthUser = "auth_user"

// CurrentUser 从 gin.Context 获取当前认证用户
func CurrentUser(c *gin.Context) (*AuthUser, bool) {
	val, exists := c.Get(contextKeyAuthUser)
	if !exists {
		return nil, false
	}
	return val.(*AuthUser), true
}

// MustCurrentUser 获取当前用户，未认证时 panic
func MustCurrentUser(c *gin.Context) *AuthUser {
	user, ok := CurrentUser(c)
	if !ok {
		panic("auth_user not found in context")
	}
	return user
}

// SetAuthUser 将 AuthUser 注入到 gin.Context（供中间件内部使用）
func SetAuthUser(c *gin.Context, user *AuthUser) {
	c.Set(contextKeyAuthUser, user)
}
```

- [ ] **Step 4: 创建 `pkg/auth/checker.go`**

```go
package auth

import "context"

// PermissionChecker 权限校验器接口
//
// 由 internal/auth 模块实现，middleware 和业务代码通过此接口判断权限。
type PermissionChecker interface {
	// HasPermission 判断用户是否持有指定权限码
	HasPermission(ctx context.Context, userID int64, code string) (bool, error)

	// HasAllPermissions 判断用户是否持有所有指定权限码
	HasAllPermissions(ctx context.Context, userID int64, codes []string) (bool, error)

	// HasRole 判断用户是否持有指定角色
	HasRole(ctx context.Context, userID int64, role string) (bool, error)
}
```

- [ ] **Step 5: 验证编译**

```bash
go build ./pkg/auth/...
```

Expected: 零错误

- [ ] **Step 6: 提交**

```bash
git add pkg/auth/ go.mod go.sum
git commit -m "feat(auth): add pkg/auth with JWT manager, AuthUser, and PermissionChecker interface"
```

---

### Task 2: Gin 基础中间件

**Files:**
- Create: `middleware/recovery.go`
- Create: `middleware/cors.go`
- Create: `middleware/request_id.go`
- Create: `middleware/logger.go`

- [ ] **Step 1: 创建 `middleware/recovery.go`**

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	applog "github.com/example/dd-frame/pkg/log"
)

// Recovery Panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				applog.Error("panic recovered", "error", r, "path", c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    50000,
					"message": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
```

- [ ] **Step 2: 创建 `middleware/cors.go`**

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: 创建 `middleware/request_id.go`**

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"

// RequestID 请求 ID 中间件
//
// 优先使用客户端传入的 X-Request-ID，否则生成新的 UUID。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Header(headerRequestID, rid)
		c.Set("request_id", rid)
		c.Next()
	}
}
```

- [ ] **Step 4: 创建 `middleware/logger.go`**

```go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	applog "github.com/example/dd-frame/pkg/log"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		applog.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"request_id", c.GetString("request_id"),
		)
	}
}
```

- [ ] **Step 5: 验证编译**

```bash
go build ./middleware/...
```

Expected: 零错误

- [ ] **Step 6: 提交**

```bash
git add middleware/
git commit -m "feat(middleware): add recovery, cors, request_id, and logger middleware"
```

---

### Task 3: Auth + RBAC 中间件

**Files:**
- Create: `middleware/auth.go`
- Create: `middleware/rbac.go`

- [ ] **Step 1: 创建 `middleware/auth.go`**

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/auth"
	"github.com/example/dd-frame/pkg/response"
)

// RequireAuth JWT 强制认证中间件
//
// 从 Authorization: Bearer <token> 提取并校验 Token，
// 有效时将 *auth.AuthUser 注入 Context，否则返回 401。
func RequireAuth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, 40100, "authorization header required")
			c.Abort()
			return
		}

		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, 40101, "invalid or expired token")
			c.Abort()
			return
		}

		auth.SetAuthUser(c, &auth.AuthUser{
			UserID:   claims.UserID,
			Username: claims.Username,
			Roles:    claims.Roles,
		})
		c.Next()
	}
}

// OptionalAuth 可选认证中间件
//
// 有 Token 则解析注入，无 Token 也放行。
func OptionalAuth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}

		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			c.Next()
			return
		}

		auth.SetAuthUser(c, &auth.AuthUser{
			UserID:   claims.UserID,
			Username: claims.Username,
			Roles:    claims.Roles,
		})
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
```

- [ ] **Step 2: 创建 `middleware/rbac.go`**

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/auth"
	"github.com/example/dd-frame/pkg/response"
)

// RequirePermission 要求持有任一指定权限码
//
// 使用方式：
//
//	orderGroup.POST("", middleware.RequirePermission(checker, "order:create"), handler.Create)
func RequirePermission(checker auth.PermissionChecker, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.CurrentUser(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, 40100, "unauthorized")
			c.Abort()
			return
		}

		// 超级管理员直接放行
		if user.IsAdmin() {
			c.Next()
			return
		}

		for _, code := range codes {
			ok, err := checker.HasPermission(c.Request.Context(), user.UserID, code)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, 50000, "permission check failed")
				c.Abort()
				return
			}
			if ok {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, 40300, "permission denied")
		c.Abort()
	}
}

// RequireAllPermissions 要求持有所有指定权限码
func RequireAllPermissions(checker auth.PermissionChecker, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.CurrentUser(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, 40100, "unauthorized")
			c.Abort()
			return
		}

		if user.IsAdmin() {
			c.Next()
			return
		}

		ok, err := checker.HasAllPermissions(c.Request.Context(), user.UserID, codes)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 50000, "permission check failed")
			c.Abort()
			return
		}
		if !ok {
			response.Error(c, http.StatusForbidden, 40300, "permission denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole 要求持有任一指定角色
func RequireRole(checker auth.PermissionChecker, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.CurrentUser(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, 40100, "unauthorized")
			c.Abort()
			return
		}

		if user.IsAdmin() {
			c.Next()
			return
		}

		for _, role := range roles {
			ok, err := checker.HasRole(c.Request.Context(), user.UserID, role)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, 50000, "role check failed")
				c.Abort()
				return
			}
			if ok {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, 40300, "role required")
		c.Abort()
	}
}
```

- [ ] **Step 3: 验证编译**

```bash
go build ./middleware/...
```

Expected: 零错误

- [ ] **Step 4: 提交**

```bash
git add middleware/auth.go middleware/rbac.go
git commit -m "feat(middleware): add JWT auth and RBAC permission middleware"
```

---

### Task 4: Auth 领域层（domain）

**Files:**
- Create: `internal/auth/domain/entity.go`
- Create: `internal/auth/domain/enums.go`
- Create: `internal/auth/domain/errors.go`
- Create: `internal/auth/domain/service.go`

- [ ] **Step 1: 创建 `internal/auth/domain/entity.go`**

```go
package domain

import "time"

// User 用户聚合根
type User struct {
	ID        int64      // 主键
	Username  string     // 用户名（唯一）
	Password  string     // bcrypt 哈希密码
	Nickname  string     // 昵称
	Email     string     // 邮箱
	Phone     string     // 手机号
	Status    UserStatus // 状态
	Roles     []Role     // 关联角色（查询时填充）
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsActive 用户是否启用
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// Disable 禁用用户
func (u *User) Disable() {
	u.Status = UserStatusDisabled
}

// UpdateProfile 更新基础信息
func (u *User) UpdateProfile(nickname, email, phone string) {
	u.Nickname = nickname
	u.Email = email
	u.Phone = phone
}

// Role 角色聚合根
type Role struct {
	ID          int64  // 主键
	Code        string // 角色编码（唯一，如 admin）
	Name        string // 角色名称
	Description string // 描述
	Status      int    // 1=启用 0=禁用
	Permissions []Permission // 关联权限（查询时填充）
	CreatedAt   time.Time
}

// IsActive 角色是否启用
func (r *Role) IsActive() bool {
	return r.Status == 1
}

// Permission 权限聚合根
type Permission struct {
	ID          int64  // 主键
	Code        string // 权限码（唯一，如 order:create）
	Resource    string // 资源名（如 order）
	Action      string // 操作名（如 create）
	Name        string // 权限名称
	Description string // 描述
	CreatedAt   time.Time
}

// BuildCode 从 resource + action 构建权限码
func BuildCode(resource, action string) string {
	return resource + ":" + action
}
```

- [ ] **Step 2: 创建 `internal/auth/domain/enums.go`**

```go
package domain

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusDisabled UserStatus = 0 // 禁用
	UserStatusActive   UserStatus = 1 // 启用
)

// IsValid 校验状态值
func (s UserStatus) IsValid() bool {
	return s == UserStatusDisabled || s == UserStatusActive
}

// String 返回描述
func (s UserStatus) String() string {
	switch s {
	case UserStatusActive:
		return "启用"
	case UserStatusDisabled:
		return "禁用"
	default:
		return "未知"
	}
}
```

- [ ] **Step 3: 创建 `internal/auth/domain/errors.go`**

```go
package domain

import "fmt"

const (
	ReasonUserNotFound       = "USER_NOT_FOUND"
	ReasonUserAlreadyExists  = "USER_ALREADY_EXISTS"
	ReasonUserDisabled       = "USER_DISABLED"
	ReasonInvalidCredentials = "INVALID_CREDENTIALS"
	ReasonRoleNotFound       = "ROLE_NOT_FOUND"
	ReasonRoleAlreadyExists  = "ROLE_ALREADY_EXISTS"
	ReasonPermNotFound       = "PERMISSION_NOT_FOUND"
	ReasonPermAlreadyExists  = "PERMISSION_ALREADY_EXISTS"
)

func ErrUserNotFound(identifier string) error {
	return fmt.Errorf("[%s] user not found: %s", ReasonUserNotFound, identifier)
}

func ErrUserAlreadyExists(username string) error {
	return fmt.Errorf("[%s] user already exists: %s", ReasonUserAlreadyExists, username)
}

func ErrUserDisabled() error {
	return fmt.Errorf("[%s] user is disabled", ReasonUserDisabled)
}

func ErrInvalidCredentials() error {
	return fmt.Errorf("[%s] invalid username or password", ReasonInvalidCredentials)
}

func ErrRoleNotFound(code string) error {
	return fmt.Errorf("[%s] role not found: %s", ReasonRoleNotFound, code)
}

func ErrRoleAlreadyExists(code string) error {
	return fmt.Errorf("[%s] role already exists: %s", ReasonRoleAlreadyExists, code)
}

func ErrPermNotFound(code string) error {
	return fmt.Errorf("[%s] permission not found: %s", ReasonPermNotFound, code)
}

func ErrPermAlreadyExists(code string) error {
	return fmt.Errorf("[%s] permission already exists: %s", ReasonPermAlreadyExists, code)
}
```

- [ ] **Step 4: 创建 `internal/auth/domain/service.go`**

```go
package domain

// PasswordHasher 密码加密端口接口
//
// 实现层在 biz 层（bcrypt 实现），领域层仅定义契约。
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}
```

- [ ] **Step 5: 验证编译**

```bash
go build ./internal/auth/domain/...
```

Expected: 零错误

- [ ] **Step 6: 提交**

```bash
git add internal/auth/domain/
git commit -m "feat(auth): add domain layer — entities, enums, errors, PasswordHasher port"
```

---

### Task 5: Auth 仓储层（model）

**Files:**
- Create: `internal/auth/model/repo.go`
- Create: `internal/auth/model/dao.go`

- [ ] **Step 1: 创建 `internal/auth/model/repo.go`**

```go
package model

import (
	"context"

	"github.com/example/dd-frame/internal/auth/domain"
)

// UserRepo 用户仓储接口
type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	QueryByID(ctx context.Context, id int64) (*domain.User, error)
	QueryByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateStatus(ctx context.Context, id int64, status domain.UserStatus) error
	List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
	AssignRole(ctx context.Context, userID int64, roleID int64) error
	RevokeRole(ctx context.Context, userID int64, roleCode string) error
	QueryRolesByUserID(ctx context.Context, userID int64) ([]domain.Role, error)
	QueryPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error)
}

// RoleRepo 角色仓储接口
type RoleRepo interface {
	Create(ctx context.Context, role *domain.Role) error
	QueryByCode(ctx context.Context, code string) (*domain.Role, error)
	QueryByID(ctx context.Context, id int64) (*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, code string) error
	List(ctx context.Context) ([]*domain.Role, error)
	AssignPermission(ctx context.Context, roleID int64, permissionID int64) error
	RevokePermission(ctx context.Context, roleID int64, permCode string) error
	QueryPermissionsByRoleID(ctx context.Context, roleID int64) ([]domain.Permission, error)
}

// PermissionRepo 权限仓储接口
type PermissionRepo interface {
	Create(ctx context.Context, perm *domain.Permission) error
	QueryByCode(ctx context.Context, code string) (*domain.Permission, error)
	Update(ctx context.Context, perm *domain.Permission) error
	Delete(ctx context.Context, code string) error
	List(ctx context.Context, resource string) ([]*domain.Permission, error)
}
```

- [ ] **Step 2: 创建 `internal/auth/model/dao.go`**

```go
package model

import (
	"context"
	"time"

	"github.com/example/dd-frame/internal/auth/domain"
	"gorm.io/gorm"
)

// ==================== DB 模型 ====================

// UserModel 用户表
type UserModel struct {
	ID        int64     `gorm:"primary_key;auto_increment" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:64" json:"username"`
	Password  string    `gorm:"size:255" json:"-"`
	Nickname  string    `gorm:"size:64" json:"nickname"`
	Email     string    `gorm:"size:128" json:"email"`
	Phone     string    `gorm:"size:20" json:"phone"`
	Status    int       `gorm:"default:1;index" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserModel) TableName() string { return "users" }

// RoleModel 角色表
type RoleModel struct {
	ID          int64     `gorm:"primary_key;auto_increment" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:32" json:"code"`
	Name        string    `gorm:"size:64" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RoleModel) TableName() string { return "roles" }

// PermissionModel 权限表
type PermissionModel struct {
	ID          int64     `gorm:"primary_key;auto_increment" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:64" json:"code"`
	Resource    string    `gorm:"size:32;index" json:"resource"`
	Action      string    `gorm:"size:32" json:"action"`
	Name        string    `gorm:"size:64" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PermissionModel) TableName() string { return "permissions" }

// UserRoleModel 用户-角色关联表
type UserRoleModel struct {
	UserID    int64     `gorm:"primaryKey" json:"user_id"`
	RoleID    int64     `gorm:"primaryKey;index" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserRoleModel) TableName() string { return "user_roles" }

// RolePermissionModel 角色-权限关联表
type RolePermissionModel struct {
	RoleID       int64     `gorm:"primaryKey" json:"role_id"`
	PermissionID int64     `gorm:"primaryKey;index" json:"permission_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RolePermissionModel) TableName() string { return "role_permissions" }

// ==================== DAO 实现 ====================

// 编译期校验
var (
	_ domain.UserRepo       = (*UserDAO)(nil)
	_ domain.RoleRepo       = (*RoleDAO)(nil)
	_ domain.PermissionRepo = (*PermissionDAO)(nil)
)

// ---------- UserDAO ----------

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) Create(ctx context.Context, user *domain.User) error {
	m := userToModel(user)
	if err := d.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	return nil
}

func (d *UserDAO) QueryByID(ctx context.Context, id int64) (*domain.User, error) {
	var m UserModel
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return modelToUser(&m), nil
}

func (d *UserDAO) QueryByUsername(ctx context.Context, username string) (*domain.User, error) {
	var m UserModel
	if err := d.db.WithContext(ctx).Where("username = ?", username).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return modelToUser(&m), nil
}

func (d *UserDAO) Update(ctx context.Context, user *domain.User) error {
	return d.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"nickname": user.Nickname,
			"email":    user.Email,
			"phone":    user.Phone,
			"status":   int(user.Status),
			"password": user.Password,
		}).Error
}

func (d *UserDAO) UpdateStatus(ctx context.Context, id int64, status domain.UserStatus) error {
	return d.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).
		Update("status", int(status)).Error
}

func (d *UserDAO) List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	var models []UserModel
	var total int64

	d.db.WithContext(ctx).Model(&UserModel{}).Count(&total)
	offset := (page - 1) * pageSize
	if err := d.db.WithContext(ctx).Offset(offset).Limit(pageSize).
		Order("id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = modelToUser(&models[i])
	}
	return users, total, nil
}

func (d *UserDAO) AssignRole(ctx context.Context, userID int64, roleID int64) error {
	ur := UserRoleModel{UserID: userID, RoleID: roleID}
	return d.db.WithContext(ctx).FirstOrCreate(&ur, ur).Error
}

func (d *UserDAO) RevokeRole(ctx context.Context, userID int64, roleCode string) error {
	var role RoleModel
	if err := d.db.WithContext(ctx).Where("code = ?", roleCode).First(&role).Error; err != nil {
		return err
	}
	return d.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, role.ID).
		Delete(&UserRoleModel{}).Error
}

func (d *UserDAO) QueryRolesByUserID(ctx context.Context, userID int64) ([]domain.Role, error) {
	var roles []RoleModel
	if err := d.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Role, len(roles))
	for i, r := range roles {
		result[i] = roleModelToDomain(&r)
	}
	return result, nil
}

func (d *UserDAO) QueryPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error) {
	var codes []string
	if err := d.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

// ---------- RoleDAO ----------

type RoleDAO struct {
	db *gorm.DB
}

func NewRoleDAO(db *gorm.DB) *RoleDAO {
	return &RoleDAO{db: db}
}

func (d *RoleDAO) Create(ctx context.Context, role *domain.Role) error {
	m := roleToModel(role)
	if err := d.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	role.ID = m.ID
	return nil
}

func (d *RoleDAO) QueryByCode(ctx context.Context, code string) (*domain.Role, error) {
	var m RoleModel
	if err := d.db.WithContext(ctx).Where("code = ?", code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return roleModelToDomain(&m), nil
}

func (d *RoleDAO) QueryByID(ctx context.Context, id int64) (*domain.Role, error) {
	var m RoleModel
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return roleModelToDomain(&m), nil
}

func (d *RoleDAO) Update(ctx context.Context, role *domain.Role) error {
	return d.db.WithContext(ctx).Model(&RoleModel{}).Where("code = ?", role.Code).
		Updates(map[string]interface{}{
			"name":        role.Name,
			"description": role.Description,
			"status":      role.Status,
		}).Error
}

func (d *RoleDAO) Delete(ctx context.Context, code string) error {
	return d.db.WithContext(ctx).Where("code = ?", code).Delete(&RoleModel{}).Error
}

func (d *RoleDAO) List(ctx context.Context) ([]*domain.Role, error) {
	var models []RoleModel
	if err := d.db.WithContext(ctx).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(models))
	for i := range models {
		roles[i] = roleModelToDomain(&models[i])
	}
	return roles, nil
}

func (d *RoleDAO) AssignPermission(ctx context.Context, roleID int64, permissionID int64) error {
	rp := RolePermissionModel{RoleID: roleID, PermissionID: permissionID}
	return d.db.WithContext(ctx).FirstOrCreate(&rp, rp).Error
}

func (d *RoleDAO) RevokePermission(ctx context.Context, roleID int64, permCode string) error {
	var perm PermissionModel
	if err := d.db.WithContext(ctx).Where("code = ?", permCode).First(&perm).Error; err != nil {
		return err
	}
	return d.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, perm.ID).
		Delete(&RolePermissionModel{}).Error
}

func (d *RoleDAO) QueryPermissionsByRoleID(ctx context.Context, roleID int64) ([]domain.Permission, error) {
	var perms []PermissionModel
	if err := d.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Permission, len(perms))
	for i, p := range perms {
		result[i] = permModelToDomain(&p)
	}
	return result, nil
}

// ---------- PermissionDAO ----------

type PermissionDAO struct {
	db *gorm.DB
}

func NewPermissionDAO(db *gorm.DB) *PermissionDAO {
	return &PermissionDAO{db: db}
}

func (d *PermissionDAO) Create(ctx context.Context, perm *domain.Permission) error {
	m := permToModel(perm)
	if err := d.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	perm.ID = m.ID
	return nil
}

func (d *PermissionDAO) QueryByCode(ctx context.Context, code string) (*domain.Permission, error) {
	var m PermissionModel
	if err := d.db.WithContext(ctx).Where("code = ?", code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return permModelToDomain(&m), nil
}

func (d *PermissionDAO) Update(ctx context.Context, perm *domain.Permission) error {
	return d.db.WithContext(ctx).Model(&PermissionModel{}).Where("code = ?", perm.Code).
		Updates(map[string]interface{}{
			"name":        perm.Name,
			"description": perm.Description,
		}).Error
}

func (d *PermissionDAO) Delete(ctx context.Context, code string) error {
	return d.db.WithContext(ctx).Where("code = ?", code).Delete(&PermissionModel{}).Error
}

func (d *PermissionDAO) List(ctx context.Context, resource string) ([]*domain.Permission, error) {
	var models []PermissionModel
	query := d.db.WithContext(ctx).Order("code ASC")
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	perms := make([]*domain.Permission, len(models))
	for i := range models {
		perms[i] = permModelToDomain(&models[i])
	}
	return perms, nil
}

// ==================== Converter ====================

func userToModel(u *domain.User) *UserModel {
	return &UserModel{
		ID:       u.ID,
		Username: u.Username,
		Password: u.Password,
		Nickname: u.Nickname,
		Email:    u.Email,
		Phone:    u.Phone,
		Status:   int(u.Status),
	}
}

func modelToUser(m *UserModel) *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		ID:        m.ID,
		Username:  m.Username,
		Password:  m.Password,
		Nickname:  m.Nickname,
		Email:     m.Email,
		Phone:     m.Phone,
		Status:    domain.UserStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func roleToModel(r *domain.Role) *RoleModel {
	return &RoleModel{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Status:      r.Status,
	}
}

func roleModelToDomain(m *RoleModel) domain.Role {
	return domain.Role{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
	}
}

func permToModel(p *domain.Permission) *PermissionModel {
	return &PermissionModel{
		ID:          p.ID,
		Code:        p.Code,
		Resource:    p.Resource,
		Action:      p.Action,
		Name:        p.Name,
		Description: p.Description,
	}
}

func permModelToDomain(m *PermissionModel) domain.Permission {
	return domain.Permission{
		ID:          m.ID,
		Code:        m.Code,
		Resource:    m.Resource,
		Action:      m.Action,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}
```

注意：repo.go 中的仓储接口需要引用 `domain` 包中的类型。由于 Go import 路径是 `internal/auth/domain`，但 repo 接口中使用的 `domain.User` 等类型需要在 `repo.go` 的 import 中声明为 `authdomain "github.com/example/dd-frame/internal/auth/domain"`。

**重要：** 仓储接口应定义在 `domain` 层或 `model` 层。这里遵循 order 示例的模式（repo 接口在 `model/repo.go`，实现在 `model/dao.go`），所以 repo.go 的 package 是 `model`，import domain 包。

修正 repo.go 中的 import：

```go
import (
    "context"

    authdomain "github.com/example/dd-frame/internal/auth/domain"
)
```

同时 dao.go 中编译期校验的 `_ domain.UserRepo` 应改为 `_ authdomain.UserRepo`（如果 repo.go 和 dao.go 在同一个 package `model` 内，则不需要 import alias，直接用 `domain.User` 即可，因为它们在同一个 package）。

实际上，由于 repo.go 和 dao.go 都在 `package model`，import alias 统一即可。

- [ ] **Step 3: 验证编译**

```bash
go build ./internal/auth/model/...
```

Expected: 零错误（可能需要先确保 domain 包编译通过）

- [ ] **Step 4: 提交**

```bash
git add internal/auth/model/
git commit -m "feat(auth): add repository interfaces and GORM DAO implementations"
```

---

### Task 6: Auth 业务层（biz）

**Files:**
- Create: `internal/auth/biz/ports.go`
- Create: `internal/auth/biz/service.go`

- [ ] **Step 1: 创建 `internal/auth/biz/ports.go`**

```go
package biz

import "context"

// CacheInvalidator 权限缓存失效端口
type CacheInvalidator interface {
	// InvalidateUserPermissions 清除指定用户的权限缓存
	InvalidateUserPermissions(ctx context.Context, userID int64) error
}

// AuditLogger 审计日志端口
type AuditLogger interface {
	LogLogin(ctx context.Context, userID int64, username string, success bool)
	LogPermissionChange(ctx context.Context, operatorID int64, action string, detail string)
}

// StubCacheInvalidator 无缓存时的空实现
type StubCacheInvalidator struct{}

func (s *StubCacheInvalidator) InvalidateUserPermissions(_ context.Context, _ int64) error {
	return nil
}

// StubAuditLogger 无审计时的空实现
type StubAuditLogger struct{}

func (s *StubAuditLogger) LogLogin(_ context.Context, _ int64, _ string, _ bool)          {}
func (s *StubAuditLogger) LogPermissionChange(_ context.Context, _ int64, _ string, _ string) {}
```

- [ ] **Step 2: 创建 `internal/auth/biz/service.go`**

```go
package biz

import (
	"context"
	"fmt"

	"github.com/example/dd-frame/internal/auth/domain"
	authmodel "github.com/example/dd-frame/internal/auth/model"
	"github.com/example/dd-frame/pkg/auth"
)

// AuthBizService auth 业务能力接口
type AuthBizService interface {
	// 认证
	Login(ctx context.Context, username, password string) (string, []string, error) // token, roles, error
	ChangePassword(ctx context.Context, userID int64, oldPwd, newPwd string) error

	// 用户管理
	CreateUser(ctx context.Context, req *CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, id int64) (*domain.User, error)
	UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) error
	DisableUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
	AssignRole(ctx context.Context, userID int64, roleCode string) error
	RevokeRole(ctx context.Context, userID int64, roleCode string) error

	// 角色管理
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*domain.Role, error)
	GetRole(ctx context.Context, code string) (*domain.Role, error)
	UpdateRole(ctx context.Context, code string, req *UpdateRoleRequest) error
	DeleteRole(ctx context.Context, code string) error
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	AssignPermission(ctx context.Context, roleCode string, permCode string) error
	RevokePermission(ctx context.Context, roleCode string, permCode string) error

	// 权限管理
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.Permission, error)
	UpdatePermission(ctx context.Context, code string, req *UpdatePermissionRequest) error
	DeletePermission(ctx context.Context, code string) error
	ListPermissions(ctx context.Context, resource string) ([]*domain.Permission, error)

	// 权限查询（供 PermissionChecker 使用）
	GetUserPermissionCodes(ctx context.Context, userID int64) ([]string, error)
	GetUserRoleCodes(ctx context.Context, userID int64) ([]string, error)
}

// ==================== DTO ====================

type CreateUserRequest struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
}

type UpdateUserRequest struct {
	Nickname string
	Email    string
	Phone    string
}

type CreateRoleRequest struct {
	Code        string
	Name        string
	Description string
}

type UpdateRoleRequest struct {
	Name        string
	Description string
}

type CreatePermissionRequest struct {
	Resource    string
	Action      string
	Name        string
	Description string
}

type UpdatePermissionRequest struct {
	Name        string
	Description string
}

// ==================== 实现 ====================

type authBizService struct {
	userRepo       authmodel.UserRepo
	roleRepo       authmodel.RoleRepo
	permRepo       authmodel.PermissionRepo
	hasher         domain.PasswordHasher
	jwtMgr         *auth.JWTManager
	cache          CacheInvalidator
	audit          AuditLogger
}

// NewAuthBizService 创建 auth 业务服务
func NewAuthBizService(
	userRepo authmodel.UserRepo,
	roleRepo authmodel.RoleRepo,
	permRepo authmodel.PermissionRepo,
	hasher domain.PasswordHasher,
	jwtMgr *auth.JWTManager,
	cache CacheInvalidator,
	audit AuditLogger,
) AuthBizService {
	return &authBizService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		permRepo: permRepo,
		hasher:   hasher,
		jwtMgr:   jwtMgr,
		cache:    cache,
		audit:    audit,
	}
}

// ---------- 认证 ----------

func (s *authBizService) Login(ctx context.Context, username, password string) (string, []string, error) {
	user, err := s.userRepo.QueryByUsername(ctx, username)
	if err != nil {
		return "", nil, fmt.Errorf("query user failed: %w", err)
	}
	if user == nil {
		s.audit.LogLogin(ctx, 0, username, false)
		return "", nil, domain.ErrInvalidCredentials()
	}
	if !user.IsActive() {
		return "", nil, domain.ErrUserDisabled()
	}
	if !s.hasher.Verify(user.Password, password) {
		s.audit.LogLogin(ctx, user.ID, username, false)
		return "", nil, domain.ErrInvalidCredentials()
	}

	roles, err := s.userRepo.QueryRolesByUserID(ctx, user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("query roles failed: %w", err)
	}

	roleCodes := make([]string, len(roles))
	for i, r := range roles {
		roleCodes[i] = r.Code
	}

	token, err := s.jwtMgr.Generate(user.ID, user.Username, roleCodes)
	if err != nil {
		return "", nil, fmt.Errorf("generate token failed: %w", err)
	}

	s.audit.LogLogin(ctx, user.ID, username, true)
	return token, roleCodes, nil
}

func (s *authBizService) ChangePassword(ctx context.Context, userID int64, oldPwd, newPwd string) error {
	user, err := s.userRepo.QueryByID(ctx, userID)
	if err != nil || user == nil {
		return domain.ErrUserNotFound(fmt.Sprintf("%d", userID))
	}
	if !s.hasher.Verify(user.Password, oldPwd) {
		return domain.ErrInvalidCredentials()
	}

	hash, err := s.hasher.Hash(newPwd)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}
	user.Password = hash
	return s.userRepo.Update(ctx, user)
}

// ---------- 用户管理 ----------

func (s *authBizService) CreateUser(ctx context.Context, req *CreateUserRequest) (*domain.User, error) {
	existing, _ := s.userRepo.QueryByUsername(ctx, req.Username)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists(req.Username)
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	user := &domain.User{
		Username: req.Username,
		Password: hash,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   domain.UserStatusActive,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}
	return user, nil
}

func (s *authBizService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	user, err := s.userRepo.QueryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound(fmt.Sprintf("%d", id))
	}
	roles, _ := s.userRepo.QueryRolesByUserID(ctx, id)
	user.Roles = roles
	return user, nil
}

func (s *authBizService) UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) error {
	user, err := s.userRepo.QueryByID(ctx, id)
	if err != nil || user == nil {
		return domain.ErrUserNotFound(fmt.Sprintf("%d", id))
	}
	user.UpdateProfile(req.Nickname, req.Email, req.Phone)
	return s.userRepo.Update(ctx, user)
}

func (s *authBizService) DisableUser(ctx context.Context, id int64) error {
	user, err := s.userRepo.QueryByID(ctx, id)
	if err != nil || user == nil {
		return domain.ErrUserNotFound(fmt.Sprintf("%d", id))
	}
	user.Disable()
	if err := s.userRepo.UpdateStatus(ctx, id, user.Status); err != nil {
		return err
	}
	_ = s.cache.InvalidateUserPermissions(ctx, id)
	return nil
}

func (s *authBizService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	return s.userRepo.List(ctx, page, pageSize)
}

func (s *authBizService) AssignRole(ctx context.Context, userID int64, roleCode string) error {
	role, err := s.roleRepo.QueryByCode(ctx, roleCode)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(roleCode)
	}
	if err := s.userRepo.AssignRole(ctx, userID, role.ID); err != nil {
		return err
	}
	_ = s.cache.InvalidateUserPermissions(ctx, userID)
	s.audit.LogPermissionChange(ctx, userID, "assign_role", fmt.Sprintf("user=%d role=%s", userID, roleCode))
	return nil
}

func (s *authBizService) RevokeRole(ctx context.Context, userID int64, roleCode string) error {
	if err := s.userRepo.RevokeRole(ctx, userID, roleCode); err != nil {
		return err
	}
	_ = s.cache.InvalidateUserPermissions(ctx, userID)
	s.audit.LogPermissionChange(ctx, userID, "revoke_role", fmt.Sprintf("user=%d role=%s", userID, roleCode))
	return nil
}

// ---------- 角色管理 ----------

func (s *authBizService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*domain.Role, error) {
	existing, _ := s.roleRepo.QueryByCode(ctx, req.Code)
	if existing != nil {
		return nil, domain.ErrRoleAlreadyExists(req.Code)
	}
	role := &domain.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      1,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *authBizService) GetRole(ctx context.Context, code string) (*domain.Role, error) {
	role, err := s.roleRepo.QueryByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrRoleNotFound(code)
	}
	perms, _ := s.roleRepo.QueryPermissionsByRoleID(ctx, role.ID)
	role.Permissions = perms
	return role, nil
}

func (s *authBizService) UpdateRole(ctx context.Context, code string, req *UpdateRoleRequest) error {
	role, err := s.roleRepo.QueryByCode(ctx, code)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(code)
	}
	role.Name = req.Name
	role.Description = req.Description
	return s.roleRepo.Update(ctx, role)
}

func (s *authBizService) DeleteRole(ctx context.Context, code string) error {
	return s.roleRepo.Delete(ctx, code)
}

func (s *authBizService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *authBizService) AssignPermission(ctx context.Context, roleCode string, permCode string) error {
	role, err := s.roleRepo.QueryByCode(ctx, roleCode)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(roleCode)
	}
	perm, err := s.permRepo.QueryByCode(ctx, permCode)
	if err != nil || perm == nil {
		return domain.ErrPermNotFound(permCode)
	}
	if err := s.roleRepo.AssignPermission(ctx, role.ID, perm.ID); err != nil {
		return err
	}
	s.audit.LogPermissionChange(ctx, 0, "assign_perm", fmt.Sprintf("role=%s perm=%s", roleCode, permCode))
	return nil
}

func (s *authBizService) RevokePermission(ctx context.Context, roleCode string, permCode string) error {
	role, err := s.roleRepo.QueryByCode(ctx, roleCode)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(roleCode)
	}
	if err := s.roleRepo.RevokePermission(ctx, role.ID, permCode); err != nil {
		return err
	}
	s.audit.LogPermissionChange(ctx, 0, "revoke_perm", fmt.Sprintf("role=%s perm=%s", roleCode, permCode))
	return nil
}

// ---------- 权限管理 ----------

func (s *authBizService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.Permission, error) {
	code := domain.BuildCode(req.Resource, req.Action)
	existing, _ := s.permRepo.QueryByCode(ctx, code)
	if existing != nil {
		return nil, domain.ErrPermAlreadyExists(code)
	}
	perm := &domain.Permission{
		Code:        code,
		Resource:    req.Resource,
		Action:      req.Action,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.permRepo.Create(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *authBizService) UpdatePermission(ctx context.Context, code string, req *UpdatePermissionRequest) error {
	perm, err := s.permRepo.QueryByCode(ctx, code)
	if err != nil || perm == nil {
		return domain.ErrPermNotFound(code)
	}
	perm.Name = req.Name
	perm.Description = req.Description
	return s.permRepo.Update(ctx, perm)
}

func (s *authBizService) DeletePermission(ctx context.Context, code string) error {
	return s.permRepo.Delete(ctx, code)
}

func (s *authBizService) ListPermissions(ctx context.Context, resource string) ([]*domain.Permission, error) {
	return s.permRepo.List(ctx, resource)
}

// ---------- 权限查询（PermissionChecker） ----------

func (s *authBizService) GetUserPermissionCodes(ctx context.Context, userID int64) ([]string, error) {
	return s.userRepo.QueryPermissionCodesByUserID(ctx, userID)
}

func (s *authBizService) GetUserRoleCodes(ctx context.Context, userID int64) ([]string, error) {
	roles, err := s.userRepo.QueryRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(roles))
	for i, r := range roles {
		codes[i] = r.Code
	}
	return codes, nil
}
```

- [ ] **Step 3: 验证编译**

```bash
go build ./internal/auth/biz/...
```

Expected: 零错误

- [ ] **Step 4: 提交**

```bash
git add internal/auth/biz/
git commit -m "feat(auth): add biz layer — AuthBizService with full CRUD and auth use cases"
```

---

### Task 7: Auth 应用边界层 + API 层

**Files:**
- Create: `internal/auth/service/app_service.go`
- Create: `internal/auth/api/http_handler.go`

- [ ] **Step 1: 创建 `internal/auth/service/app_service.go`**

负责 HTTP DTO ↔ 领域对象的转换，调用 biz 层。包含所有 API 的 Input/Output 结构体和 AppService 方法。

关键 DTO 示例：

```go
type LoginInput struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type LoginOutput struct {
    Token string   `json:"token"`
    Roles []string `json:"roles"`
}

type CreateUserInput struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    Nickname string `json:"nickname"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
}
// ... 更多 Input/Output 结构体
```

AppService 方法模式与 order 示例一致：接收 Input → 转换为 biz DTO → 调用 biz → 转换为 Output。

- [ ] **Step 2: 创建 `internal/auth/api/http_handler.go`**

包含 `AuthAPI` 结构体 + `RegisterRoutes(rg)` + `RegisterPublicRoutes(rg)` 方法。所有 Handler 遵循项目现有模式：ShouldBindJSON → 调用 AppService → response.Success/Error。

路由注册示例：

```go
func (a *AuthAPI) RegisterPublicRoutes(rg *gin.RouterGroup) {
    auth := rg.Group("/auth")
    auth.POST("/login", a.LoginHandler)
}

func (a *AuthAPI) RegisterRoutes(rg *gin.RouterGroup) {
    auth := rg.Group("/auth")
    auth.POST("/refresh", a.RefreshHandler)
    auth.POST("/logout", a.LogoutHandler)
    auth.GET("/me", a.MeHandler)
    auth.PUT("/password", a.ChangePasswordHandler)

    // 用户管理
    user := rg.Group("/user")
    user.GET("", a.ListUsersHandler)
    user.GET("/:id", a.GetUserHandler)
    user.POST("", a.CreateUserHandler)
    user.PUT("/:id", a.UpdateUserHandler)
    user.DELETE("/:id", a.DisableUserHandler)
    user.POST("/:id/roles", a.AssignRoleHandler)
    user.DELETE("/:id/roles/:roleCode", a.RevokeRoleHandler)

    // 角色管理
    role := rg.Group("/role")
    role.GET("", a.ListRolesHandler)
    role.GET("/:code", a.GetRoleHandler)
    role.POST("", a.CreateRoleHandler)
    role.PUT("/:code", a.UpdateRoleHandler)
    role.DELETE("/:code", a.DeleteRoleHandler)
    role.POST("/:code/permissions", a.AssignPermHandler)
    role.DELETE("/:code/permissions/:permCode", a.RevokePermHandler)

    // 权限管理
    perm := rg.Group("/permission")
    perm.GET("", a.ListPermissionsHandler)
    perm.POST("", a.CreatePermissionHandler)
    perm.PUT("/:code", a.UpdatePermissionHandler)
    perm.DELETE("/:code", a.DeletePermissionHandler)
}
```

- [ ] **Step 3: 验证编译**

```bash
go build ./internal/auth/service/... ./internal/auth/api/...
```

Expected: 零错误

- [ ] **Step 4: 提交**

```bash
git add internal/auth/service/ internal/auth/api/
git commit -m "feat(auth): add service layer (DTO) and API layer (HTTP handlers + routes)"
```

---

### Task 8: Auth IoC 装配 + PermissionChecker 实现 + 种子数据

**Files:**
- Create: `internal/auth/wire.go`

- [ ] **Step 1: 创建 `internal/auth/wire.go`**

模块内 IoC 装配函数，创建所有依赖并返回 `(*api.AuthAPI, auth.PermissionChecker)`。

关键设计：

```go
// bcryptHasher 实现 domain.PasswordHasher
type bcryptHasher struct{}
func (h *bcryptHasher) Hash(password string) (string, error) { ... }
func (h *bcryptHasher) Verify(hash, password string) bool { ... }

// permissionChecker 实现 auth.PermissionChecker
// 内部使用 AuthBizService 的 GetUserPermissionCodes / GetUserRoleCodes
type permissionChecker struct {
    biz biz.AuthBizService
}
func (c *permissionChecker) HasPermission(ctx, userID, code) (bool, error) { ... }
func (c *permissionChecker) HasAllPermissions(ctx, userID, codes) (bool, error) { ... }
func (c *permissionChecker) HasRole(ctx, userID, role) (bool, error) { ... }

// seedData 种子数据初始化
func seedData(db *gorm.DB, bizSvc biz.AuthBizService) { ... }

// Wire 装配函数
func Wire(db *gorm.DB, jwtMgr *auth.JWTManager, seedEnabled bool) (*api.AuthAPI, auth.PermissionChecker) { ... }
```

- [ ] **Step 2: 验证编译**

```bash
go build ./internal/auth/...
```

Expected: 零错误

- [ ] **Step 3: 提交**

```bash
git add internal/auth/wire.go
git commit -m "feat(auth): add module IoC wire with PermissionChecker impl and seed data"
```

---

### Task 9: 配置 + 全局装配 + 集成验证

**Files:**
- Modify: `app/config.go`
- Modify: `app/wire.go`
- Modify: `config/config.example.yaml`

- [ ] **Step 1: 在 `app/config.go` 中新增 RBACConfig**

```go
type RBACConfig struct {
    AdminRole          string `mapstructure:"admin_role"`
    PermissionCacheTTL int    `mapstructure:"permission_cache_ttl"`
    SeedEnabled        bool   `mapstructure:"seed_enabled"`
}

// 在 Config 结构体中新增
type Config struct {
    // ... 已有字段
    RBAC RBACConfig `mapstructure:"rbac"`
}
```

- [ ] **Step 2: 更新 `app/wire.go`**

集成 auth 模块 + 全局中间件：

```go
func Wire(cfg *Config) *gin.Engine {
    r := gin.New()

    // 全局中间件
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    r.Use(middleware.RequestID())
    r.Use(middleware.Logger())

    jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresIn)

    // 公开路由
    public := r.Group("/api/v1")

    // 认证路由
    v1 := r.Group("/api/v1")
    v1.Use(middleware.RequireAuth(jwtMgr))

    // auth 模块
    authAPI, permChecker := authmodule.Wire(app.GlobalDB, jwtMgr, cfg.RBAC.SeedEnabled)
    authAPI.RegisterPublicRoutes(public)
    authAPI.RegisterRoutes(v1)

    // 订单模块
    orderAPI := order.Wire()
    orderAPI.RegisterRoutes(v1)

    return r
}
```

- [ ] **Step 3: 更新 `config/config.example.yaml`**

新增 rbac 配置段。

- [ ] **Step 4: 编译验证**

```bash
go build ./...
go vet ./...
```

Expected: 零错误

- [ ] **Step 5: 提交**

```bash
git add app/config.go app/wire.go config/config.example.yaml
git commit -m "feat: integrate RBAC module into app config and Composition Root"
```
