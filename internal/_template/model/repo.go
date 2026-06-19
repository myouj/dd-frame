package model
package model

import (
	"context"

	"github.com/example/dd-frame/internal/_template/domain"
)

// EntityRepo 仓储接口
//
// 使用领域对象，不暴露 DB 模型。方法命名体现领域语义。
type EntityRepo interface {
	Create(ctx context.Context, entity *domain.EntityName) error
	QueryByID(ctx context.Context, id int64) (*domain.EntityName, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
}
