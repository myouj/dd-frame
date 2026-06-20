package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	authdomain "github.com/example/dd-frame/internal/auth/domain"
	"github.com/example/dd-frame/internal/auth/service"
	"github.com/example/dd-frame/middleware"
	"github.com/example/dd-frame/pkg/auth"
	"github.com/example/dd-frame/pkg/response"
)

// AuthAPI auth 模块 HTTP Handler
type AuthAPI struct {
	svc     *service.AuthAppService
	checker auth.PermissionChecker
}

// NewAuthAPI 创建 auth API handler
func NewAuthAPI(svc *service.AuthAppService, checker auth.PermissionChecker) *AuthAPI {
	return &AuthAPI{svc: svc, checker: checker}
}

// RegisterPublicRoutes 注册无需认证的公开路由
func (a *AuthAPI) RegisterPublicRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.POST("/login", a.LoginHandler)
}

// RegisterRoutes 注册需要认证的路由
func (a *AuthAPI) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.GET("/me", a.MeHandler)
	g.PUT("/password", a.ChangePasswordHandler)

	// 用户管理（需要 user:* 权限）
	user := rg.Group("/user")
	user.Use(middleware.RequirePermission(a.checker, "user:read"))
	user.GET("", a.ListUsersHandler)
	user.GET("/:id", a.GetUserHandler)
	user.POST("", middleware.RequirePermission(a.checker, "user:create"), a.CreateUserHandler)
	user.PUT("/:id", middleware.RequirePermission(a.checker, "user:update"), a.UpdateUserHandler)
	user.DELETE("/:id", middleware.RequirePermission(a.checker, "user:delete"), a.DisableUserHandler)
	user.POST("/:id/roles", middleware.RequirePermission(a.checker, "user:update"), a.AssignRoleHandler)
	user.DELETE("/:id/roles/:roleCode", middleware.RequirePermission(a.checker, "user:update"), a.RevokeRoleHandler)

	// 角色管理（需要 role:* 权限）
	role := rg.Group("/role")
	role.Use(middleware.RequirePermission(a.checker, "role:read"))
	role.GET("", a.ListRolesHandler)
	role.GET("/:code", a.GetRoleHandler)
	role.POST("", middleware.RequirePermission(a.checker, "role:create"), a.CreateRoleHandler)
	role.PUT("/:code", middleware.RequirePermission(a.checker, "role:update"), a.UpdateRoleHandler)
	role.DELETE("/:code", middleware.RequirePermission(a.checker, "role:delete"), a.DeleteRoleHandler)
	role.POST("/:code/permissions", middleware.RequirePermission(a.checker, "role:update"), a.AssignPermHandler)
	role.DELETE("/:code/permissions/:permCode", middleware.RequirePermission(a.checker, "role:update"), a.RevokePermHandler)

	// 权限管理（需要 permission:* 权限）
	perm := rg.Group("/permission")
	perm.Use(middleware.RequirePermission(a.checker, "permission:read"))
	perm.GET("", a.ListPermissionsHandler)
	perm.POST("", middleware.RequirePermission(a.checker, "permission:create"), a.CreatePermissionHandler)
	perm.PUT("/:code", middleware.RequirePermission(a.checker, "permission:update"), a.UpdatePermissionHandler)
	perm.DELETE("/:code", middleware.RequirePermission(a.checker, "permission:delete"), a.DeletePermissionHandler)
}

// ==================== Auth Handlers ====================

// LoginHandler 登录
//
//	@Summary	用户登录
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		service.LoginInput	true	"登录参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	401		{object}	response.Response
//	@Router		/auth/login [post]
func (a *AuthAPI) LoginHandler(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	output, err := a.svc.Login(c.Request.Context(), &input)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, output)
}

// MeHandler 获取当前用户信息
//
//	@Summary	获取当前用户信息
//	@Tags		Auth
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Failure	401	{object}	response.Response
//	@Router		/auth/me [get]
func (a *AuthAPI) MeHandler(c *gin.Context) {
	user, ok := auth.CurrentUser(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40100, "unauthorized")
		return
	}

	output, err := a.svc.Me(c.Request.Context(), user)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, output)
}

// ChangePasswordHandler 修改密码
//
//	@Summary	修改密码
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		service.ChangePasswordInput	true	"修改密码参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	401		{object}	response.Response
//	@Router		/auth/password [put]
func (a *AuthAPI) ChangePasswordHandler(c *gin.Context) {
	user, ok := auth.CurrentUser(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40100, "unauthorized")
		return
	}

	var input service.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	if err := a.svc.ChangePassword(c.Request.Context(), user.UserID, &input); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// ==================== User Handlers ====================

