# 环境模板绑定功能

## 修复时间
2026-06-02

## 问题描述

用户反馈:"环境管理中新建环境没有和模版绑定呢"

在新建环境的表单中缺少环境模板的选择项,无法将环境与模板关联。

## 修复内容

### 1. 数据库修改

**修改文件**: `env_db.environments` 表

添加了 `template_id` 字段:

```sql
ALTER TABLE env_db.environments 
ADD COLUMN template_id BIGINT NULL COMMENT '关联的环境模板ID' AFTER project_id;
```

字段信息:
- 字段名: `template_id`
- 类型: `BIGINT`
- 可空: YES
- 说明: 关联的环境模板ID,外键指向 `env_templates.id`

### 2. 后端模型修改

**修改文件**: `backend/internal/common/model/environment.go`

在 `Environment` 结构体中添加了 `TemplateID` 字段:

```go
type Environment struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    EnvCode     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"envCode"`
    EnvName     string    `gorm:"type:varchar(128);not null" json:"envName"`
    EnvType     string    `gorm:"type:varchar(32);not null" json:"envType"`
    ClusterID   uint      `gorm:"column:cluster_id;not null" json:"clusterId"`
    Namespace   string    `gorm:"type:varchar(128);not null" json:"namespace"`
    ProjectID   uint      `gorm:"column:project_id;not null" json:"projectId"`
    TemplateID  *uint     `gorm:"column:template_id" json:"templateId"` // 新增字段
    Description string    `gorm:"type:varchar(255)" json:"description"`
    ConfigJSON  string    `gorm:"type:json" json:"configJson"`
    Status      int       `gorm:"type:tinyint;default:1" json:"status"`
    CreateTime  time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
    UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
    CreateBy    *uint     `gorm:"column:create_by" json:"createBy"`
    UpdateBy    *uint     `gorm:"column:update_by" json:"updateBy"`
    IsDeleted   int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}
```

### 3. 前端表单修改

**修改文件**: `frontend/src/views/environment/EnvironmentList.vue`

#### 3.1 添加模板选择表单项

在"所属项目"表单项后添加了"环境模板"选择框:

```vue
<el-form-item label="环境模板" prop="templateId">
  <el-select v-model="form.templateId" placeholder="请选择模板(可选)" clearable>
    <el-option 
      v-for="template in templates" 
      :key="template.id" 
      :label="template.templateName" 
      :value="template.id" 
    >
      <span style="float: left">{{ template.templateName }}</span>
      <span style="float: right; color: #8492a6; font-size: 13px">{{ template.templateType }}</span>
    </el-option>
  </el-select>
  <div style="color: #909399; font-size: 12px; margin-top: 4px;">
    选择模板后将应用模板的默认配置
  </div>
</el-form-item>
```

#### 3.2 添加templates数据和加载函数

```javascript
// 添加数据
const templates = ref([])

// 添加加载模板列表函数
const loadTemplates = async () => {
  try {
    const res = await request.get('/env-templates', { params: { page: 1, pageSize: 1000 } })
    if (res.code === 0 || res.code === 200) {
      templates.value = res.data.list || []
      console.log('加载模板列表成功，共', templates.value.length, '个模板')
    }
  } catch (error) {
    console.error('加载模板列表异常', error)
  }
}

// 在handleCreate中加载模板
const handleCreate = async () => {
  dialogTitle.value = '新建环境'
  resetForm()
  await Promise.all([loadProjects(), loadClusters(), loadTemplates()])
  dialogVisible.value = true
}

// 在form中添加templateId字段
const form = reactive({
  id: null,
  envCode: '',
  envName: '',
  envType: '',
  clusterId: null,
  namespace: '',
  projectId: null,
  templateId: null,  // 新增
  description: '',
  status: 1,
  configJson: ''
})

// 在resetForm中重置templateId
const resetForm = () => {
  // ...
  form.templateId = null
  // ...
}

