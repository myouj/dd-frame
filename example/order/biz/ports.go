package biz

import (
	"context"

	orderdomain "github.com/example/dd-frame/example/order/domain"
)

// ExternalPaymentClient 定义外部支付平台调用能力
type ExternalPaymentClient interface {
	CreatePayment(ctx context.Context, orderNo string, amount orderdomain.Money, paymentType orderdomain.PaymentType) (string, error)
	QueryPayment(ctx context.Context, paymentID string) (*PaymentResult, error)
}

// PaymentResult 支付查询结果
type PaymentResult struct {
	PaymentID string
	OrderNo   string
	Amount    orderdomain.Money
	Status    string // paid / pending / failed
}

// ExternalInventoryClient 定义外部库存系统调用能力
type ExternalInventoryClient interface {
	DeductStock(ctx context.Context, productID int64, quantity int) error
	QueryStock(ctx context.Context, productID int64) (int, error)
}

// OrderNotificationPort 定义订单通知能力
type OrderNotificationPort interface {
	NotifyOrderCreated(ctx context.Context, order *orderdomain.Order) error
	NotifyOrderCancelled(ctx context.Context, order *orderdomain.Order, reason string) error
}