// CreateUserHandler 创建用户
//
//	@Summary	创建用户
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		service.CreateUserInput	true	"创建用户参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	409		{object}	response.Response
//	@Router		/user [post]
func (a *AuthAPI) CreateUserHandler(c *gin.Context) {
	var input service.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	output, err := a.svc.CreateUser(c.Request.Context(), &input)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, output)
}

// GetUserHandler 获取用户详情
//
//	@Summary	获取用户详情
//	@Tags		User
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		int	true	"用户 ID"
//	@Success	200		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/user/{id} [get]
func (a *AuthAPI) GetUserHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	output, err := a.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, output)
}

// UpdateUserHandler 更新用户
//
//	@Summary	更新用户
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		int						true	"用户 ID"
//	@Param		body	body		service.UpdateUserInput	true	"更新用户参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/user/{id} [put]
func (a *AuthAPI) UpdateUserHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	if err := a.svc.UpdateUser(c.Request.Context(), id, &input); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// DisableUserHandler 禁用用户
//
//	@Summary	禁用用户
//	@Tags		User
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		int	true	"用户 ID"
//	@Success	200		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/user/{id} [delete]
func (a *AuthAPI) DisableUserHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	if err := a.svc.DisableUser(c.Request.Context(), id); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListUsersHandler 用户列表
//
//	@Summary	用户列表
//	@Tags		User
//	@Produce	json
//	@Security	BearerAuth
//	@Param		page		query		int	false	"页码（默认 1）"
//	@Param		pageSize	query		int	false	"每页条数（默认 20）"
//	@Success	200			{object}	response.Response
//	@Router		/user [get]
func (a *AuthAPI) ListUsersHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	output, err := a.svc.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "list users failed")
		return
	}

	response.Success(c, output)
}

// AssignRoleHandler 分配角色
//
//	@Summary	为用户分配角色
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		int						true	"用户 ID"
//	@Param		body	body		service.AssignRoleInput	true	"角色参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/user/{id}/roles [post]
func (a *AuthAPI) AssignRoleHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.AssignRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	if err := a.svc.AssignRole(c.Request.Context(), id, &input); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// RevokeRoleHandler 移除角色
//
//	@Summary	移除用户角色
//	@Tags		User
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id			path		int		true	"用户 ID"
//	@Param		roleCode	path		string	true	"角色编码"
//	@Success	200			{object}	response.Response
//	@Failure	404			{object}	response.Response
//	@Router		/user/{id}/roles/{roleCode} [delete]
func (a *AuthAPI) RevokeRoleHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	roleCode := c.Param("roleCode")
	if err := a.svc.RevokeRole(c.Request.Context(), id, roleCode); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// ==================== Role Handlers ====================

// CreateRoleHandler 创建角色
//
//	@Summary	创建角色
//	@Tags		Role
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		service.CreateRoleInput	true	"创建角色参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	409		{object}	response.Response
//	@Router		/role [post]
func (a *AuthAPI) CreateRoleHandler(c *gin.Context) {
	var input service.CreateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	output, err := a.svc.CreateRole(c.Request.Context(), &input)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, output)
}

// GetRoleHandler 获取角色详情
//
//	@Summary	获取角色详情
//	@Tags		Role
//	@Produce	json
//	@Security	BearerAuth
//	@Param		code	path		string	true	"角色编码"
//	@Success	200		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/role/{code} [get]
func (a *AuthAPI) GetRoleHandler(c *gin.Context) {
	code := c.Param("code")

	output, err := a.svc.GetRole(c.Request.Context(), code)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, output)
}

// UpdateRoleHandler 更新角色
//
//	@Summary	更新角色
//	@Tags		Role
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		code	path		string						true	"角色编码"
//	@Param		body	body		service.UpdateRoleInput	true	"更新角色参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/role/{code} [put]
func (a *AuthAPI) UpdateRoleHandler(c *gin.Context) {
	code := c.Param("code")

	var input service.UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	if err := a.svc.UpdateRole(c.Request.Context(), code, &input); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteRoleHandler 删除角色
//
//	@Summary	删除角色
//	@Tags		Role
//	@Produce	json
//	@Security	BearerAuth
//	@Param		code	path		string	true	"角色编码"
//	@Success	200		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/role/{code} [delete]
func (a *AuthAPI) DeleteRoleHandler(c *gin.Context) {
	code := c.Param("code")

	if err := a.svc.DeleteRole(c.Request.Context(), code); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListRolesHandler 角色列表
//
//	@Summary	角色列表
//	@Tags		Role
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/role [get]
func (a *AuthAPI) ListRolesHandler(c *gin.Context) {
	output, err := a.svc.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "list roles failed")
		return
	}

	response.Success(c, output)
}