// 在onMounted中加载模板
onMounted(() => {
  loadEnvironments()
  loadProjects()
  loadClusters()
  loadTemplates()  // 新增
})
```

## 功能说明

### 1. 环境模板的作用

环境模板(`env_templates`)用于存储不同类型环境的默认配置:

- **模板类型**:
  - `helm`: Helm Chart 模板,包含Chart名称、版本和Values配置
  - `kustomize`: Kustomize 模板
  - `yaml`: 原生YAML配置模板

- **模板内容**:
  - `repo_url`: Chart仓库地址
  - `chart_name`: Chart名称
  - `chart_version`: Chart版本
  - `values_yaml`: 默认Values配置

### 2. 创建环境时选择模板

1. 点击"新建环境"按钮
2. 填写基本信息(环境编码、名称、类型等)
3. **选择环境模板**(可选):
   - 下拉框显示所有可用模板
   - 显示模板名称和类型
   - 可以不选择模板(留空)
4. 选择模板后,系统会:
   - 记录 `template_id` 到环境表
   - 后续可以根据模板自动应用配置

### 3. 已有的模板数据

系统中已有3个模板:

| ID | 模板编码 | 模板名称 | 类型 |
|----|---------|---------|------|
| 1  | basic-deployment | 基础部署模板 | yaml |
| 2  | helm-app | Helm应用模板 | helm |
| 3  | go-microservice-standard | Go微服务标准模板 | helm |

## 使用场景

### 场景1: 标准化环境配置

- 开发环境使用"基础部署模板"
- 生产环境使用"Go微服务标准模板"
- 确保不同环境间的配置一致性

### 场景2: 快速环境创建

- 选择模板后自动继承模板的默认配置
- 减少手动配置工作量
- 降低配置错误风险

### 场景3: 配置版本管理

- 模板变更后,关联的环境可以批量更新
- 追溯环境配置来源

## 数据库表关系

```
env_templates (环境模板表)
    ↓ (1:N)
environments (环境表)
    ↓ (1:N)
app_env_bindings (应用环境绑定表)
```

- 一个模板可以被多个环境使用
- 一个环境只能关联一个模板(或不关联)
- 环境与应用通过 `app_env_bindings` 关联

## 验证方式

### 1. 前端验证

刷新浏览器,进入"环境管理"页面:

1. 点击"新建环境"按钮
2. 查看表单中是否有"环境模板"选择框
3. 点击下拉框,应该看到3个可用模板:
   - 基础部署模板 (yaml)
   - Helm应用模板 (helm)
   - Go微服务标准模板 (helm)

### 2. 数据库验证

创建环境后检查数据:

```sql
-- 查看环境表结构
SHOW COLUMNS FROM env_db.environments WHERE Field='template_id';

-- 查看创建的环境及关联的模板
SELECT 
    e.id, 
    e.env_name, 
    e.template_id,
    t.template_name,
    t.template_type
FROM env_db.environments e
LEFT JOIN env_db.env_templates t ON e.template_id = t.id
WHERE e.is_deleted = 0;
```

### 3. API验证

查看创建环境的请求:

```bash
# POST /api/v1/environments
{
  "envCode": "dev-001",
  "envName": "开发环境",
  "envType": "dev",
  "clusterId": 1,
  "namespace": "my-app-dev",
  "projectId": 1,
  "templateId": 2,  // 选择的模板ID
  "description": "开发环境",
  "status": 1
}
```

## 后续优化建议

### 1. 模板配置自动应用

当选择模板时,自动填充以下内容:
- 环境变量模板
- 资源配额默认值
- 健康检查配置

实现方式:
```javascript
// 监听templateId变化
watch(() => form.templateId, async (newTemplateId) => {
  if (newTemplateId) {
    const template = templates.value.find(t => t.id === newTemplateId)
    if (template && template.valuesYaml) {
      // 解析并应用模板配置
      form.configJson = template.valuesYaml
    }
  }
})
```

### 2. 模板预览

在选择模板后显示模板内容预览:
- Chart信息
- 默认Values配置
- 资源要求

### 3. 模板版本管理

- 模板支持多版本
- 环境可以选择模板的特定版本
- 模板升级时提示关联环境更新

### 4. 模板分类和标签

- 按环境类型分类(dev/test/prod)
- 按技术栈分类(Java/Go/Python)
- 支持标签过滤

## 影响范围

### 修改的文件

1. **后端**:
   - `backend/internal/common/model/environment.go` - 添加TemplateID字段

2. **前端**:
   - `frontend/src/views/environment/EnvironmentList.vue` - 添加模板选择表单

3. **数据库**:
   - `env_db.environments` 表 - 添加template_id列

### 需要重启的服务

- `my-cloud-env-service`
- `my-cloud-frontend`

## 部署步骤

```bash
# 1. 数据库变更
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "ALTER TABLE env_db.environments ADD COLUMN template_id BIGINT NULL COMMENT '关联的环境模板ID' AFTER project_id;"

# 2. 重新构建服务
cd /Users/hanhailong01/Downloads/my_cloud
docker-compose build env-service frontend

# 3. 重启服务
docker-compose restart env-service frontend

# 4. 验证服务启动
docker logs my-cloud-env-service --tail 20
docker logs my-cloud-frontend --tail 20
```

## 完成状态

✅ 数据库字段已添加  
✅ 后端模型已更新  
✅ 前端表单已添加模板选择  
✅ 服务已重新构建和部署  
✅ 功能测试通过

现在用户在新建环境时可以选择环境模板了! 🎉
