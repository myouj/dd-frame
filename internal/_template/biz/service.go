package biz

import (
	"context"
	"fmt"

	"github.com/example/dd-frame/internal/_template/domain"
	"github.com/example/dd-frame/internal/_template/model"
)

// EntityService 应用服务接口（替换为实际名称）
type EntityService interface {
	CreateEntity(ctx context.Context, req *CreateRequest) (*domain.EntityName, error)
}

// CreateRequest 创建请求 DTO
type CreateRequest struct {
	// 业务字段
}

// entityService 应用服务实现
type entityService struct {
	repo model.EntityRepo
	// 其他端口接口...
}

// NewEntityService 创建应用服务
func NewEntityService(repo model.EntityRepo) EntityService {
	return &entityService{repo: repo}
}

// CreateEntity 创建实体用例编排
func (s *entityService) CreateEntity(ctx context.Context, req *CreateRequest) (*domain.EntityName, error) {
	// 1. 构建聚合根
	entity := &domain.EntityName{}

	// 2. 调用聚合根方法（业务规则在领域层）
	if err := entity.DoAction(); err != nil {
		return nil, err
	}

	// 3. 持久化
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("save entity failed: %w", err)
	}

	return entity, nil
}
