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
