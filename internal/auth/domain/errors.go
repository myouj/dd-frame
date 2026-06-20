package domain

import "fmt"

// 领域错误 reason 常量
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

// ErrUserNotFound 用户不存在
func ErrUserNotFound(identifier string) error {
	return fmt.Errorf("[%s] user not found: %s", ReasonUserNotFound, identifier)
}

// ErrUserAlreadyExists 用户名已存在
func ErrUserAlreadyExists(username string) error {
	return fmt.Errorf("[%s] user already exists: %s", ReasonUserAlreadyExists, username)
}

// ErrUserDisabled 用户已禁用
func ErrUserDisabled() error {
	return fmt.Errorf("[%s] user is disabled", ReasonUserDisabled)
}

// ErrInvalidCredentials 用户名或密码错误
func ErrInvalidCredentials() error {
	return fmt.Errorf("[%s] invalid username or password", ReasonInvalidCredentials)
}

// ErrRoleNotFound 角色不存在
func ErrRoleNotFound(code string) error {
	return fmt.Errorf("[%s] role not found: %s", ReasonRoleNotFound, code)
}

// ErrRoleAlreadyExists 角色已存在
func ErrRoleAlreadyExists(code string) error {
	return fmt.Errorf("[%s] role already exists: %s", ReasonRoleAlreadyExists, code)
}

// ErrPermNotFound 权限不存在
func ErrPermNotFound(code string) error {
	return fmt.Errorf("[%s] permission not found: %s", ReasonPermNotFound, code)
}

// ErrPermAlreadyExists 权限已存在
func ErrPermAlreadyExists(code string) error {
	return fmt.Errorf("[%s] permission already exists: %s", ReasonPermAlreadyExists, code)
}
