package model

// Application 应用模型
type Application struct {
	BaseModel
	Name         string `gorm:"size:100;not null;comment:应用名称" json:"name"`
	Code         string `gorm:"size:100;uniqueIndex;not null;comment:应用编码" json:"code"`
	ProjectID    uint   `gorm:"not null;index;comment:项目ID" json:"projectId"`
	Description  string `gorm:"size:500;comment:应用描述" json:"description"`
	Type         string `gorm:"size:50;comment:应用类型:web,api,job,function" json:"type"`
	Language     string `gorm:"size:50;comment:开发语言" json:"language"`
	Framework    string `gorm:"size:50;comment:开发框架" json:"framework"`
	RepoURL      string `gorm:"size:500;comment:代码仓库地址" json:"repoUrl"`
	RepoBranch   string `gorm:"size:100;comment:默认分支" json:"repoBranch"`
	BuildTool    string `gorm:"size:50;comment:构建工具" json:"buildTool"`
	BuildPath    string `gorm:"size:200;comment:构建路径" json:"buildPath"`
	DockerFile   string `gorm:"size:200;comment:Dockerfile路径" json:"dockerFile"`
	HealthCheck  string `gorm:"type:text;comment:健康检查配置" json:"healthCheck"`
	Labels       string `gorm:"type:text;comment:标签(JSON)" json:"labels"`
	Owner        string `gorm:"size:100;comment:负责人" json:"owner"`
}

// Component 组件模型
type Component struct {
	BaseModel
	ApplicationID uint   `gorm:"not null;index;comment:应用ID" json:"applicationId"`
	Name          string `gorm:"size:100;not null;comment:组件名称" json:"name"`
	Type          string `gorm:"size:50;comment:组件类型:frontend,backend,database,cache" json:"type"`
	Version       string `gorm:"size:50;comment:版本" json:"version"`
	Image         string `gorm:"size:500;comment:镜像地址" json:"image"`
	Port          int    `gorm:"comment:端口" json:"port"`
	Replicas      int    `gorm:"default:1;comment:副本数" json:"replicas"`
	CPU           string `gorm:"size:20;comment:CPU限制" json:"cpu"`
	Memory        string `gorm:"size:20;comment:内存限制" json:"memory"`
	EnvVars       string `gorm:"type:text;comment:环境变量(JSON)" json:"envVars"`
	ConfigMaps    string `gorm:"type:text;comment:配置映射(JSON)" json:"configMaps"`
	Secrets       string `gorm:"type:text;comment:密钥(JSON)" json:"secrets"`
	Volumes       string `gorm:"type:text;comment:存储卷(JSON)" json:"volumes"`
}

func (Application) TableName() string {
	return "applications"
}

func (Component) TableName() string {
	return "components"
}
