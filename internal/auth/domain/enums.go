package domain

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusDisabled UserStatus = 0 // 禁用
	UserStatusActive   UserStatus = 1 // 启用
)

// IsValid 校验状态值
func (s UserStatus) IsValid() bool {
	return s == UserStatusDisabled || s == UserStatusActive
}

// String 返回描述
func (s UserStatus) String() string {
	switch s {
	case UserStatusActive:
		return "启用"
	case UserStatusDisabled:
		return "禁用"
	default:
		return "未知"
	}
}
