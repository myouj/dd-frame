package model

import (
	"context"

	"github.com/example/dd-frame/internal/auth/domain"
)

// UserRepo 用户仓储接口
type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	QueryByID(ctx context.Context, id int64) (*domain.User, error)
	QueryByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateStatus(ctx context.Context, id int64, status domain.UserStatus) error
	List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
	AssignRole(ctx context.Context, userID int64, roleID int64) error
	RevokeRole(ctx context.Context, userID int64, roleCode string) error
	QueryRolesByUserID(ctx context.Context, userID int64) ([]domain.Role, error)
	QueryPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error)
}

// RoleRepo 角色仓储接口
type RoleRepo interface {
	Create(ctx context.Context, role *domain.Role) error
	QueryByCode(ctx context.Context, code string) (*domain.Role, error)
	QueryByID(ctx context.Context, id int64) (*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, code string) error
	List(ctx context.Context) ([]*domain.Role, error)
	AssignPermission(ctx context.Context, roleID int64, permissionID int64) error
	RevokePermission(ctx context.Context, roleID int64, permCode string) error
	QueryPermissionsByRoleID(ctx context.Context, roleID int64) ([]domain.Permission, error)
}

// PermissionRepo 权限仓储接口
type PermissionRepo interface {
	Create(ctx context.Context, perm *domain.Permission) error
	QueryByCode(ctx context.Context, code string) (*domain.Permission, error)
	Update(ctx context.Context, perm *domain.Permission) error
	Delete(ctx context.Context, code string) error
	List(ctx context.Context, resource string) ([]*domain.Permission, error)
}
