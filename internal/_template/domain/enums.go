package domain

// StatusEnum 状态枚举示例（替换为实际枚举）
//
// 每个枚举必须包含 IsValid() 和 String() 方法。
type StatusEnum int

const (
	StatusA StatusEnum = 0 // 状态A
	StatusB StatusEnum = 1 // 状态B
)

// IsValid 校验枚举值是否合法
func (s StatusEnum) IsValid() bool {
	return s >= StatusA && s <= StatusB
}

// String 返回描述
func (s StatusEnum) String() string {
	switch s {
	case StatusA:
		return "状态A"
	case StatusB:
		return "状态B"
	default:
		return "未知"
	}
}
