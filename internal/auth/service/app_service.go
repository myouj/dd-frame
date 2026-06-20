package service

import (
	"context"

	authbiz "github.com/example/dd-frame/internal/auth/biz"
	"github.com/example/dd-frame/pkg/auth"
)

// AuthAppService auth 应用边界服务
//
// 负责 HTTP DTO ↔ 业务 DTO 的转换，调用 biz 层编排。
type AuthAppService struct {
	biz authbiz.AuthBizService
}

// NewAuthAppService 创建 auth 应用边界服务
func NewAuthAppService(biz authbiz.AuthBizService) *AuthAppService {
	return &AuthAppService{biz: biz}
}

// ==================== Auth DTO ====================

// LoginInput 登录入参
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginOutput 登录出参
type LoginOutput struct {
	Token string   `json:"token"`
	Roles []string `json:"roles"`
}

// ChangePasswordInput 修改密码入参
type ChangePasswordInput struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// MeOutput 当前用户信息
type MeOutput struct {
	UserID      int64    `json:"userId"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// ==================== User DTO ====================

// CreateUserInput 创建用户入参
type CreateUserInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// UpdateUserInput 更新用户入参
type UpdateUserInput struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// UserOutput 用户出参
type UserOutput struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles,omitempty"`
}

// ListUsersOutput 用户列表出参
type ListUsersOutput struct {
	Total int64         `json:"total"`
	Users []*UserOutput `json:"users"`
}

// AssignRoleInput 分配角色入参
type AssignRoleInput struct {
	RoleCode string `json:"roleCode" binding:"required"`
}

// ==================== Role DTO ====================

