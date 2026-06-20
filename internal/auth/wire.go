package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/example/dd-frame/internal/auth/api"
	"github.com/example/dd-frame/internal/auth/biz"
	authmodel "github.com/example/dd-frame/internal/auth/model"
	"github.com/example/dd-frame/internal/auth/service"
	"github.com/example/dd-frame/pkg/auth"
	applog "github.com/example/dd-frame/pkg/log"
)

// ==================== bcryptHasher ====================

// bcryptHasher 实现 domain.PasswordHasher
type bcryptHasher struct{}

func (h *bcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *bcryptHasher) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ==================== permissionChecker ====================

// permissionChecker 实现 auth.PermissionChecker
type permissionChecker struct {
	biz biz.AuthBizService
}

func (c *permissionChecker) HasPermission(ctx context.Context, userID int64, code string) (bool, error) {
	codes, err := c.biz.GetUserPermissionCodes(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, pc := range codes {
		if pc == code {
			return true, nil
		}
	}
	return false, nil
}

func (c *permissionChecker) HasAllPermissions(ctx context.Context, userID int64, codes []string) (bool, error) {
	userCodes, err := c.biz.GetUserPermissionCodes(ctx, userID)
	if err != nil {
		return false, err
	}
	codeSet := make(map[string]struct{}, len(userCodes))
	for _, c := range userCodes {
		codeSet[c] = struct{}{}
	}
	for _, code := range codes {
		if _, ok := codeSet[code]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func (c *permissionChecker) HasRole(ctx context.Context, userID int64, role string) (bool, error) {
	roles, err := c.biz.GetUserRoleCodes(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

// ==================== Seed Data ====================

// seedData 初始化种子数据（admin 角色 + 基础权限）
func seedData(bizSvc biz.AuthBizService) {
	ctx := context.Background()

	// 创建 admin 角色
	_, err := bizSvc.CreateRole(ctx, &biz.CreateRoleRequest{
		Code:        "admin",
		Name:        "超级管理员",
		Description: "系统超级管理员，拥有所有权限",
	})
	if err != nil {
		applog.Info("seed: admin role already exists or creation failed", "error", err)
	}

	// 创建基础权限
	basePerms := []biz.CreatePermissionRequest{
		{Resource: "user", Action: "create", Name: "创建用户"},
		{Resource: "user", Action: "read", Name: "查看用户"},
		{Resource: "user", Action: "update", Name: "更新用户"},
		{Resource: "user", Action: "delete", Name: "删除用户"},
		{Resource: "role", Action: "create", Name: "创建角色"},
		{Resource: "role", Action: "read", Name: "查看角色"},
		{Resource: "role", Action: "update", Name: "更新角色"},
		{Resource: "role", Action: "delete", Name: "删除角色"},
		{Resource: "permission", Action: "create", Name: "创建权限"},
		{Resource: "permission", Action: "read", Name: "查看权限"},
		{Resource: "permission", Action: "update", Name: "更新权限"},
		{Resource: "permission", Action: "delete", Name: "删除权限"},
	}

	for _, p := range basePerms {
		perm, err := bizSvc.CreatePermission(ctx, &p)
		if err != nil {
			applog.Info("seed: permission already exists or creation failed", "code", fmt.Sprintf("%s:%s", p.Resource, p.Action), "error", err)
			continue
		}
		// 将权限分配给 admin 角色
		if err := bizSvc.AssignPermission(ctx, "admin", perm.Code); err != nil {
			applog.Info("seed: assign permission to admin failed", "code", perm.Code, "error", err)
		}
	}

	// 创建默认 admin 用户
	adminUser, err := bizSvc.CreateUser(ctx, &biz.CreateUserRequest{
		Username: "admin",
		Password: "admin123",
		Nickname: "管理员",
		Email:    "admin@example.com",
	})
	if err != nil {
		applog.Info("seed: admin user already exists or creation failed", "error", err)
		return
	}

	// 为 admin 用户分配 admin 角色
	if err := bizSvc.AssignRole(ctx, adminUser.ID, "admin"); err != nil {
		applog.Info("seed: assign admin role to admin user failed", "error", err)
	}

	applog.Info("seed: default data initialized")
}

// ==================== Wire ====================

// Wire auth 模块 IoC 装配
//
// 创建所有依赖并返回 AuthAPI 和 PermissionChecker。
func Wire(db *gorm.DB, jwtMgr *auth.JWTManager, seedEnabled bool) (*api.AuthAPI, auth.PermissionChecker) {
	// 仓储层
	userDAO := authmodel.NewUserDAO(db)
	roleDAO := authmodel.NewRoleDAO(db)
	permDAO := authmodel.NewPermissionDAO(db)

	// 端口实现
	hasher := &bcryptHasher{}
	cache := &biz.StubCacheInvalidator{}
	audit := &biz.StubAuditLogger{}

	// 业务层
	bizSvc := biz.NewAuthBizService(userDAO, roleDAO, permDAO, hasher, jwtMgr, cache, audit)

	// 应用层
	appSvc := service.NewAuthAppService(bizSvc)

	// PermissionChecker
	checker := &permissionChecker{biz: bizSvc}

	// API 层
	authAPI := api.NewAuthAPI(appSvc, checker)

	// 种子数据
	if seedEnabled {
		// 自动迁移表结构
		if err := db.AutoMigrate(
			&authmodel.UserModel{},
			&authmodel.RoleModel{},
			&authmodel.PermissionModel{},
			&authmodel.UserRoleModel{},
			&authmodel.RolePermissionModel{},
		); err != nil {
			applog.Info("auth: auto migrate failed", "error", err)
		}
		seedData(bizSvc)
	}

	return authAPI, checker
}
