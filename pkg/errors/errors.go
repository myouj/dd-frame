package errors

import (
	"fmt"
	"runtime"
)

// AppError 应用层统一错误结构
type AppError struct {
	Code    int    // 业务错误码
	Message string // 用户可见的错误消息
	Reason  string // 内部原因
	Stack   string // 调用栈
}

func (e *AppError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Reason)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建应用错误
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WithReason 附加内部原因
func (e *AppError) WithReason(reason string) *AppError {
	e.Reason = reason
	return e
}

// WithStack 附加调用栈
func (e *AppError) WithStack() *AppError {
	_, file, line, _ := runtime.Caller(1)
	e.Stack = fmt.Sprintf("%s:%d", file, line)
	return e
}

// Wrap 包装已有错误
func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Reason:  err.Error(),
	}
}
