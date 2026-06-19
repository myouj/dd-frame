package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperr "github.com/example/dd-frame/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: msg,
	})
}

// FromError 从 AppError 生成响应
func FromError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperr.AppError); ok {
		httpStatus := mapCodeToHTTP(appErr.Code)
		c.JSON(httpStatus, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    50000,
		Message: "internal server error",
	})
}

// mapCodeToHTTP 业务错误码 → HTTP 状态码
func mapCodeToHTTP(code int) int {
	switch {
	case code >= 40100 && code < 40200:
		return http.StatusUnauthorized
	case code >= 40300 && code < 40400:
		return http.StatusForbidden
	case code >= 40400 && code < 40500:
		return http.StatusNotFound
	case code >= 40900 && code < 41000:
		return http.StatusConflict
	case code >= 40000 && code < 50000:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
