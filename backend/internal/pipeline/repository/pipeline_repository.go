package repository

import (
	"my-cloud/internal/pipeline/model"

	"gorm.io/gorm"
)

type PipelineRepository struct {
	db *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) *PipelineRepository {
	return &PipelineRepository{db: db}
}

// Create 创建流水线
func (r *PipelineRepository) Create(pipeline *model.Pipeline) error {
	return r.db.Create(pipeline).Error
}

// GetByID 根据ID获取流水线
func (r *PipelineRepository) GetByID(id uint) (*model.Pipeline, error) {
	var pipeline model.Pipeline
	err := r.db.First(&pipeline, id).Error
	return &pipeline, err
}

// GetByCode 根据Code获取流水线
func (r *PipelineRepository) GetByCode(code string) (*model.Pipeline, error) {
	var pipeline model.Pipeline
	err := r.db.Where("pipeline_code = ?", code).First(&pipeline).Error
	if err != nil {
		return nil, err
	}
	return &pipeline, nil
}

// List 获取流水线列表
func (r *PipelineRepository) List(appID uint, page, pageSize int) ([]*model.Pipeline, int64, error) {
	var pipelines []*model.Pipeline
	var total int64

	query := r.db.Model(&model.Pipeline{})
	if appID > 0 {
		query = query.Where("app_id = ?", appID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&pipelines).Error
	return pipelines, total, err
}

// Update 更新流水线
func (r *PipelineRepository) Update(pipeline *model.Pipeline) error {
	return r.db.Save(pipeline).Error
}

// Delete 删除流水线
func (r *PipelineRepository) Delete(id uint) error {
	return r.db.Delete(&model.Pipeline{}, id).Error
}

// PipelineRunRepository 流水线执行记录仓库
type PipelineRunRepository struct {
	db *gorm.DB
}

func NewPipelineRunRepository(db *gorm.DB) *PipelineRunRepository {
	return &PipelineRunRepository{db: db}
}

// Create 创建流水线执行记录
func (r *PipelineRunRepository) Create(run *model.PipelineRun) error {
	return r.db.Create(run).Error
}

// GetByID 根据ID获取执行记录
func (r *PipelineRunRepository) GetByID(id uint) (*model.PipelineRun, error) {
	var run model.PipelineRun
	err := r.db.First(&run, id).Error
	return &run, err
}

// GetByRunNo 根据RunNo获取执行记录
func (r *PipelineRunRepository) GetByRunNo(runNo string) (*model.PipelineRun, error) {
	var run model.PipelineRun
	err := r.db.Where("run_no = ?", runNo).First(&run).Error
	return &run, err
}

// ListByPipeline 获取流水线的执行记录列表
func (r *PipelineRunRepository) ListByPipeline(pipelineID uint, page, pageSize int) ([]*model.PipelineRun, int64, error) {
	var runs []*model.PipelineRun
	var total int64

	query := r.db.Model(&model.PipelineRun{}).Where("pipeline_id = ?", pipelineID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&runs).Error
	return runs, total, err
}

// ListAll 获取所有流水线执行记录列表（不限定pipelineID）
func (r *PipelineRunRepository) ListAll(page, pageSize int, startDate, sortBy, sortOrder string) ([]*model.PipelineRun, int64, error) {
	var runs []*model.PipelineRun
	var total int64

	query := r.db.Model(&model.PipelineRun{})

	// 如果提供了startDate参数，按创建时间筛选
	if startDate != "" {
		query = query.Where("DATE(create_time) = ?", startDate)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 排序 - 映射前端字段到数据库字段
	orderClause := "id DESC"
	if sortBy != "" {
		// 将camelCase转换为snake_case
		dbField := sortBy
		if sortBy == "createTime" {
			dbField = "create_time"
		} else if sortBy == "updateTime" {
			dbField = "update_time"
		}
		
		direction := "DESC"
		if sortOrder == "asc" {
			direction = "ASC"
		}
		orderClause = dbField + " " + direction
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&runs).Error
	return runs, total, err
}

// Update 更新执行记录
func (r *PipelineRunRepository) Update(run *model.PipelineRun) error {
	return r.db.Save(run).Error
}

// ArtifactRepository 制品仓库
type ArtifactRepository struct {
	db *gorm.DB
}

func NewArtifactRepository(db *gorm.DB) *ArtifactRepository {
	return &ArtifactRepository{db: db}
}

// Create 创建制品
func (r *ArtifactRepository) Create(artifact *model.Artifact) error {
	return r.db.Create(artifact).Error
}

// GetByID 根据ID获取制品
func (r *ArtifactRepository) GetByID(id uint) (*model.Artifact, error) {
	var artifact model.Artifact
	err := r.db.First(&artifact, id).Error
	return &artifact, err
}

// ListByPipelineRun 获取流水线执行的制品列表
func (r *ArtifactRepository) ListByPipelineRun(pipelineRunID uint) ([]*model.Artifact, error) {
	var artifacts []*model.Artifact
	err := r.db.Where("pipeline_run_id = ?", pipelineRunID).Find(&artifacts).Error
	return artifacts, err
}

// List 获取制品列表
func (r *ArtifactRepository) List(artifactType string, page, pageSize int) ([]*model.Artifact, int64, error) {
	var artifacts []*model.Artifact
	var total int64

	query := r.db.Model(&model.Artifact{})
	if artifactType != "" {
		query = query.Where("artifact_type = ?", artifactType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&artifacts).Error
	return artifacts, total, err
}

// Delete 删除制品
func (r *ArtifactRepository) Update(artifact *model.Artifact) error {
	return r.db.Save(artifact).Error
}

func (r *ArtifactRepository) Delete(id uint) error {
	return r.db.Delete(&model.Artifact{}, id).Error
}
