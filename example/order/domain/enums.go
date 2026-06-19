package order

// OrderStatus 订单状态枚举
type OrderStatus int

const (
	OrderStatusDraft     OrderStatus = 0 // 草稿
	OrderStatusSubmitted OrderStatus = 1 // 已提交
	OrderStatusPaid      OrderStatus = 2 // 已支付
	OrderStatusShipped   OrderStatus = 3 // 已发货
	OrderStatusCompleted OrderStatus = 4 // 已完成
	OrderStatusCancelled OrderStatus = 5 // 已取消
)

// IsValid 校验订单状态是否为合法值
func (s OrderStatus) IsValid() bool {
	return s >= OrderStatusDraft && s <= OrderStatusCancelled
}

// IsFinal 是否为终态（终态不允许再流转）
func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusCompleted || s == OrderStatusCancelled
}

// String 返回中文描述
func (s OrderStatus) String() string {
	switch s {
	case OrderStatusDraft:
		return "草稿"
	case OrderStatusSubmitted:
		return "已提交"
	case OrderStatusPaid:
		return "已支付"
	case OrderStatusShipped:
		return "已发货"
	case OrderStatusCompleted:
		return "已完成"
	case OrderStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// PaymentType 支付方式枚举
type PaymentType int

const (
	PaymentTypeAlipay PaymentType = 1 // 支付宝
	PaymentTypeWechat PaymentType = 2 // 微信支付
	PaymentTypeBank   PaymentType = 3 // 银行转账
)

// IsValid 校验支付方式是否合法
func (p PaymentType) IsValid() bool {
	return p >= PaymentTypeAlipay && p <= PaymentTypeBank
}

// OrderType 订单类型枚举
type OrderType int

const (
	OrderTypeNormal   OrderType = 1 // 普通订单
	OrderTypePreorder OrderType = 2 // 预售订单
	OrderTypeGift     OrderType = 3 // 礼品订单
)

// IsValid 校验订单类型是否合法
func (t OrderType) IsValid() bool {
	return t >= OrderTypeNormal && t <= OrderTypeGift
}

// RequiresPrepay 是否需要预付款
func (t OrderType) RequiresPrepay() bool {
	return t == OrderTypePreorder
}
