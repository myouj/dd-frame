package model

import (
	"context"

	orderdomain "github.com/example/dd-frame/example/order/domain"
)

// OrderRepo 订单仓储接口
//
// 使用领域对象，不暴露 DB 模型。方法命名体现领域语义。
type OrderRepo interface {
	Create(ctx context.Context, order *orderdomain.Order) error
	QueryByID(ctx context.Context, id int64) (*orderdomain.Order, error)
	QueryByOrderNo(ctx context.Context, orderNo string) (*orderdomain.Order, error)
	UpdateStatus(ctx context.Context, id int64, status orderdomain.OrderStatus) error
	QueryByCustomerID(ctx context.Context, customerID int64, page, pageSize int) ([]*orderdomain.Order, int64, error)
	QueryByCompanyIDWithPage(ctx context.Context, companyID int64, page, pageSize int, orderBy, sort string) ([]*orderdomain.Order, int64, error)
}
