package service

import (
	"context"

	orderbiz "github.com/example/dd-frame/example/order/biz"
	orderdomain "github.com/example/dd-frame/example/order/domain"
)

// OrderAppService 订单应用边界服务
//
// 负责 proto/HTTP DTO ↔ domain 对象的转换，调用 biz 层编排。
type OrderAppService struct {
	usecase orderbiz.OrderService
}

// NewOrderAppService 创建订单应用边界服务
func NewOrderAppService(usecase orderbiz.OrderService) *OrderAppService {
	return &OrderAppService{usecase: usecase}
}

// CreateOrderInput 创建订单 HTTP 入参
type CreateOrderInput struct {
	CustomerID int64            `json:"customerId"`
	Items      []OrderItemInput `json:"items"`
}

// OrderItemInput 订单项入参
type OrderItemInput struct {
	ProductID int64 `json:"productId"`
	Quantity  int   `json:"quantity"`
	UnitPrice int64 `json:"unitPrice"` // 分
}

// CreateOrderOutput 创建订单 HTTP 出参
type CreateOrderOutput struct {
	OrderNo  string `json:"orderNo"`
	OrderID  int64  `json:"orderId"`
	Status   string `json:"status"`
	TotalAmt string `json:"totalAmount"`
}

// CreateOrder 创建订单
func (s *OrderAppService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error) {
	// 1. HTTP DTO → 业务 DTO
	bizReq := &orderbiz.CreateOrderRequest{
		CustomerID: input.CustomerID,
		OrderType:  orderdomain.OrderTypeNormal,
	}
	for _, item := range input.Items {
		bizReq.Items = append(bizReq.Items, orderbiz.CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: orderdomain.NewMoney(item.UnitPrice),
		})
	}

	// 2. 调用 biz 层
	order, err := s.usecase.CreateOrder(ctx, 0, 0, bizReq)
	if err != nil {
		return nil, err
	}

	// 3. 领域对象 → HTTP DTO
	return &CreateOrderOutput{
		OrderNo:  order.OrderNo,
		OrderID:  order.ID,
		Status:   order.Status.String(),
		TotalAmt: order.Amount.ToYuanString(),
	}, nil
}

// SubmitOrder 提交订单
func (s *OrderAppService) SubmitOrder(ctx context.Context, orderNo string) error {
	return s.usecase.SubmitOrder(ctx, 0, orderNo)
}

// CancelOrder 取消订单
func (s *OrderAppService) CancelOrder(ctx context.Context, orderNo string, reason string) error {
	return s.usecase.CancelOrder(ctx, 0, orderNo, reason)
}
