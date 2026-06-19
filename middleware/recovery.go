package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/log"
	"github.com/example/dd-frame/pkg/response"
)

// Recovery Panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", "err", r, "path", c.Request.URL.Path)
				response.Error(c, http.StatusInternalServerError, 50000, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
