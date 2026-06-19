package domain

import "fmt"

// 领域错误 reason 常量
const (
	ReasonNotFound      = "ENTITY_NOT_FOUND"
	ReasonStatusInvalid = "ENTITY_STATUS_INVALID"
)

// ErrNotFound 实体不存在
func ErrNotFound(id string) error {
	return fmt.Errorf("[%s] entity not found: %s", ReasonNotFound, id)
}

// ErrStatusInvalid 状态不合法
func ErrStatusInvalid(current int, action string) error {
	return fmt.Errorf("[%s] cannot %s in status %d", ReasonStatusInvalid, action, current)
}
