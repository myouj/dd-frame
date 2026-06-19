package order

// Money 金额值对象（Value Object）
//
// 不可变，通过属性值判等。使用毫分作为内部统一单位避免浮点精度问题。
type Money int64

// NewMoney 从分（最小单位）创建金额
func NewMoney(cents int64) Money {
	return Money(cents)
}

// NewMoneyFromYuan 从元创建金额（内部转为分）
func NewMoneyFromYuan(yuan float64) Money {
	return Money(yuan * 100)
}

// ToYuanString 转换为元的字符串表示
func (m Money) ToYuanString() string {
	yuan := float64(m) / 100.0
	return formatMoney(yuan)
}

// IsZero 是否为零
func (m Money) IsZero() bool {
	return m == 0
}

// Add 金额相加（值对象操作返回新值对象）
func (m Money) Add(other Money) Money {
	return Money(int64(m) + int64(other))
}

// Multiply 金额乘数量
func (m Money) Multiply(quantity int) Money {
	return Money(int64(m) * int64(quantity))
}

// formatMoney 格式化金额
func formatMoney(yuan float64) string {
	return "" // 实际项目中使用 shopspring/decimal 等库
}

// OptionalMoney 可选金额值对象（区分"有值"和"未知"）
type OptionalMoney struct {
	Value Money // 金额值
	Valid bool  // true=有值, false=未知/未提供
}

// NewOptionalMoney 创建有效的可选金额
func NewOptionalMoney(v Money) OptionalMoney {
	return OptionalMoney{Value: v, Valid: true}
}

// Get 获取金额值及是否有效标志
func (o OptionalMoney) Get() (Money, bool) {
	return o.Value, o.Valid
}
