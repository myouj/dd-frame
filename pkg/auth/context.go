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
