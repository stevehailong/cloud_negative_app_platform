package handler

import (
	"my-cloud/internal/auth/repository"
	"my-cloud/internal/common/model"
	"my-cloud/internal/common/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	permRepo *repository.PermissionRepository
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{
		permRepo: repository.NewPermissionRepository(db),
	}
}

// GetPermissionList 获取权限列表
func (h *PermissionHandler) GetPermissionList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	resourceType := c.Query("resourceType")

	permissions, total, err := h.permRepo.List(page, pageSize, resourceType)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items":    permissions,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetPermissionByID 根据ID获取权限
func (h *PermissionHandler) GetPermissionByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的权限ID")
		return
	}

	permission, err := h.permRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "权限不存在")
		return
	}

	response.Success(c, permission)
}

// CreatePermission 创建权限
type CreatePermissionRequest struct {
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	ResourceType string `json:"resourceType"`
	HttpMethod   string `json:"httpMethod"`
	Path         string `json:"path"`
	Description  string `json:"description"`
}

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 获取当前用户
	username, _ := c.Get("username")

	permission := &model.Permission{
		Code:         req.Code,
		Name:         req.Name,
		ResourceType: req.ResourceType,
		HttpMethod:   req.HttpMethod,
		Path:         req.Path,
		Description:  req.Description,
		Status:       1,
		CreatedBy:    username.(string),
	}

	if err := h.permRepo.Create(permission); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, permission)
}

// UpdatePermission 更新权限
type UpdatePermissionRequest struct {
	Name         string `json:"name"`
	ResourceType string `json:"resourceType"`
	HttpMethod   string `json:"httpMethod"`
	Path         string `json:"path"`
	Description  string `json:"description"`
	Status       *int   `json:"status"`
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的权限ID")
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 获取权限
	permission, err := h.permRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "权限不存在")
		return
	}

	// 获取当前用户
	username, _ := c.Get("username")

	// 更新字段
	if req.Name != "" {
		permission.Name = req.Name
	}
	if req.ResourceType != "" {
		permission.ResourceType = req.ResourceType
	}
	if req.HttpMethod != "" {
		permission.HttpMethod = req.HttpMethod
	}
	if req.Path != "" {
		permission.Path = req.Path
	}
	if req.Description != "" {
		permission.Description = req.Description
	}
	if req.Status != nil {
		permission.Status = *req.Status
	}
	permission.UpdatedBy = username.(string)

	if err := h.permRepo.Update(permission); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, permission)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的权限ID")
		return
	}

	if err := h.permRepo.Delete(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// GetRolePermissions 获取角色的权限列表
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的角色ID")
		return
	}

	permissions, err := h.permRepo.GetRolePermissions(uint(roleID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, permissions)
}

// AssignPermissionsToRole 为角色分配权限
type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permissionIds" binding:"required"`
}

func (h *PermissionHandler) AssignPermissionsToRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的角色ID")
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.permRepo.AssignPermissionsToRole(uint(roleID), req.PermissionIDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "权限分配成功"})
}

// RemovePermissionFromRole 从角色移除权限
func (h *PermissionHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的角色ID")
		return
	}

	permID, err := strconv.ParseUint(c.Param("permId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的权限ID")
		return
	}

	if err := h.permRepo.RemovePermissionFromRole(uint(roleID), uint(permID)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "权限移除成功"})
}

// GetUserPermissions 获取用户的权限列表
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	permissions, err := h.permRepo.GetUserPermissions(uint(userID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, permissions)
}
