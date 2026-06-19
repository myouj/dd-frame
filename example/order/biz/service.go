package biz

import (
	"context"
	"fmt"

	orderdomain "github.com/example/dd-frame/example/order/domain"
	ordermodel "github.com/example/dd-frame/example/order/model"
)

// OrderService 定义订单应用服务能力
type OrderService interface {
	CreateOrder(ctx context.Context, companyID int64, userID uint, req *CreateOrderRequest) (*orderdomain.Order, error)
	SubmitOrder(ctx context.Context, companyID int64, orderNo string) error
	CancelOrder(ctx context.Context, companyID int64, orderNo string, reason string) error
}

// CreateOrderRequest 创建订单请求 DTO
type CreateOrderRequest struct {
	CustomerID int64
	OrderType  orderdomain.OrderType
	Items      []CreateOrderItemRequest
}

// CreateOrderItemRequest 创建订单项请求
type CreateOrderItemRequest struct {
	ProductID int64
	Quantity  int
	UnitPrice orderdomain.Money
}

// orderService 订单应用服务实现
type orderService struct {
	orderRepo       ordermodel.OrderRepo
	paymentClient   ExternalPaymentClient
	inventoryClient ExternalInventoryClient
	notification    OrderNotificationPort
	numberGen       orderdomain.OrderNumberGenerator
}

// NewOrderService 创建订单应用服务
func NewOrderService(
	orderRepo ordermodel.OrderRepo,
	paymentClient ExternalPaymentClient,
	inventoryClient ExternalInventoryClient,
	notification OrderNotificationPort,
	numberGen orderdomain.OrderNumberGenerator,
) OrderService {
	return &orderService{
		orderRepo:       orderRepo,
		paymentClient:   paymentClient,
		inventoryClient: inventoryClient,
		notification:    notification,
		numberGen:       numberGen,
	}
}

// CreateOrder 创建订单用例编排
func (s *orderService) CreateOrder(ctx context.Context, companyID int64, userID uint, req *CreateOrderRequest) (*orderdomain.Order, error) {
	// 1. 生成订单号
	orderNo, err := s.numberGen.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate order number failed: %w", err)
	}

	// 2. 构建聚合根
	order := &orderdomain.Order{
		OrderNo:    orderNo,
		CustomerID: req.CustomerID,
		Status:     orderdomain.OrderStatusDraft,
	}
	for _, item := range req.Items {
		order.Items = append(order.Items, orderdomain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	// 3. 计算总金额（领域逻辑）
	order.CalculateTotal()

	// 4. 校验库存（调用外部端口）
	for _, item := range order.Items {
		stock, err := s.inventoryClient.QueryStock(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("query stock failed for product %d: %w", item.ProductID, err)
		}
		if stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %d: available %d, required %d", item.ProductID, stock, item.Quantity)
		}
	}

	// 5. 持久化
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("save order failed: %w", err)
	}

	// 6. 异步通知
	_ = s.notification.NotifyOrderCreated(ctx, order)

	return order, nil
}

// SubmitOrder 提交订单用例编排
func (s *orderService) SubmitOrder(ctx context.Context, companyID int64, orderNo string) error {
	order, err := s.orderRepo.QueryByOrderNo(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("query order failed: %w", err)
	}
	if order == nil {
		return orderdomain.ErrOrderNotFound(orderNo)
	}

	if err := order.Submit(); err != nil {
		return err
	}

	for _, item := range order.Items {
		if err := s.inventoryClient.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("deduct stock failed: %w", err)
		}
	}

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return fmt.Errorf("update order status failed: %w", err)
	}

	return nil
}

// CancelOrder 取消订单用例编排
func (s *orderService) CancelOrder(ctx context.Context, companyID int64, orderNo string, reason string) error {
	order, err := s.orderRepo.QueryByOrderNo(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("query order failed: %w", err)
	}
	if order == nil {
		return orderdomain.ErrOrderNotFound(orderNo)
	}

	if err := order.Cancel(reason); err != nil {
		return err
	}

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return fmt.Errorf("update order status failed: %w", err)
	}

	_ = s.notification.NotifyOrderCancelled(ctx, order, reason)

	return nil
}
