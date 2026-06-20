package model

import (
	"context"
	"time"

	"github.com/example/dd-frame/internal/auth/domain"
	"gorm.io/gorm"
)

// ==================== DB 模型 ====================

// UserModel 用户表
type UserModel struct {
	ID        int64     `gorm:"primary_key;auto_increment" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:64" json:"username"`
	Password  string    `gorm:"size:255" json:"-"`
	Nickname  string    `gorm:"size:64" json:"nickname"`
	Email     string    `gorm:"size:128" json:"email"`
	Phone     string    `gorm:"size:20" json:"phone"`
	Status    int       `gorm:"default:1;index" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserModel) TableName() string { return "users" }

// RoleModel 角色表
type RoleModel struct {
	ID          int64     `gorm:"primary_key;auto_increment" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:32" json:"code"`
	Name        string    `gorm:"size:64" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RoleModel) TableName() string { return "roles" }

// PermissionModel 权限表
type PermissionModel struct {
	ID          int64     `gorm:"primary_key;auto_increment" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:64" json:"code"`
	Resource    string    `gorm:"size:32;index" json:"resource"`
	Action      string    `gorm:"size:32" json:"action"`
	Name        string    `gorm:"size:64" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PermissionModel) TableName() string { return "permissions" }

// UserRoleModel 用户-角色关联表
type UserRoleModel struct {
	UserID    int64     `gorm:"primaryKey" json:"user_id"`
	RoleID    int64     `gorm:"primaryKey;index" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserRoleModel) TableName() string { return "user_roles" }

