package handler

import (
	"my-cloud/internal/common/model"
	"my-cloud/internal/project/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
	tenantRepo  *repository.TenantRepository
}

func NewProjectHandler(db *gorm.DB, iamDB *gorm.DB) *ProjectHandler {
	return &ProjectHandler{
		projectRepo: repository.NewProjectRepository(db, iamDB),
		tenantRepo:  repository.NewTenantRepository(db),
	}
}

// ListTenants 获取租户列表
func (h *ProjectHandler) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	tenants, total, err := h.tenantRepo.List(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取租户列表失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list": tenants,
			"total": total,
			"page":  page,
			"pageSize": pageSize,
		},
	})
}

// CreateTenant 创建租户
func (h *ProjectHandler) CreateTenant(c *gin.Context) {
	var req struct {
		TenantCode   string `json:"tenantCode" binding:"required"`
		TenantName   string `json:"tenantName" binding:"required"`
		ContactEmail string `json:"contactEmail"`
		ContactPhone string `json:"contactPhone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 检查租户编码是否已存在
	if _, err := h.tenantRepo.GetByCode(req.TenantCode); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "租户编码已存在", "data": nil})
		return
	}

	tenant := &model.Tenant{
		TenantCode:   req.TenantCode,
		TenantName:   req.TenantName,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		Status:       1,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}

	if err := h.tenantRepo.Create(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建租户失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": tenant})
}

// GetTenant 获取租户详情
func (h *ProjectHandler) GetTenant(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	tenant, err := h.tenantRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "租户不存在", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": tenant})
}

// UpdateTenant 更新租户
func (h *ProjectHandler) UpdateTenant(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	tenant, err := h.tenantRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "租户不存在", "data": nil})
		return
	}

	var req struct {
		TenantName   string `json:"tenantName"`
		ContactEmail string `json:"contactEmail"`
		ContactPhone string `json:"contactPhone"`
		Status       *int   `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	updateFields := make(map[string]interface{})
	
	if req.TenantName != "" {
		tenant.TenantName = req.TenantName
		updateFields["tenant_name"] = req.TenantName
	}
	// 联系方式允许清空
	tenant.ContactEmail = req.ContactEmail
	updateFields["contact_email"] = req.ContactEmail
	
	tenant.ContactPhone = req.ContactPhone
	updateFields["contact_phone"] = req.ContactPhone
	
	if req.Status != nil {
		tenant.Status = *req.Status
		updateFields["status"] = *req.Status
	}
	tenant.UpdateTime = time.Now()
	updateFields["update_time"] = tenant.UpdateTime

	if err := h.tenantRepo.UpdateFields(uint(id), updateFields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新租户失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": tenant})
}

// DeleteTenant 删除租户
func (h *ProjectHandler) DeleteTenant(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// 检查租户下是否有项目
	projects, _, err := h.projectRepo.List(1, 1, "", uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检查租户关联数据失败", "data": nil})
		return
	}
	if len(projects) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该租户下还有项目,无法删除。请先删除或迁移租户下的所有项目", "data": nil})
		return
	}

	if err := h.tenantRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除租户失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// ListProjects 获取项目列表
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	tenantIDStr := c.Query("tenantId")
	tenantID := uint(0)
	if tenantIDStr != "" {
		tid, _ := strconv.ParseUint(tenantIDStr, 10, 32)
		tenantID = uint(tid)
	}

	// 获取当前用户ID
	userIDVal, _ := c.Get("userId")
	userID, _ := userIDVal.(uint)

	var projects []*model.Project
	var total int64
	var err error

	// 管理员可以看到所有项目，普通用户只能看到公开项目和自己参与的私有项目
	if h.projectRepo.IsAdmin(userID) {
		projects, total, err = h.projectRepo.List(page, pageSize, keyword, tenantID)
	} else {
		projects, total, err = h.projectRepo.ListAccessible(page, pageSize, keyword, tenantID, userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取项目列表失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":    projects,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req struct {
		TenantID    uint   `json:"tenantId" binding:"required"`
		OrgID       *uint  `json:"orgId"`
		ProjectCode string `json:"projectCode" binding:"required"`
		ProjectName string `json:"projectName" binding:"required"`
		OwnerUserID *uint  `json:"ownerUserId"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 检查项目编码是否已存在
	if _, err := h.projectRepo.GetByCode(req.ProjectCode); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "项目编码已存在", "data": nil})
		return
	}

	// 获取当前用户ID
	userIDVal, _ := c.Get("userId")
	userID, _ := userIDVal.(uint)

	visibility := req.Visibility
	if visibility == "" {
		visibility = "private"
	}

	// 如果未指定owner，则设置为当前创建用户
	ownerUserID := req.OwnerUserID
	if ownerUserID == nil && userID > 0 {
		ownerUserID = &userID
	}

	project := &model.Project{
		TenantID:    req.TenantID,
		OrgID:       req.OrgID,
		ProjectCode: req.ProjectCode,
		ProjectName: req.ProjectName,
		OwnerUserID: ownerUserID,
		Description: req.Description,
		Visibility:  visibility,
		Status:      1,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	if err := h.projectRepo.Create(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建项目失败", "data": nil})
		return
	}

	// 自动将创建者添加为项目成员(owner角色)
	if userID > 0 {
		member := &model.ProjectMember{
			ProjectID:  project.ID,
			UserID:     userID,
			RoleCode:   "owner",
			CreateTime: time.Now(),
			CreateBy:   &userID,
		}
		if err := h.projectRepo.AddMember(member); err != nil {
			// 添加成员失败不影响项目创建,只记录日志
			c.JSON(http.StatusOK, gin.H{
				"code":    0, 
				"message": "项目创建成功，但添加成员失败", 
				"data":    project,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": project})
}

// GetProject 获取项目详情
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	project, err := h.projectRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "项目不存在", "data": nil})
		return
	}

	// 私有项目访问控制：非成员、非Owner、非管理员不可查看
	if project.Visibility == "private" {
		userIDVal, _ := c.Get("userId")
		userID, _ := userIDVal.(uint)

		isOwner := project.OwnerUserID != nil && *project.OwnerUserID == userID
		if !isOwner && !h.projectRepo.IsMember(uint(id), userID) && !h.projectRepo.IsAdmin(userID) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问该私有项目", "data": nil})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": project})
}

// UpdateProject 更新项目
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	project, err := h.projectRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "项目不存在", "data": nil})
		return
	}

	var req struct {
		TenantID    *uint  `json:"tenantId"`
		ProjectName string `json:"projectName"`
		OwnerUserID *uint  `json:"ownerUserId"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		Status      *int   `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	// 记录哪些字段需要更新
	updateFields := make(map[string]interface{})
	
	// 如果要更改租户,需要验证新租户是否存在
	if req.TenantID != nil && *req.TenantID != project.TenantID {
		newTenant, err := h.tenantRepo.GetByID(*req.TenantID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "目标租户不存在", "data": nil})
			return
		}
		if newTenant.Status != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "目标租户已禁用", "data": nil})
			return
		}
		project.TenantID = *req.TenantID
		updateFields["tenant_id"] = *req.TenantID
	}
	
	if req.ProjectName != "" {
		project.ProjectName = req.ProjectName
		updateFields["project_name"] = req.ProjectName
	}
	if req.OwnerUserID != nil {
		project.OwnerUserID = req.OwnerUserID
		updateFields["owner_user_id"] = req.OwnerUserID
	}
	// Description允许设置为空字符串
	project.Description = req.Description
	updateFields["description"] = req.Description
	
	if req.Visibility != "" {
		project.Visibility = req.Visibility
		updateFields["visibility"] = req.Visibility
	}
	if req.Status != nil {
		project.Status = *req.Status
		updateFields["status"] = *req.Status
	}
	project.UpdateTime = time.Now()
	updateFields["update_time"] = project.UpdateTime

	if err := h.projectRepo.UpdateFields(uint(id), updateFields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新项目失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": project})
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.projectRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除项目失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// GetProjectMembers 获取项目成员
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 32)

	members, err := h.projectRepo.GetMembers(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取项目成员失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": members})
}

// AddProjectMember 添加项目成员
func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 32)

	var req struct {
		UserID   uint   `json:"userId" binding:"required"`
		RoleCode string `json:"roleCode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "data": nil})
		return
	}

	member := &model.ProjectMember{
		ProjectID:  uint(projectID),
		UserID:     req.UserID,
		RoleCode:   req.RoleCode,
		CreateTime: time.Now(),
	}

	if err := h.projectRepo.AddMember(member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "添加项目成员失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": member})
}

// RemoveProjectMember 移除项目成员
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 32)
	userID, _ := strconv.ParseUint(c.Param("userId"), 10, 32)

	if err := h.projectRepo.RemoveMember(uint(projectID), uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "移除项目成员失败", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}
