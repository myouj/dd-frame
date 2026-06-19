package order

import "fmt"

// 领域错误 reason 常量
const (
	ReasonOrderNotFound       = "ORDER_NOT_FOUND"
	ReasonOrderStatusInvalid  = "ORDER_STATUS_INVALID"
	ReasonOrderEmpty          = "ORDER_EMPTY"
	ReasonOrderCannotCancel   = "ORDER_CANNOT_CANCEL"
	ReasonOrderAmountMismatch = "ORDER_AMOUNT_MISMATCH"
)

// ErrOrderNotFound 订单不存在
func ErrOrderNotFound(orderNo string) error {
	return fmt.Errorf("[%s] order not found: %s", ReasonOrderNotFound, orderNo)
}

// ErrOrderStatusInvalid 订单状态不合法
func ErrOrderStatusInvalid(current OrderStatus, action string) error {
	return fmt.Errorf("[%s] cannot %s order in status %s", ReasonOrderStatusInvalid, action, current)
}

// ErrOrderEmpty 订单无商品
func ErrOrderEmpty() error {
	return fmt.Errorf("[%s] order must have at least one item", ReasonOrderEmpty)
}

// ErrOrderCannotCancel 订单不可取消
func ErrOrderCannotCancel(current OrderStatus) error {
	return fmt.Errorf("[%s] cannot cancel order in status %s", ReasonOrderCannotCancel, current)
}

// ErrOrderAmountMismatch 金额不匹配
func ErrOrderAmountMismatch(expected, actual Money) error {
	return fmt.Errorf("[%s] amount mismatch: expected %s, got %s",
		ReasonOrderAmountMismatch, expected.ToYuanString(), actual.ToYuanString())
}
