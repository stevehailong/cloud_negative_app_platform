package repository

import (
	"fmt"
	"my-cloud/internal/deploy/model"

	"gorm.io/gorm"
)

type AppDeploymentRepository struct {
	db *gorm.DB
}

func NewAppDeploymentRepository(db *gorm.DB) *AppDeploymentRepository {
	return &AppDeploymentRepository{db: db}
}

// Create 创建应用部署记录
func (r *AppDeploymentRepository) Create(deployment *model.AppDeployment) error {
	return r.db.Create(deployment).Error
}

// GetByID 根据ID获取
func (r *AppDeploymentRepository) GetByID(id int64) (*model.AppDeployment, error) {
	var deployment model.AppDeployment
	err := r.db.Where("id = ?", id).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetByAppAndEnv 根据app_id和env_id获取stable部署(向后兼容)
func (r *AppDeploymentRepository) GetByAppAndEnv(appID, envID int64) (*model.AppDeployment, error) {
	// 优先查找stable部署(workload_name = app-{appID})
	workloadName := fmt.Sprintf("app-%d", appID)
	var deployment model.AppDeployment
	err := r.db.Where("app_id = ? AND env_id = ? AND workload_name = ?", appID, envID, workloadName).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// ListByAppAndEnv 根据app_id和env_id获取所有部署(包括stable和canary)
func (r *AppDeploymentRepository) ListByAppAndEnv(appID, envID int64) ([]model.AppDeployment, error) {
	var deployments []model.AppDeployment
	err := r.db.Where("app_id = ? AND env_id = ?", appID, envID).Order("workload_name").Find(&deployments).Error
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

// GetByNamespace 根据namespace获取
func (r *AppDeploymentRepository) GetByNamespace(namespace string) (*model.AppDeployment, error) {
	var deployment model.AppDeployment
	err := r.db.Where("namespace = ?", namespace).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// List 列表查询
func (r *AppDeploymentRepository) List(appID, envID *int64, page, pageSize int) ([]model.AppDeployment, int64, error) {
	var deployments []model.AppDeployment
	var total int64

	query := r.db.Model(&model.AppDeployment{})
	
	if appID != nil {
		query = query.Where("app_id = ?", *appID)
	}
	if envID != nil {
		query = query.Where("env_id = ?", *envID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("update_time DESC").Find(&deployments).Error; err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// Update 更新
func (r *AppDeploymentRepository) Update(deployment *model.AppDeployment) error {
	return r.db.Save(deployment).Error
}

// UpdateFields 更新指定字段
func (r *AppDeploymentRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.AppDeployment{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除
func (r *AppDeploymentRepository) Delete(id int64) error {
	return r.db.Delete(&model.AppDeployment{}, id).Error
}

// GetByWorkloadName 根据namespace和workload_name查询
func (r *AppDeploymentRepository) GetByWorkloadName(namespace, workloadName string) (*model.AppDeployment, error) {
	var deployment model.AppDeployment
	err := r.db.Where("namespace = ? AND workload_name = ?", namespace, workloadName).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// BatchResolveAppNames 批量解析 app_id → app_name，填充到部署记录的 AppName 字段
// appDB 是连接到 app_db 的数据库连接
func (r *AppDeploymentRepository) BatchResolveAppNames(appDB *gorm.DB, deployments []model.AppDeployment) error {
	if len(deployments) == 0 || appDB == nil {
		return nil
	}

	// 收集唯一的 app_id
	appIDs := make(map[int64]bool)
	for _, d := range deployments {
		appIDs[d.AppID] = true
	}

	// 批量查询应用名称
	type AppInfo struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	var apps []AppInfo
	idList := make([]int64, 0, len(appIDs))
	for id := range appIDs {
		idList = append(idList, id)
	}
	if err := appDB.Table("applications").Where("id IN ?", idList).Find(&apps).Error; err != nil {
		return fmt.Errorf("failed to resolve app names: %w", err)
	}

	// 构建映射
	nameMap := make(map[int64]string, len(apps))
	for _, a := range apps {
		nameMap[a.ID] = a.Name
	}

	// 填充到部署记录
	for i := range deployments {
		if name, ok := nameMap[deployments[i].AppID]; ok {
			deployments[i].AppName = name
		}
	}

	return nil
}

// HasDeployingRecord 检查指定 app 是否有正在部署中的记录
func (r *AppDeploymentRepository) HasDeployingRecord(appID int64) (bool, *model.AppDeployment, error) {
	var deploying model.AppDeployment
	err := r.db.Where("app_id = ? AND deployment_status IN ?", appID, []string{"progressing", "deploying"}).
		First(&deploying).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	return true, &deploying, nil
}
