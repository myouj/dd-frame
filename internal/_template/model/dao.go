package model

import (
	"context"
	"time"

	"github.com/example/dd-frame/internal/_template/domain"
)

// 编译期校验：确保 DAO 实现了 Repo 接口
var _ EntityRepo = (*EntityDAO)(nil)

// EntityModel DB 表模型
type EntityModel struct {
	ID        int64     `gorm:"primary_key;auto_increment"`
	Status    int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName 表名
func (EntityModel) TableName() string {
	return "t_entity" // 替换为实际表名
}

// EntityDAO 仓储 GORM 实现
type EntityDAO struct {
	// db *gorm.DB // 注入 GORM 实例
}

// NewEntityDAO 创建 DAO
func NewEntityDAO() *EntityDAO {
	return &EntityDAO{}
}

func (d *EntityDAO) Create(_ context.Context, entity *domain.EntityName) error {
	// model := entityToModel(entity)
	// d.db.Create(&model)
	// entity.ID = model.ID
	return nil
}

func (d *EntityDAO) QueryByID(_ context.Context, _ int64) (*domain.EntityName, error) {
	return nil, nil
}

func (d *EntityDAO) UpdateStatus(_ context.Context, _ int64, _ int) error {
	return nil
}

// entityToModel 领域对象 → DB 模型
func entityToModel(e *domain.EntityName) *EntityModel {
	return &EntityModel{ID: e.ID, Status: e.Status}
}

// modelToEntity DB 模型 → 领域对象
func modelToEntity(m *EntityModel) *domain.EntityName {
	if m == nil {
		return nil
	}
	return &domain.EntityName{ID: m.ID, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}
