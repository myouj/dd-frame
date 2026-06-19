package domain
// Package domain 包含聚合根、实体定义。
// 领域层零外部依赖，仅使用标准库。
//
// 使用方式：复制 _template/ 为 internal/{module}/，替换 EntityName。
package domain

import "time"

// EntityName 聚合根（替换为实际业务名称，如 Order、Payment）
//
// 聚合根是外部访问聚合内部对象的唯一入口。
// 所有对聚合内部状态的修改必须通过聚合根方法完成。
type EntityName struct {
	ID        int64     // 主键 ID
	Status    int       // 业务状态（使用枚举常量）
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// EntityNameItem 聚合内实体（替换为实际名称，如 OrderItem）
type EntityNameItem struct {
	ID int64 // 主键 ID
}

// DoAction 聚合根行为方法（替换为实际业务操作）
//
// 业务规则内聚在聚合根方法中，不要在 biz 层写业务逻辑。
func (e *EntityName) DoAction() error {
	// 1. 业务校验
	// if e.Status != ValidStatus { return ErrInvalidStatus }
	// 2. 状态流转
	// e.Status = NextStatus
	return nil
}
