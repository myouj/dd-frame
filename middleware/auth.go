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
			// Token 无效也放行，不注入 AuthUser
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