// RolePermissionModel 角色-权限关联表
type RolePermissionModel struct {
	RoleID       int64     `gorm:"primaryKey" json:"role_id"`
	PermissionID int64     `gorm:"primaryKey;index" json:"permission_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RolePermissionModel) TableName() string { return "role_permissions" }

// ==================== DAO 实现 ====================

// 编译期校验
var (
	_ UserRepo       = (*UserDAO)(nil)
	_ RoleRepo       = (*RoleDAO)(nil)
	_ PermissionRepo = (*PermissionDAO)(nil)
)

// ---------- UserDAO ----------

// UserDAO 用户仓储 GORM 实现
type UserDAO struct {
	db *gorm.DB
}

// NewUserDAO 创建用户 DAO
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) Create(ctx context.Context, user *domain.User) error {
	m := userToModel(user)
	if err := d.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	return nil
}

func (d *UserDAO) QueryByID(ctx context.Context, id int64) (*domain.User, error) {
	var m UserModel
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return modelToUser(&m), nil
}

func (d *UserDAO) QueryByUsername(ctx context.Context, username string) (*domain.User, error) {
	var m UserModel
	if err := d.db.WithContext(ctx).Where("username = ?", username).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return modelToUser(&m), nil
}

func (d *UserDAO) Update(ctx context.Context, user *domain.User) error {
	return d.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"nickname": user.Nickname,
			"email":    user.Email,
			"phone":    user.Phone,
			"status":   int(user.Status),
			"password": user.Password,
		}).Error
}

func (d *UserDAO) UpdateStatus(ctx context.Context, id int64, status domain.UserStatus) error {
	return d.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).
		Update("status", int(status)).Error
}

func (d *UserDAO) List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	var models []UserModel
	var total int64

	if err := d.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := d.db.WithContext(ctx).Offset(offset).Limit(pageSize).
		Order("id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = modelToUser(&models[i])
	}
	return users, total, nil
}

func (d *UserDAO) AssignRole(ctx context.Context, userID int64, roleID int64) error {
	ur := UserRoleModel{UserID: userID, RoleID: roleID}
	return d.db.WithContext(ctx).FirstOrCreate(&ur, ur).Error
}

func (d *UserDAO) RevokeRole(ctx context.Context, userID int64, roleCode string) error {
	var role RoleModel
	if err := d.db.WithContext(ctx).Where("code = ?", roleCode).First(&role).Error; err != nil {
		return err
	}
	return d.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, role.ID).
		Delete(&UserRoleModel{}).Error
}

func (d *UserDAO) QueryRolesByUserID(ctx context.Context, userID int64) ([]domain.Role, error) {
	var roles []RoleModel
	if err := d.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Role, len(roles))
	for i, r := range roles {
		result[i] = *roleModelToDomain(&r)
	}
	return result, nil
}

func (d *UserDAO) QueryPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error) {
	var codes []string
	if err := d.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

// ---------- RoleDAO ----------

// RoleDAO 角色仓储 GORM 实现
type RoleDAO struct {
	db *gorm.DB
}

// NewRoleDAO 创建角色 DAO
func NewRoleDAO(db *gorm.DB) *RoleDAO {
	return &RoleDAO{db: db}
}

func (d *RoleDAO) Create(ctx context.Context, role *domain.Role) error {
	m := roleToModel(role)
	if err := d.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	role.ID = m.ID
	return nil
}

func (d *RoleDAO) QueryByCode(ctx context.Context, code string) (*domain.Role, error) {
	var m RoleModel
	if err := d.db.WithContext(ctx).Where("code = ?", code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return roleModelToDomain(&m), nil
}

func (d *RoleDAO) QueryByID(ctx context.Context, id int64) (*domain.Role, error) {
	var m RoleModel
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return roleModelToDomain(&m), nil
}

func (d *RoleDAO) Update(ctx context.Context, role *domain.Role) error {
	return d.db.WithContext(ctx).Model(&RoleModel{}).Where("code = ?", role.Code).
		Updates(map[string]interface{}{
			"name":        role.Name,
			"description": role.Description,
			"status":      role.Status,
		}).Error
}

func (d *RoleDAO) Delete(ctx context.Context, code string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role RoleModel
		if err := tx.Where("code = ?", code).First(&role).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&UserRoleModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&RolePermissionModel{}).Error; err != nil {
			return err
		}
		return tx.Delete(&role).Error
	})
}

func (d *RoleDAO) List(ctx context.Context) ([]*domain.Role, error) {
	var models []RoleModel
	if err := d.db.WithContext(ctx).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(models))
	for i := range models {
		roles[i] = roleModelToDomain(&models[i])
	}
	return roles, nil
}

func (d *RoleDAO) AssignPermission(ctx context.Context, roleID int64, permissionID int64) error {
	rp := RolePermissionModel{RoleID: roleID, PermissionID: permissionID}
	return d.db.WithContext(ctx).FirstOrCreate(&rp, rp).Error
}

func (d *RoleDAO) RevokePermission(ctx context.Context, roleID int64, permCode string) error {
	var perm PermissionModel
	if err := d.db.WithContext(ctx).Where("code = ?", permCode).First(&perm).Error; err != nil {
		return err
	}
	return d.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, perm.ID).
		Delete(&RolePermissionModel{}).Error
}

func (d *RoleDAO) QueryPermissionsByRoleID(ctx context.Context, roleID int64) ([]domain.Permission, error) {
	var perms []PermissionModel
	if err := d.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Permission, len(perms))
	for i, p := range perms {
		result[i] = *permModelToDomain(&p)
	}
	return result, nil
}

// ---------- PermissionDAO ----------

// PermissionDAO 权限仓储 GORM 实现
type PermissionDAO struct {
	db *gorm.DB
}

// NewPermissionDAO 创建权限 DAO
func NewPermissionDAO(db *gorm.DB) *PermissionDAO {
	return &PermissionDAO{db: db}
}

func (d *PermissionDAO) Create(ctx context.Context, perm *domain.Permission) error {
	m := permToModel(perm)
	if err := d.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	perm.ID = m.ID
	return nil
}

func (d *PermissionDAO) QueryByCode(ctx context.Context, code string) (*domain.Permission, error) {
	var m PermissionModel
	if err := d.db.WithContext(ctx).Where("code = ?", code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return permModelToDomain(&m), nil
}

func (d *PermissionDAO) Update(ctx context.Context, perm *domain.Permission) error {
	return d.db.WithContext(ctx).Model(&PermissionModel{}).Where("code = ?", perm.Code).
		Updates(map[string]interface{}{
			"name":        perm.Name,
			"description": perm.Description,
		}).Error
}

func (d *PermissionDAO) Delete(ctx context.Context, code string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var perm PermissionModel
		if err := tx.Where("code = ?", code).First(&perm).Error; err != nil {
			return err
		}
		if err := tx.Where("permission_id = ?", perm.ID).Delete(&RolePermissionModel{}).Error; err != nil {
			return err
		}
		return tx.Delete(&perm).Error
	})
}

func (d *PermissionDAO) List(ctx context.Context, resource string) ([]*domain.Permission, error) {
	var models []PermissionModel
	query := d.db.WithContext(ctx).Order("code ASC")
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	perms := make([]*domain.Permission, len(models))
	for i := range models {
		perms[i] = permModelToDomain(&models[i])
	}
	return perms, nil
}

// ==================== Converter ====================

func userToModel(u *domain.User) *UserModel {
	return &UserModel{
		ID:       u.ID,
		Username: u.Username,
		Password: u.Password,
		Nickname: u.Nickname,
		Email:    u.Email,
		Phone:    u.Phone,
		Status:   int(u.Status),
	}
}

func modelToUser(m *UserModel) *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		ID:        m.ID,
		Username:  m.Username,
		Password:  m.Password,
		Nickname:  m.Nickname,
		Email:     m.Email,
		Phone:     m.Phone,
		Status:    domain.UserStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func roleToModel(r *domain.Role) *RoleModel {
	return &RoleModel{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Status:      r.Status,
	}
}

func roleModelToDomain(m *RoleModel) *domain.Role {
	return &domain.Role{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
	}
}

func permToModel(p *domain.Permission) *PermissionModel {
	return &PermissionModel{
		ID:          p.ID,
		Code:        p.Code,
		Resource:    p.Resource,
		Action:      p.Action,
		Name:        p.Name,
		Description: p.Description,
	}
}

func permModelToDomain(m *PermissionModel) *domain.Permission {
	return &domain.Permission{
		ID:          m.ID,
		Code:        m.Code,
		Resource:    m.Resource,
		Action:      m.Action,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}
