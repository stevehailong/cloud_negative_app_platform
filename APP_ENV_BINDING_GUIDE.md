# 应用环境绑定功能使用指南

## 功能概述

应用环境绑定功能允许将应用与特定环境进行绑定，实现应用的环境隔离部署。绑定后，应用部署时将自动使用绑定环境的命名空间和集群配置。

## 功能特性

### 1. 环境绑定管理
- ✅ 查看应用已绑定的环境列表
- ✅ 绑定新环境到应用
- ✅ 配置环境专属的资源限制（副本数、CPU、内存）
- ✅ 解绑环境
- ✅ 查看绑定状态（已配置/待配置）

### 2. 自动关联环境信息
- ✅ 显示环境名称和类型
- ✅ 显示命名空间
- ✅ 显示集群名称
- ✅ 环境选择时实时显示详细信息

## 使用流程

### 前置条件
1. 确保已创建应用（应用管理）
2. 确保已创建环境（环境管理）
3. 环境需要关联到具体的集群和命名空间

### 绑定环境操作步骤

#### 1. 进入应用详情页
- 访问：应用管理 → 选择应用 → 详情
- 页面URL格式：`/applications/{应用ID}`

#### 2. 绑定环境
1. 在"环境绑定"卡片中，点击【绑定环境】按钮
2. 在弹出的对话框中：
   - **选择环境**：从下拉框中选择要绑定的环境
     - 显示格式：环境名称 (环境类型)
     - 支持搜索筛选
   - **配置资源**：
     - 副本数：默认1，范围1-100
     - CPU请求：默认100m
     - CPU限制：默认500m
     - 内存请求：默认128Mi
     - 内存限制：默认512Mi
   - **预览信息**：选择环境后会显示：
     - 所属集群
     - 命名空间
3. 点击【确定】完成绑定

#### 3. 查看绑定列表
绑定成功后，环境会显示在"环境绑定"表格中，包含以下信息：
- 环境名称
- 环境类型（如：dev/test/prod）
- 命名空间
- 集群名称
- 配置状态（已配置/待配置）
- 创建时间
- 操作按钮：配置、解绑

#### 4. 配置环境
点击【配置】按钮可以跳转到环境配置页面，配置：
- ConfigMaps（配置文件）
- Secrets（敏感信息）
- 其他环境变量

#### 5. 解绑环境
点击【解绑】按钮，确认后即可解除应用与环境的绑定关系。

## API接口说明

### 1. 查询绑定列表
```
GET /api/v1/app-env-bindings?applicationId={应用ID}
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "appId": 1,
        "envId": 1,
        "envName": "开发环境",
        "envType": "dev",
        "namespace": "app-dev",
        "clusterName": "k8s-dev-cluster",
        "replicas": 1,
        "cpuRequest": "100m",
        "cpuLimit": "500m",
        "memoryRequest": "128Mi",
        "memoryLimit": "512Mi",
        "configStatus": "ready",
        "createTime": "2026-06-01T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 100
  }
}
```

### 2. 创建绑定
```
POST /api/v1/app-env-bindings
Content-Type: application/json

{
  "appId": 1,
  "envId": 1,
  "replicas": 1,
  "cpuRequest": "100m",
  "cpuLimit": "500m",
  "memoryRequest": "128Mi",
  "memoryLimit": "512Mi",
  "configJson": "{}"
}
```

### 3. 删除绑定
```
DELETE /api/v1/app-env-bindings/{绑定ID}
```

## 数据库表结构

