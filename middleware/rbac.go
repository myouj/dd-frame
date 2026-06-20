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

		// 超级管理员直接放行（查询数据库，不依赖 JWT）
		isAdmin, err := checker.HasRole(c.Request.Context(), user.UserID, "admin")
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 50000, "role check failed")
			c.Abort()
			return
		}
		if isAdmin {
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

		isAdmin, err := checker.HasRole(c.Request.Context(), user.UserID, "admin")
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 50000, "role check failed")
			c.Abort()
			return
		}
		if isAdmin {
			c.Next()
			return
		}

		ok, err = checker.HasAllPermissions(c.Request.Context(), user.UserID, codes)
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

		isAdmin, err := checker.HasRole(c.Request.Context(), user.UserID, "admin")
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 50000, "role check failed")
			c.Abort()
			return
		}
		if isAdmin {
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
