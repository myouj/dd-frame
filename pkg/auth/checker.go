package auth

import "context"

// PermissionChecker 权限校验器接口
//
// 由 internal/auth 模块实现，middleware 和业务代码通过此接口判断权限。
type PermissionChecker interface {
	// HasPermission 判断用户是否持有指定权限码
	HasPermission(ctx context.Context, userID int64, code string) (bool, error)

	// HasAllPermissions 判断用户是否持有所有指定权限码
	HasAllPermissions(ctx context.Context, userID int64, codes []string) (bool, error)

	// HasRole 判断用户是否持有指定角色
	HasRole(ctx context.Context, userID int64, role string) (bool, error)
}
