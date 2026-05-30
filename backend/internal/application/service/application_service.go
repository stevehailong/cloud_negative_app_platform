package service

import (
	"errors"
	"my-cloud/internal/application/repository"
	"my-cloud/internal/common/model"
)

type ApplicationService struct {
	appRepo       *repository.ApplicationRepository
	componentRepo *repository.ComponentRepository
}

func NewApplicationService(appRepo *repository.ApplicationRepository, componentRepo *repository.ComponentRepository) *ApplicationService {
	return &ApplicationService{
		appRepo:       appRepo,
		componentRepo: componentRepo,
	}
}

// CreateApplication 创建应用
func (s *ApplicationService) CreateApplication(app *model.Application) error {
	// 检查应用编码是否已存在
	if existApp, _ := s.appRepo.GetByCode(app.Code); existApp != nil {
		return errors.New("应用编码已存在")
	}

	return s.appRepo.Create(app)
}

// GetApplication 获取应用详情
func (s *ApplicationService) GetApplication(id uint) (*model.Application, []*model.Component, error) {
	app, err := s.appRepo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	components, err := s.componentRepo.GetByApplicationID(id)
	if err != nil {
		return nil, nil, err
	}

	return app, components, nil
}

// UpdateApplication 更新应用
func (s *ApplicationService) UpdateApplication(app *model.Application) error {
	// 检查应用是否存在
	existApp, err := s.appRepo.GetByID(app.ID)
	if err != nil {
		return errors.New("应用不存在")
	}

	// 如果修改了编码，检查新编码是否已被使用
	if app.Code != existApp.Code {
		if dupApp, _ := s.appRepo.GetByCode(app.Code); dupApp != nil {
			return errors.New("应用编码已存在")
		}
	}

	return s.appRepo.Update(app)
}

// DeleteApplication 删除应用
func (s *ApplicationService) DeleteApplication(id uint) error {
	// 检查是否有关联的组件
	components, err := s.componentRepo.GetByApplicationID(id)
	if err != nil {
		return err
	}
	if len(components) > 0 {
		return errors.New("应用下还有组件，无法删除")
	}

	return s.appRepo.Delete(id)
}

// ListApplications 应用列表
func (s *ApplicationService) ListApplications(page, pageSize int, projectID uint, keyword string) ([]*model.Application, int64, error) {
	return s.appRepo.List(page, pageSize, projectID, keyword)
}

// CreateComponent 创建组件
func (s *ApplicationService) CreateComponent(component *model.Component) error {
	// 检查应用是否存在
	_, err := s.appRepo.GetByID(component.ApplicationID)
	if err != nil {
		return errors.New("应用不存在")
	}

	return s.componentRepo.Create(component)
}

// GetComponent 获取组件详情
func (s *ApplicationService) GetComponent(id uint) (*model.Component, error) {
	return s.componentRepo.GetByID(id)
}

// UpdateComponent 更新组件
func (s *ApplicationService) UpdateComponent(component *model.Component) error {
	_, err := s.componentRepo.GetByID(component.ID)
	if err != nil {
		return errors.New("组件不存在")
	}

	return s.componentRepo.Update(component)
}

// DeleteComponent 删除组件
func (s *ApplicationService) DeleteComponent(id uint) error {
	return s.componentRepo.Delete(id)
}

// ListComponents 组件列表
func (s *ApplicationService) ListComponents(page, pageSize int, appID uint) ([]*model.Component, int64, error) {
	return s.componentRepo.List(page, pageSize, appID)
}
