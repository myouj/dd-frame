package biz

import (
	"context"
	"fmt"

	"github.com/example/dd-frame/internal/auth/domain"
	authmodel "github.com/example/dd-frame/internal/auth/model"
	"github.com/example/dd-frame/pkg/auth"
)

// AuthBizService auth 业务能力接口
type AuthBizService interface {
	// 认证
	Login(ctx context.Context, username, password string) (string, []string, error)
	ChangePassword(ctx context.Context, userID int64, oldPwd, newPwd string) error

	// 用户管理
	CreateUser(ctx context.Context, req *CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, id int64) (*domain.User, error)
	UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) error
	DisableUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
	AssignRole(ctx context.Context, userID int64, roleCode string) error
	RevokeRole(ctx context.Context, userID int64, roleCode string) error

	// 角色管理
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*domain.Role, error)
	GetRole(ctx context.Context, code string) (*domain.Role, error)
	UpdateRole(ctx context.Context, code string, req *UpdateRoleRequest) error
	DeleteRole(ctx context.Context, code string) error
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	AssignPermission(ctx context.Context, roleCode string, permCode string) error
	RevokePermission(ctx context.Context, roleCode string, permCode string) error

	// 权限管理
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.Permission, error)
	UpdatePermission(ctx context.Context, code string, req *UpdatePermissionRequest) error
	DeletePermission(ctx context.Context, code string) error
	ListPermissions(ctx context.Context, resource string) ([]*domain.Permission, error)

	// 权限查询（供 PermissionChecker 使用）
	GetUserPermissionCodes(ctx context.Context, userID int64) ([]string, error)
	GetUserRoleCodes(ctx context.Context, userID int64) ([]string, error)
}

// ==================== DTO ====================

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string
	Email    string
	Phone    string
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Code        string
	Name        string
	Description string
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string
	Description string
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Resource    string
	Action      string
	Name        string
	Description string
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string
	Description string
}

// ==================== 实现 ====================

type authBizService struct {
	userRepo authmodel.UserRepo
	roleRepo authmodel.RoleRepo
	permRepo authmodel.PermissionRepo
	hasher   domain.PasswordHasher
	jwtMgr   *auth.JWTManager
	cache    CacheInvalidator
	audit    AuditLogger
}

// NewAuthBizService 创建 auth 业务服务
func NewAuthBizService(
	userRepo authmodel.UserRepo,
	roleRepo authmodel.RoleRepo,
	permRepo authmodel.PermissionRepo,
	hasher domain.PasswordHasher,
	jwtMgr *auth.JWTManager,
	cache CacheInvalidator,
	audit AuditLogger,
) AuthBizService {
	return &authBizService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		permRepo: permRepo,
		hasher:   hasher,
		jwtMgr:   jwtMgr,
		cache:    cache,
		audit:    audit,
	}
}

// ---------- 认证 ----------

func (s *authBizService) Login(ctx context.Context, username, password string) (string, []string, error) {
	user, err := s.userRepo.QueryByUsername(ctx, username)
	if err != nil {
		return "", nil, fmt.Errorf("query user failed: %w", err)
	}
	if user == nil {
		s.audit.LogLogin(ctx, 0, username, false)
		return "", nil, domain.ErrInvalidCredentials()
	}
	if !user.IsActive() {
		return "", nil, domain.ErrUserDisabled()
	}
	if !s.hasher.Verify(user.Password, password) {
		s.audit.LogLogin(ctx, user.ID, username, false)
		return "", nil, domain.ErrInvalidCredentials()
	}

	roles, err := s.userRepo.QueryRolesByUserID(ctx, user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("query roles failed: %w", err)
	}

	roleCodes := make([]string, len(roles))
	for i, r := range roles {
		roleCodes[i] = r.Code
	}

	token, err := s.jwtMgr.Generate(user.ID, user.Username, roleCodes)
	if err != nil {
		return "", nil, fmt.Errorf("generate token failed: %w", err)
	}

	s.audit.LogLogin(ctx, user.ID, username, true)
	return token, roleCodes, nil
}

func (s *authBizService) ChangePassword(ctx context.Context, userID int64, oldPwd, newPwd string) error {
	user, err := s.userRepo.QueryByID(ctx, userID)
	if err != nil || user == nil {
		return domain.ErrUserNotFound(fmt.Sprintf("%d", userID))
	}
	if !s.hasher.Verify(user.Password, oldPwd) {
		return domain.ErrInvalidCredentials()
	}

	hash, err := s.hasher.Hash(newPwd)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}
	user.Password = hash
	return s.userRepo.Update(ctx, user)
}

// ---------- 用户管理 ----------

func (s *authBizService) CreateUser(ctx context.Context, req *CreateUserRequest) (*domain.User, error) {
	existing, _ := s.userRepo.QueryByUsername(ctx, req.Username)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists(req.Username)
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	user := &domain.User{
		Username: req.Username,
		Password: hash,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   domain.UserStatusActive,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}
	return user, nil
}

func (s *authBizService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	user, err := s.userRepo.QueryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound(fmt.Sprintf("%d", id))
	}
	roles, _ := s.userRepo.QueryRolesByUserID(ctx, id)
	user.Roles = roles
	return user, nil
}

func (s *authBizService) UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) error {
	user, err := s.userRepo.QueryByID(ctx, id)
	if err != nil || user == nil {
		return domain.ErrUserNotFound(fmt.Sprintf("%d", id))
	}
	user.UpdateProfile(req.Nickname, req.Email, req.Phone)
	return s.userRepo.Update(ctx, user)
}

