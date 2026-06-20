package domain

import "time"

// User 用户聚合根
type User struct {
	ID        int64      // 主键
	Username  string     // 用户名（唯一）
	Password  string     // bcrypt 哈希密码
	Nickname  string     // 昵称
	Email     string     // 邮箱
	Phone     string     // 手机号
	Status    UserStatus // 状态
	Roles     []Role     // 关联角色（查询时填充）
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsActive 用户是否启用
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// Disable 禁用用户
func (u *User) Disable() {
	u.Status = UserStatusDisabled
}

// UpdateProfile 更新基础信息
func (u *User) UpdateProfile(nickname, email, phone string) {
	u.Nickname = nickname
	u.Email = email
	u.Phone = phone
}

// Role 角色聚合根
type Role struct {
	ID          int64        // 主键
	Code        string       // 角色编码（唯一，如 admin）
	Name        string       // 角色名称
	Description string       // 描述
	Status      int          // 1=启用 0=禁用
	Permissions []Permission // 关联权限（查询时填充）
	CreatedAt   time.Time
}

// IsActive 角色是否启用
func (r *Role) IsActive() bool {
	return r.Status == 1
}

// Permission 权限聚合根
type Permission struct {
	ID          int64  // 主键
	Code        string // 权限码（唯一，如 order:create）
	Resource    string // 资源名（如 order）
	Action      string // 操作名（如 create）
	Name        string // 权限名称
	Description string // 描述
	CreatedAt   time.Time
}

// BuildCode 从 resource + action 构建权限码
func BuildCode(resource, action string) string {
	return resource + ":" + action
}
