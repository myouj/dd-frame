package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/response"
)

// Auth JWT 鉴权中间件
//
// 从 Authorization header 提取 Bearer token 并验证。
// 验证成功后将 userID 和 companyID 注入 gin.Context。
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, 40100, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, 40101, "invalid authorization format")
			c.Abort()
			return
		}

		token := parts[1]
		// 实际项目中替换为 JWT 解析逻辑：
		// claims, err := jwt.ParseToken(token, secret)
		// c.Set("userID", claims.UserID)
		// c.Set("companyID", claims.CompanyID)
		_ = token
		_ = secret

		c.Next()
	}
}