// CreateRoleInput 创建角色入参
type CreateRoleInput struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateRoleInput 更新角色入参
type UpdateRoleInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoleOutput 角色出参
type RoleOutput struct {
	ID          int64             `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      int               `json:"status"`
	Permissions []*PermissionOutput `json:"permissions,omitempty"`
}

// AssignPermInput 分配权限入参
type AssignPermInput struct {
	PermCode string `json:"permCode" binding:"required"`
}

// ==================== Permission DTO ====================

// CreatePermissionInput 创建权限入参
type CreatePermissionInput struct {
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdatePermissionInput 更新权限入参
type UpdatePermissionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PermissionOutput 权限出参
type PermissionOutput struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ==================== Service 方法 ====================

// Login 登录
func (s *AuthAppService) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	token, roles, err := s.biz.Login(ctx, input.Username, input.Password)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{Token: token, Roles: roles}, nil
}

// ChangePassword 修改密码
func (s *AuthAppService) ChangePassword(ctx context.Context, userID int64, input *ChangePasswordInput) error {
	return s.biz.ChangePassword(ctx, userID, input.OldPassword, input.NewPassword)
}

// Me 获取当前用户信息
func (s *AuthAppService) Me(ctx context.Context, user *auth.AuthUser) (*MeOutput, error) {
	domainUser, err := s.biz.GetUser(ctx, user.UserID)
	if err != nil {
		return nil, err
	}

	perms, err := s.biz.GetUserPermissionCodes(ctx, user.UserID)
	if err != nil {
		return nil, err
	}

	roles := make([]string, len(domainUser.Roles))
	for i, r := range domainUser.Roles {
		roles[i] = r.Code
	}

	return &MeOutput{
		UserID:      domainUser.ID,
		Username:    domainUser.Username,
		Nickname:    domainUser.Nickname,
		Email:       domainUser.Email,
		Phone:       domainUser.Phone,
		Roles:       roles,
		Permissions: perms,
	}, nil
}

// CreateUser 创建用户
func (s *AuthAppService) CreateUser(ctx context.Context, input *CreateUserInput) (*UserOutput, error) {
	user, err := s.biz.CreateUser(ctx, &authbiz.CreateUserRequest{
		Username: input.Username,
		Password: input.Password,
		Nickname: input.Nickname,
		Email:    input.Email,
		Phone:    input.Phone,
	})
	if err != nil {
		return nil, err
	}
	return &UserOutput{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Status:   user.Status.String(),
	}, nil
}

// GetUser 获取用户详情
func (s *AuthAppService) GetUser(ctx context.Context, id int64) (*UserOutput, error) {
	user, err := s.biz.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Code
	}
	return &UserOutput{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Status:   user.Status.String(),
		Roles:    roles,
	}, nil
}

// UpdateUser 更新用户
func (s *AuthAppService) UpdateUser(ctx context.Context, id int64, input *UpdateUserInput) error {
	return s.biz.UpdateUser(ctx, id, &authbiz.UpdateUserRequest{
		Nickname: input.Nickname,
		Email:    input.Email,
		Phone:    input.Phone,
	})
}

// DisableUser 禁用用户
func (s *AuthAppService) DisableUser(ctx context.Context, id int64) error {
	return s.biz.DisableUser(ctx, id)
}

// ListUsers 用户列表
func (s *AuthAppService) ListUsers(ctx context.Context, page, pageSize int) (*ListUsersOutput, error) {
	users, total, err := s.biz.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	outputs := make([]*UserOutput, len(users))
	for i, u := range users {
		outputs[i] = &UserOutput{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Phone:    u.Phone,
			Status:   u.Status.String(),
		}
	}
	return &ListUsersOutput{Total: total, Users: outputs}, nil
}

// AssignRole 分配角色
func (s *AuthAppService) AssignRole(ctx context.Context, userID int64, input *AssignRoleInput) error {
	return s.biz.AssignRole(ctx, userID, input.RoleCode)
}

// RevokeRole 移除角色
func (s *AuthAppService) RevokeRole(ctx context.Context, userID int64, roleCode string) error {
	return s.biz.RevokeRole(ctx, userID, roleCode)
}

// CreateRole 创建角色
func (s *AuthAppService) CreateRole(ctx context.Context, input *CreateRoleInput) (*RoleOutput, error) {
	role, err := s.biz.CreateRole(ctx, &authbiz.CreateRoleRequest{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
	})
	if err != nil {
		return nil, err
	}
	return &RoleOutput{ID: role.ID, Code: role.Code, Name: role.Name, Description: role.Description, Status: role.Status}, nil
}

// GetRole 获取角色详情
func (s *AuthAppService) GetRole(ctx context.Context, code string) (*RoleOutput, error) {
	role, err := s.biz.GetRole(ctx, code)
	if err != nil {
		return nil, err
	}
	perms := make([]*PermissionOutput, len(role.Permissions))
	for i, p := range role.Permissions {
		perms[i] = &PermissionOutput{ID: p.ID, Code: p.Code, Resource: p.Resource, Action: p.Action, Name: p.Name, Description: p.Description}
	}
	return &RoleOutput{ID: role.ID, Code: role.Code, Name: role.Name, Description: role.Description, Status: role.Status, Permissions: perms}, nil
}

// UpdateRole 更新角色
func (s *AuthAppService) UpdateRole(ctx context.Context, code string, input *UpdateRoleInput) error {
	return s.biz.UpdateRole(ctx, code, &authbiz.UpdateRoleRequest{Name: input.Name, Description: input.Description})
}

// DeleteRole 删除角色
func (s *AuthAppService) DeleteRole(ctx context.Context, code string) error {
	return s.biz.DeleteRole(ctx, code)
}

// ListRoles 角色列表
func (s *AuthAppService) ListRoles(ctx context.Context) ([]*RoleOutput, error) {
	roles, err := s.biz.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	outputs := make([]*RoleOutput, len(roles))
	for i, r := range roles {
		outputs[i] = &RoleOutput{ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description, Status: r.Status}
	}
	return outputs, nil
}

// AssignPermission 分配权限
func (s *AuthAppService) AssignPermission(ctx context.Context, roleCode string, input *AssignPermInput) error {
	return s.biz.AssignPermission(ctx, roleCode, input.PermCode)
}

// RevokePermission 移除权限
func (s *AuthAppService) RevokePermission(ctx context.Context, roleCode string, permCode string) error {
	return s.biz.RevokePermission(ctx, roleCode, permCode)
}

// CreatePermission 创建权限
func (s *AuthAppService) CreatePermission(ctx context.Context, input *CreatePermissionInput) (*PermissionOutput, error) {
	perm, err := s.biz.CreatePermission(ctx, &authbiz.CreatePermissionRequest{
		Resource:    input.Resource,
		Action:      input.Action,
		Name:        input.Name,
		Description: input.Description,
	})
	if err != nil {
		return nil, err
	}
	return &PermissionOutput{ID: perm.ID, Code: perm.Code, Resource: perm.Resource, Action: perm.Action, Name: perm.Name, Description: perm.Description}, nil
}

// UpdatePermission 更新权限
func (s *AuthAppService) UpdatePermission(ctx context.Context, code string, input *UpdatePermissionInput) error {
	return s.biz.UpdatePermission(ctx, code, &authbiz.UpdatePermissionRequest{Name: input.Name, Description: input.Description})
}

// DeletePermission 删除权限
func (s *AuthAppService) DeletePermission(ctx context.Context, code string) error {
	return s.biz.DeletePermission(ctx, code)
}

// ListPermissions 权限列表
func (s *AuthAppService) ListPermissions(ctx context.Context, resource string) ([]*PermissionOutput, error) {
	perms, err := s.biz.ListPermissions(ctx, resource)
	if err != nil {
		return nil, err
	}
	outputs := make([]*PermissionOutput, len(perms))
	for i, p := range perms {
		outputs[i] = &PermissionOutput{ID: p.ID, Code: p.Code, Resource: p.Resource, Action: p.Action, Name: p.Name, Description: p.Description}
	}
	return outputs, nil
}
