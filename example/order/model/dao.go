package model

import (
	"context"
	"time"

	orderdomain "github.com/example/dd-frame/example/order/domain"
)

// 编译期校验：确保 DAO 实现了 OrderRepo 接口
var _ OrderRepo = (*OrderDAO)(nil)

// OrderModel DB 表模型
type OrderModel struct {
	ID          int64     `gorm:"primary_key;auto_increment"`
	OrderNo     string    `gorm:"unique_index;size:64"`
	CustomerID  int64     `gorm:"index"`
	Status      int       `gorm:"default:0"`
	AmountCents int64     `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName 表名
func (OrderModel) TableName() string {
	return "t_order"
}

// OrderDAO 订单仓储 GORM 实现
type OrderDAO struct {
	// db *gorm.DB // 实际项目中注入 GORM 实例
}

// NewOrderDAO 创建订单 DAO
func NewOrderDAO() *OrderDAO {
	return &OrderDAO{}
}

// Create 创建订单
func (d *OrderDAO) Create(ctx context.Context, order *orderdomain.Order) error {
	model := orderToModel(order)
	// d.db.WithContext(ctx).Create(&model)
	order.ID = model.ID
	return nil
}

// QueryByID 根据主键查询
func (d *OrderDAO) QueryByID(_ context.Context, _ int64) (*orderdomain.Order, error) {
	return nil, nil
}

// QueryByOrderNo 根据订单号查询
func (d *OrderDAO) QueryByOrderNo(_ context.Context, _ string) (*orderdomain.Order, error) {
	return nil, nil
}

// UpdateStatus 更新订单状态
func (d *OrderDAO) UpdateStatus(_ context.Context, _ int64, _ orderdomain.OrderStatus) error {
	return nil
}

// QueryByCustomerID 根据客户ID分页查询
func (d *OrderDAO) QueryByCustomerID(_ context.Context, _ int64, _, _ int) ([]*orderdomain.Order, int64, error) {
	return nil, 0, nil
}

// QueryByCompanyIDWithPage 公司隔离分页查询
func (d *OrderDAO) QueryByCompanyIDWithPage(_ context.Context, _ int64, _, _ int, _, _ string) ([]*orderdomain.Order, int64, error) {
	return nil, 0, nil
}

// ==================== Converter（DB 模型 ↔ 领域对象） ====================

func orderToModel(o *orderdomain.Order) *OrderModel {
	return &OrderModel{
		ID:          o.ID,
		OrderNo:     o.OrderNo,
		CustomerID:  o.CustomerID,
		Status:      int(o.Status),
		AmountCents: int64(o.Amount),
	}
}

func modelToOrder(m *OrderModel) *orderdomain.Order {
	if m == nil {
		return nil
	}
	return &orderdomain.Order{
		ID:         m.ID,
		OrderNo:    m.OrderNo,
		CustomerID: m.CustomerID,
		Status:     orderdomain.OrderStatus(m.Status),
		Amount:     orderdomain.NewMoney(m.AmountCents),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
