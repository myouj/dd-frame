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