func (s *authBizService) DisableUser(ctx context.Context, id int64) error {
	user, err := s.userRepo.QueryByID(ctx, id)
	if err != nil || user == nil {
		return domain.ErrUserNotFound(fmt.Sprintf("%d", id))
	}
	user.Disable()
	if err := s.userRepo.UpdateStatus(ctx, id, user.Status); err != nil {
		return err
	}
	_ = s.cache.InvalidateUserPermissions(ctx, id)
	return nil
}

func (s *authBizService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	return s.userRepo.List(ctx, page, pageSize)
}

func (s *authBizService) AssignRole(ctx context.Context, userID int64, roleCode string) error {
	role, err := s.roleRepo.QueryByCode(ctx, roleCode)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(roleCode)
	}
	if err := s.userRepo.AssignRole(ctx, userID, role.ID); err != nil {
		return err
	}
	_ = s.cache.InvalidateUserPermissions(ctx, userID)
	s.audit.LogPermissionChange(ctx, userID, "assign_role", fmt.Sprintf("user=%d role=%s", userID, roleCode))
	return nil
}

func (s *authBizService) RevokeRole(ctx context.Context, userID int64, roleCode string) error {
	if err := s.userRepo.RevokeRole(ctx, userID, roleCode); err != nil {
		return err
	}
	_ = s.cache.InvalidateUserPermissions(ctx, userID)
	s.audit.LogPermissionChange(ctx, userID, "revoke_role", fmt.Sprintf("user=%d role=%s", userID, roleCode))
	return nil
}

// ---------- 角色管理 ----------

func (s *authBizService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*domain.Role, error) {
	existing, _ := s.roleRepo.QueryByCode(ctx, req.Code)
	if existing != nil {
		return nil, domain.ErrRoleAlreadyExists(req.Code)
	}
	role := &domain.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      1,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *authBizService) GetRole(ctx context.Context, code string) (*domain.Role, error) {
	role, err := s.roleRepo.QueryByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrRoleNotFound(code)
	}
	perms, _ := s.roleRepo.QueryPermissionsByRoleID(ctx, role.ID)
	role.Permissions = perms
	return role, nil
}

func (s *authBizService) UpdateRole(ctx context.Context, code string, req *UpdateRoleRequest) error {
	role, err := s.roleRepo.QueryByCode(ctx, code)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(code)
	}
	role.Name = req.Name
	role.Description = req.Description
	return s.roleRepo.Update(ctx, role)
}

func (s *authBizService) DeleteRole(ctx context.Context, code string) error {
	return s.roleRepo.Delete(ctx, code)
}

func (s *authBizService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *authBizService) AssignPermission(ctx context.Context, roleCode string, permCode string) error {
	role, err := s.roleRepo.QueryByCode(ctx, roleCode)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(roleCode)
	}
	perm, err := s.permRepo.QueryByCode(ctx, permCode)
	if err != nil || perm == nil {
		return domain.ErrPermNotFound(permCode)
	}
	if err := s.roleRepo.AssignPermission(ctx, role.ID, perm.ID); err != nil {
		return err
	}
	s.audit.LogPermissionChange(ctx, 0, "assign_perm", fmt.Sprintf("role=%s perm=%s", roleCode, permCode))
	return nil
}

func (s *authBizService) RevokePermission(ctx context.Context, roleCode string, permCode string) error {
	role, err := s.roleRepo.QueryByCode(ctx, roleCode)
	if err != nil || role == nil {
		return domain.ErrRoleNotFound(roleCode)
	}
	if err := s.roleRepo.RevokePermission(ctx, role.ID, permCode); err != nil {
		return err
	}
	s.audit.LogPermissionChange(ctx, 0, "revoke_perm", fmt.Sprintf("role=%s perm=%s", roleCode, permCode))
	return nil
}

// ---------- 权限管理 ----------

func (s *authBizService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.Permission, error) {
	code := domain.BuildCode(req.Resource, req.Action)
	existing, _ := s.permRepo.QueryByCode(ctx, code)
	if existing != nil {
		return nil, domain.ErrPermAlreadyExists(code)
	}
	perm := &domain.Permission{
		Code:        code,
		Resource:    req.Resource,
		Action:      req.Action,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.permRepo.Create(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *authBizService) UpdatePermission(ctx context.Context, code string, req *UpdatePermissionRequest) error {
	perm, err := s.permRepo.QueryByCode(ctx, code)
	if err != nil || perm == nil {
		return domain.ErrPermNotFound(code)
	}
	perm.Name = req.Name
	perm.Description = req.Description
	return s.permRepo.Update(ctx, perm)
}

func (s *authBizService) DeletePermission(ctx context.Context, code string) error {
	return s.permRepo.Delete(ctx, code)
}

func (s *authBizService) ListPermissions(ctx context.Context, resource string) ([]*domain.Permission, error) {
	return s.permRepo.List(ctx, resource)
}

// ---------- 权限查询（PermissionChecker） ----------

func (s *authBizService) GetUserPermissionCodes(ctx context.Context, userID int64) ([]string, error) {
	return s.userRepo.QueryPermissionCodesByUserID(ctx, userID)
}

func (s *authBizService) GetUserRoleCodes(ctx context.Context, userID int64) ([]string, error) {
	roles, err := s.userRepo.QueryRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(roles))
	for i, r := range roles {
		codes[i] = r.Code
	}
	return codes, nil
}