// AssignPermHandler 分配权限
//
//	@Summary	为角色分配权限
//	@Tags		Role
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		code	path		string						true	"角色编码"
//	@Param		body	body		service.AssignPermInput	true	"权限参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/role/{code}/permissions [post]
func (a *AuthAPI) AssignPermHandler(c *gin.Context) {
	code := c.Param("code")

	var input service.AssignPermInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	if err := a.svc.AssignPermission(c.Request.Context(), code, &input); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// RevokePermHandler 移除权限
//
//	@Summary	移除角色权限
//	@Tags		Role
//	@Produce	json
//	@Security	BearerAuth
//	@Param		code		path		string	true	"角色编码"
//	@Param		permCode	path		string	true	"权限编码"
//	@Success	200			{object}	response.Response
//	@Failure	404			{object}	response.Response
//	@Router		/role/{code}/permissions/{permCode} [delete]
func (a *AuthAPI) RevokePermHandler(c *gin.Context) {
	code := c.Param("code")
	permCode := c.Param("permCode")

	if err := a.svc.RevokePermission(c.Request.Context(), code, permCode); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// ==================== Permission Handlers ====================

// CreatePermissionHandler 创建权限
//
//	@Summary	创建权限
//	@Tags		Permission
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		service.CreatePermissionInput	true	"创建权限参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	409		{object}	response.Response
//	@Router		/permission [post]
func (a *AuthAPI) CreatePermissionHandler(c *gin.Context) {
	var input service.CreatePermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	output, err := a.svc.CreatePermission(c.Request.Context(), &input)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, output)
}

// UpdatePermissionHandler 更新权限
//
//	@Summary	更新权限
//	@Tags		Permission
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		code	path		string								true	"权限编码"
//	@Param		body	body		service.UpdatePermissionInput	true	"更新权限参数"
//	@Success	200		{object}	response.Response
//	@Failure	400		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/permission/{code} [put]
func (a *AuthAPI) UpdatePermissionHandler(c *gin.Context) {
	code := c.Param("code")

	var input service.UpdatePermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	if err := a.svc.UpdatePermission(c.Request.Context(), code, &input); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// DeletePermissionHandler 删除权限
//
//	@Summary	删除权限
//	@Tags		Permission
//	@Produce	json
//	@Security	BearerAuth
//	@Param		code	path		string	true	"权限编码"
//	@Success	200		{object}	response.Response
//	@Failure	404		{object}	response.Response
//	@Router		/permission/{code} [delete]
func (a *AuthAPI) DeletePermissionHandler(c *gin.Context) {
	code := c.Param("code")

	if err := a.svc.DeletePermission(c.Request.Context(), code); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListPermissionsHandler 权限列表
//
//	@Summary	权限列表
//	@Tags		Permission
//	@Produce	json
//	@Security	BearerAuth
//	@Param		resource	query		string	false	"按资源筛选"
//	@Success	200			{object}	response.Response
//	@Router		/permission [get]
func (a *AuthAPI) ListPermissionsHandler(c *gin.Context) {
	resource := c.Query("resource")

	output, err := a.svc.ListPermissions(c.Request.Context(), resource)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "list permissions failed")
		return
	}

	response.Success(c, output)
}

// ==================== Error Mapping ====================

// handleDomainError 将领域错误映射为 HTTP 响应
func handleDomainError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, authdomain.ReasonUserNotFound):
		response.Error(c, http.StatusNotFound, 40400, "user not found")
	case strings.Contains(msg, authdomain.ReasonUserAlreadyExists):
		response.Error(c, http.StatusConflict, 40900, "user already exists")
	case strings.Contains(msg, authdomain.ReasonUserDisabled):
		response.Error(c, http.StatusForbidden, 40300, "user is disabled")
	case strings.Contains(msg, authdomain.ReasonInvalidCredentials):
		response.Error(c, http.StatusUnauthorized, 40101, "invalid username or password")
	case strings.Contains(msg, authdomain.ReasonRoleNotFound):
		response.Error(c, http.StatusNotFound, 40401, "role not found")
	case strings.Contains(msg, authdomain.ReasonRoleAlreadyExists):
		response.Error(c, http.StatusConflict, 40901, "role already exists")
	case strings.Contains(msg, authdomain.ReasonPermNotFound):
		response.Error(c, http.StatusNotFound, 40402, "permission not found")
	case strings.Contains(msg, authdomain.ReasonPermAlreadyExists):
		response.Error(c, http.StatusConflict, 40902, "permission already exists")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "internal server error")
	}
}