### app_env_bindings 表
```sql
CREATE TABLE `app_env_bindings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` bigint unsigned NOT NULL COMMENT '应用ID',
  `env_id` bigint unsigned NOT NULL COMMENT '环境ID',
  `template_id` bigint unsigned DEFAULT NULL COMMENT '模板ID',
  `replicas` int DEFAULT '1' COMMENT '副本数',
  `cpu_request` varchar(32) DEFAULT '100m' COMMENT 'CPU请求',
  `cpu_limit` varchar(32) DEFAULT '500m' COMMENT 'CPU限制',
  `memory_request` varchar(32) DEFAULT '128Mi' COMMENT '内存请求',
  `memory_limit` varchar(32) DEFAULT '512Mi' COMMENT '内存限制',
  `config_json` json DEFAULT NULL COMMENT '配置JSON',
  `status` tinyint DEFAULT '1' COMMENT '状态',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `create_by` bigint unsigned DEFAULT NULL COMMENT '创建人',
  `update_by` bigint unsigned DEFAULT NULL COMMENT '更新人',
  `is_deleted` tinyint DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_env_id` (`env_id`),
  UNIQUE KEY `uk_app_env` (`app_id`, `env_id`)
) COMMENT='应用环境绑定表';
```

## 与部署流程的集成

### 应用部署时选择环境
当应用绑定环境后，在部署时：

1. **自动获取配置**：
   - 根据`envId`自动查询环境信息
   - 获取环境的`namespace`和`clusterId`
   - 使用绑定配置的资源限制

2. **部署服务调用**：
   ```go
   // AppDeploymentService.CreateAppDeployment
   // 只需传递 appID, envID, workloadName
   // 自动从Environment表查询namespace和clusterId
   deployment, err := service.CreateAppDeployment(
       appID,
       envID,
       workloadName,
       workloadType,
       desiredReplicas,
   )
   ```

3. **命名空间隔离**：
   - 每个环境有独立的namespace
   - 不同环境的应用部署互不干扰
   - 符合企业级多租户隔离要求

## 测试用例

### 测试场景1：绑定新环境
1. 前提：已有应用ID=1，环境ID=1（开发环境）
2. 操作：在应用详情页点击"绑定环境"，选择开发环境
3. 预期：
   - 绑定成功提示
   - 环境列表中显示新绑定的环境
   - 显示正确的环境名称、命名空间、集群名称

### 测试场景2：重复绑定检查
1. 前提：应用已绑定环境ID=1
2. 操作：再次尝试绑定环境ID=1
3. 预期：提示"该应用已绑定此环境"

### 测试场景3：查看绑定列表
1. 前提：应用已绑定多个环境
2. 操作：打开应用详情页
3. 预期：
   - 显示所有已绑定的环境
   - 每个环境显示完整信息（环境名、类型、命名空间、集群）
   - 配置状态正确显示

### 测试场景4：解绑环境
1. 前提：应用已绑定环境
2. 操作：点击"解绑"按钮并确认
3. 预期：
   - 解绑成功提示
   - 环境从列表中移除

## 注意事项

1. **唯一性约束**：
   - 一个应用不能重复绑定同一个环境
   - 数据库有`uk_app_env`唯一索引保证

2. **环境命名空间**：
   - 命名空间在环境创建时指定
   - 符合K8s命名规范（1-63字符，小写字母、数字、连字符）
   - 同一集群的命名空间不能重复

3. **资源配置**：
   - CPU格式：支持`100m`（毫核）或`1`（核）
   - 内存格式：支持`128Mi`或`1Gi`
   - 建议CPU limit >= request，Memory limit >= request

4. **部署依赖**：
   - 部署前必须先绑定环境
   - 未绑定环境的应用无法部署
   - 环境删除会影响已绑定的应用部署

## 下一步优化建议

1. **批量绑定**：支持一次绑定多个环境
2. **环境模板**：使用环境模板快速配置资源限制
3. **权限控制**：限制只有应用负责人可以绑定/解绑环境
4. **部署限制**：在应用部署页面，下拉框只显示已绑定的环境
5. **配置继承**：支持从其他环境复制配置
6. **环境推广**：支持从dev环境推广到test/prod环境

## 相关文档

- [命名空间隔离设计](./NAMESPACE_ISOLATION_DESIGN.md)
- [环境管理指南](./ENVIRONMENT_DROPDOWN_FIX.md)
- [快速开始](./QUICK_START.md)
