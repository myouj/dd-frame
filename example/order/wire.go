package order

import (
	"context"
	"fmt"
	"time"

	"github.com/example/dd-frame/example/order/api"
	"github.com/example/dd-frame/example/order/biz"
	orderdomain "github.com/example/dd-frame/example/order/domain"
	"github.com/example/dd-frame/example/order/model"
	"github.com/example/dd-frame/example/order/service"
)

// Wire 订单模块 IoC 装配
//
// 创建本模块所有依赖并返回 OrderAPI handler。
// 在 app/wire.go 中调用此函数注册模块路由。
func Wire() *api.OrderAPI {
	// 1. 数据层
	repo := model.NewOrderDAO()

	// 2. 端口适配器（示例使用 stub 实现）
	paymentClient := &stubPaymentClient{}
	inventoryClient := &stubInventoryClient{}
	notification := &stubNotification{}
	numberGen := &stubNumberGen{}

	// 3. 业务编排层
	svc := biz.NewOrderService(repo, paymentClient, inventoryClient, notification, numberGen)

	// 4. 应用边界层
	appSvc := service.NewOrderAppService(svc)

	// 5. API 层
	return api.NewOrderAPI(appSvc)
}

// ==================== Stub 实现（仅用于示例编译通过） ====================

type stubPaymentClient struct{}

func (s *stubPaymentClient) CreatePayment(_ context.Context, orderNo string, _ orderdomain.Money, _ orderdomain.PaymentType) (string, error) {
	return fmt.Sprintf("PAY-%s-%d", orderNo, time.Now().UnixNano()), nil
}

func (s *stubPaymentClient) QueryPayment(_ context.Context, paymentID string) (*biz.PaymentResult, error) {
	return &biz.PaymentResult{PaymentID: paymentID, Status: "paid"}, nil
}

type stubInventoryClient struct{}

func (s *stubInventoryClient) DeductStock(_ context.Context, _ int64, _ int) error { return nil }
func (s *stubInventoryClient) QueryStock(_ context.Context, _ int64) (int, error)  { return 999, nil }

type stubNotification struct{}

func (s *stubNotification) NotifyOrderCreated(_ context.Context, _ *orderdomain.Order) error { return nil }
func (s *stubNotification) NotifyOrderCancelled(_ context.Context, _ *orderdomain.Order, _ string) error {
	return nil
}

type stubNumberGen struct{}

func (s *stubNumberGen) Generate() (string, error) {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano()), nil
}
