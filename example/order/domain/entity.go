package order

import "time"

// Order 订单聚合根
//
// 聚合根是外部访问聚合内部对象的唯一入口。
type Order struct {
	ID         int64       // 主键 ID
	OrderNo    string      // 订单编号（业务唯一键）
	CustomerID int64       // 客户 ID
	Status     OrderStatus // 订单状态
	Amount     Money       // 订单金额
	Items      []OrderItem // 订单项列表
	CreatedAt  time.Time   // 创建时间
	UpdatedAt  time.Time   // 更新时间
}

// Submit 提交订单
func (o *Order) Submit() error {
	if o.Status != OrderStatusDraft {
		return ErrOrderStatusInvalid(o.Status, "submit")
	}
	if len(o.Items) == 0 {
		return ErrOrderEmpty()
	}
	o.Status = OrderStatusSubmitted
	return nil
}

// Cancel 取消订单
func (o *Order) Cancel(reason string) error {
	if o.Status == OrderStatusShipped || o.Status == OrderStatusCompleted {
		return ErrOrderCannotCancel(o.Status)
	}
	o.Status = OrderStatusCancelled
	return nil
}

// CalculateTotal 计算订单总金额
func (o *Order) CalculateTotal() Money {
	var total int64
	for _, item := range o.Items {
		total += int64(item.Subtotal())
	}
	o.Amount = NewMoney(total)
	return o.Amount
}

// OrderItem 订单项实体
type OrderItem struct {
	ID        int64 // 主键 ID
	ProductID int64 // 商品 ID
	Quantity  int   // 数量
	UnitPrice Money // 单价
}

// Subtotal 计算订单项小计金额
func (i *OrderItem) Subtotal() Money {
	return NewMoney(int64(i.Quantity) * int64(i.UnitPrice))
}
