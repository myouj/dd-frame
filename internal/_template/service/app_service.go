package service

import (
	"context"

	"github.com/example/dd-frame/internal/_template/biz"
)

// EntityAppService 应用边界层（DTO 转换）
//
// 负责 HTTP/gRPC DTO ↔ 业务 DTO 的转换，调用 biz 层。
type EntityAppService struct {
	usecase biz.EntityService
}

// NewEntityAppService 创建应用边界服务
func NewEntityAppService(usecase biz.EntityService) *EntityAppService {
	return &EntityAppService{usecase: usecase}
}

// CreateInput HTTP 入参 DTO
type CreateInput struct {
	// HTTP 请求字段
}

// CreateOutput HTTP 出参 DTO
type CreateOutput struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// CreateEntity 创建实体（HTTP 入口）
func (s *EntityAppService) CreateEntity(ctx context.Context, input *CreateInput) (*CreateOutput, error) {
	// 1. HTTP DTO → 业务 DTO
	bizReq := &biz.CreateRequest{}

	// 2. 调用 biz 层
	entity, err := s.usecase.CreateEntity(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	// 3. 领域对象 → HTTP DTO
	return &CreateOutput{
		ID: entity.ID,
	}, nil
}
