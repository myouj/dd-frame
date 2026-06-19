package domain

// Money 金额值对象示例（替换为实际值对象）
//
// 值对象特征：不可变，通过属性值判等，操作方法返回新实例。
type Money int64

// NewMoney 创建金额
func NewMoney(cents int64) Money {
	return Money(cents)
}

// Add 金额相加（返回新值对象）
func (m Money) Add(other Money) Money {
	return Money(int64(m) + int64(other))
}
