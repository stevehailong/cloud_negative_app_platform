package handler

import (
	"my-cloud/internal/auth/service"
	"my-cloud/internal/common/model"
	"my-cloud/internal/common/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 登录
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	accessToken, refreshToken, user, err := h.authService.Login(req.Username, req.Password, c.ClientIP())
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	_, roles, _ := h.authService.GetUserInfo(user.ID)
	response.Success(c, gin.H{
		"token":        accessToken,
		"refreshToken": refreshToken,
		"user":         user,
		"roles":        roles,
	})
}

// Register 注册
type RegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
	Email      string `json:"email" binding:"required,email"`
	RealName   string `json:"realName"`
	Phone      string `json:"phone"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	user := &model.User{
		Username:   req.Username,
		Password:   req.Password,
		Email:      req.Email,
		RealName:   req.RealName,
		Phone:      req.Phone,
		Department: req.Department,
		Position:   req.Position,
	}

	if err := h.authService.Register(user); err != nil {
		response.Error(c, response.CodeConflict, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "注册成功"})
}

// GetUserInfo 获取用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, _ := c.Get("userId")
	user, roles, err := h.authService.GetUserInfo(userID.(uint))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, gin.H{
		"user":  user,
		"roles": roles,
	})
}

// UpdatePassword 修改密码
type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	userID, _ := c.Get("userId")
	if err := h.authService.UpdatePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}

// UpdateProfile 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	userID, _ := c.Get("userId")
	if err := h.authService.UpdateProfile(userID.(uint), updates); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "资料更新成功"})
}

// GetUserList 获取用户列表
func (h *AuthHandler) GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	users, total, err := h.authService.GetUserList(page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, users)
}

// AssignRoles 为用户分配角色
type AssignRolesRequest struct {
	RoleIDs []uint `json:"roleIds" binding:"required"`
}

func (h *AuthHandler) AssignRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.authService.AssignRoles(uint(userID), req.RoleIDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "角色分配成功"})
}

// UpdateUserStatus 更新用户状态
type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

func (h *AuthHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.authService.UpdateUserStatus(uint(userID), req.Status); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "状态更新成功"})
}

// GetAllRoles 获取所有角色
func (h *AuthHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.authService.GetAllRoles()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, roles)
}

// GetUserRoles 获取用户的角色列表
func (h *AuthHandler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	roles, err := h.authService.GetUserRoles(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, roles)
}

// RefreshToken Token刷新
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, RefreshTokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, gin.H{"message": "登出成功"})
}

// GetUserByID 获取用户详情
func (h *AuthHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	user, roles, err := h.authService.GetUserInfo(uint(id))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, gin.H{
		"user":  user,
		"roles": roles,
	})
}

// GetRoleList 获取角色列表
func (h *AuthHandler) GetRoleList(c *gin.Context) {
	roles, err := h.authService.GetAllRoles()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, roles)
}

// CreateUser 管理员创建用户
type CreateUserRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	RealName   string `json:"realName"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	user := &model.User{
		Username:   req.Username,
		Password:   req.Password,
		Email:      req.Email,
		Phone:      req.Phone,
		RealName:   req.RealName,
		Department: req.Department,
		Position:   req.Position,
	}

	if err := h.authService.CreateUser(user); err != nil {
		response.Error(c, response.CodeConflict, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "用户创建成功", "id": user.ID})
}

// UpdateUser 管理员更新用户
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.authService.UpdateUser(uint(id), updates); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "用户更新成功"})
}

// DeleteUser 管理员删除用户
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	if err := h.authService.DeleteUser(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "用户删除成功"})
}

// CreateRole 创建角色
type CreateRoleRequest struct {
	RoleCode    string `json:"roleCode" binding:"required"`
	RoleName    string `json:"roleName" binding:"required"`
	Description string `json:"description"`
}

func (h *AuthHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	role := &model.Role{
		Code:        req.RoleCode,
		Name:        req.RoleName,
		Description: req.Description,
	}

	if err := h.authService.CreateRole(role); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "角色创建成功", "id": role.ID})
}
