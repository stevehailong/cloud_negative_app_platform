-- 权限管理扩展SQL
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
USE iam_db;

-- 1. 添加permissions表（如果不存在）
CREATE TABLE IF NOT EXISTS permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(128) NOT NULL UNIQUE COMMENT '权限编码',
    name VARCHAR(128) NOT NULL COMMENT '权限名称',
    resource_type VARCHAR(32) NULL COMMENT '资源类型',
    http_method VARCHAR(16) NULL COMMENT 'HTTP方法',
    path VARCHAR(255) NULL COMMENT 'API路径',
    description VARCHAR(255) NULL COMMENT '描述',
    status TINYINT DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    INDEX idx_code (code),
    INDEX idx_resource_type (resource_type),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 2. 添加role_permissions关联表（如果不存在）
CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    permission_id BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_permission (role_id, permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- 3. 添加访客角色
INSERT INTO roles (name, code, description, created_by, status, sort) VALUES
('访客', 'GUEST', '只读权限，新用户默认角色', 'system', 1, 100)
ON DUPLICATE KEY UPDATE name='访客';

-- 4. 插入默认权限
INSERT INTO permissions (code, name, resource_type, http_method, path, description, created_by, status) VALUES

-- 应用管理权限
('application:view', '查看应用', 'application', 'GET', '/api/v1/applications*', '查看应用列表和详情', 'system', 1),
('application:create', '创建应用', 'application', 'POST', '/api/v1/applications*', '创建新应用', 'system', 1),
('application:update', '编辑应用', 'application', 'PUT', '/api/v1/applications*', '修改应用信息', 'system', 1),
('application:delete', '删除应用', 'application', 'DELETE', '/api/v1/applications*', '删除应用', 'system', 1),

-- 组件管理权限
('component:view', '查看组件', 'component', 'GET', '/api/v1/components*', '查看组件信息', 'system', 1),
('component:manage', '管理组件', 'component', 'POST,PUT,DELETE', '/api/v1/components*', '创建、编辑、删除组件', 'system', 1),

-- 用户管理权限
('user:view', '查看用户', 'user', 'GET', '/api/v1/users*', '查看用户列表', 'system', 1),
('user:create', '创建用户', 'user', 'POST', '/api/v1/users', '创建新用户', 'system', 1),
('user:edit', '编辑用户', 'user', 'PUT', '/api/v1/users*', '编辑用户信息', 'system', 1),
('user:delete', '删除用户', 'user', 'DELETE', '/api/v1/users*', '删除用户', 'system', 1),
('user:assign_role', '分配角色', 'user', 'POST', '/api/v1/users/assign-roles*', '为用户分配角色', 'system', 1),
('user:change_status', '修改用户状态', 'user', 'PUT', '/api/v1/users/*/status*', '启用或禁用用户', 'system', 1),

-- 角色管理权限
('role:view', '查看角色', 'role', 'GET', '/api/v1/roles*', '查看角色列表', 'system', 1),
('role:manage', '管理角色', 'role', 'POST,PUT,DELETE', '/api/v1/roles*', '创建、编辑、删除角色', 'system', 1),
('role:assign_perm', '分配权限', 'role', 'POST', '/api/v1/roles/*/permissions*', '为角色分配权限', 'system', 1),

-- 权限管理权限
('permission:view', '查看权限', 'permission', 'GET', '/api/v1/permissions*', '查看权限列表', 'system', 1),
('permission:manage', '管理权限', 'permission', 'POST,PUT,DELETE', '/api/v1/permissions*', '创建、编辑、删除权限', 'system', 1),

-- 租户管理权限
('tenant:view', '查看租户', 'tenant', 'GET', '/api/v1/tenants*', '查看租户列表', 'system', 1),
('tenant:manage', '管理租户', 'tenant', 'POST,PUT,DELETE', '/api/v1/tenants*', '创建、编辑、删除租户', 'system', 1),

-- 组织管理权限
('organization:view', '查看组织', 'organization', 'GET', '/api/v1/organizations*', '查看组织架构', 'system', 1),
('organization:manage', '管理组织', 'organization', 'POST,PUT,DELETE', '/api/v1/organizations*', '创建、编辑、删除组织', 'system', 1),

-- 项目管理权限
('project:view', '查看项目', 'project', 'GET', '/api/v1/projects*', '查看项目信息', 'system', 1),
('project:manage', '管理项目', 'project', 'POST,PUT,DELETE', '/api/v1/projects*', '创建、编辑、删除项目', 'system', 1),

-- 项目成员权限
('project-member:view', '查看项目成员', 'project', 'GET', '/api/v1/project-members*', '查看项目成员', 'system', 1),
('project-member:manage', '管理项目成员', 'project', 'POST,DELETE', '/api/v1/project-members*', '管理项目成员', 'system', 1),

-- 部署权限
('deploy:view', '查看部署', 'deployment', 'GET', '/api/v1/deployments*', '查看部署记录', 'system', 1),
('deploy:execute', '执行部署', 'deployment', 'POST', '/api/v1/deployments*', '创建和执行部署', 'system', 1),
('deploy:rollback', '回滚部署', 'deployment', 'POST', '/api/v1/deployments/*/rollback*', '回滚到历史版本', 'system', 1),
('deploy:delete', '删除部署', 'deployment', 'DELETE', '/api/v1/deployments*', '删除部署记录', 'system', 1),

-- 应用部署权限（新版）
('app_deployment:view', '查看应用部署', 'deployment', 'GET', '/api/v1/app-deployments*', '查看应用部署状态', 'system', 1),
('app_deployment:deploy', '部署新版本', 'deployment', 'POST', '/api/v1/app-deployments/*/deploy*', '部署新版本到环境', 'system', 1),
('app_deployment:scale', '扩缩容', 'deployment', 'POST', '/api/v1/app-deployments/*/scale*', '调整副本数', 'system', 1),
('app_deployment:restart', '重启部署', 'deployment', 'POST', '/api/v1/app-deployments/*/restart*', '重启应用实例', 'system', 1),
('app_deployment:rollback', '回滚应用部署', 'deployment', 'POST', '/api/v1/app-deployments/*/rollback*', '回滚到历史版本', 'system', 1),

-- 流水线权限
('pipeline:view', '查看流水线', 'pipeline', 'GET', '/api/v1/pipelines*', '查看流水线', 'system', 1),
('pipeline:create', '创建流水线', 'pipeline', 'POST', '/api/v1/pipelines*', '创建流水线', 'system', 1),
('pipeline:edit', '编辑流水线', 'pipeline', 'PUT', '/api/v1/pipelines*', '修改流水线配置', 'system', 1),
('pipeline:execute', '执行流水线', 'pipeline', 'POST', '/api/v1/pipelines/*/run*', '触发流水线执行', 'system', 1),
('pipeline:delete', '删除流水线', 'pipeline', 'DELETE', '/api/v1/pipelines*', '删除流水线', 'system', 1),

-- 流水线执行权限
('pipeline_run:view', '查看流水线执行', 'pipeline', 'GET', '/api/v1/pipeline-runs*', '查看流水线执行记录', 'system', 1),
('pipeline_run:create', '创建流水线执行', 'pipeline', 'POST', '/api/v1/pipeline-runs*', '触发流水线执行', 'system', 1),
('pipeline_run:cancel', '取消流水线执行', 'pipeline', 'POST', '/api/v1/pipeline-runs/*/cancel*', '取消流水线执行', 'system', 1),
('pipeline_run:rerun', '重跑流水线', 'pipeline', 'POST', '/api/v1/pipeline-runs/*/rerun*', '重跑流水线', 'system', 1),

-- 制品管理权限
('artifact:view', '查看制品', 'artifact', 'GET', '/api/v1/artifacts*', '查看构建制品', 'system', 1),
('artifact:manage', '管理制品', 'artifact', 'POST,PUT,DELETE', '/api/v1/artifacts*', '上传、删除制品', 'system', 1),

-- 环境管理权限
('env:view', '查看环境', 'environment', 'GET', '/api/v1/environments*', '查看环境信息', 'system', 1),
('env:manage', '管理环境', 'environment', 'POST,PUT,DELETE', '/api/v1/environments*', '创建、编辑、删除环境', 'system', 1),

-- 环境模板权限
('env-template:view', '环境模板查看', 'environment', 'GET', '/api/v1/env-templates*', '查看环境模板', 'system', 1),
('env-template:manage', '环境模板管理', 'environment', 'POST,PUT,DELETE', '/api/v1/env-templates*', '管理环境模板', 'system', 1),

-- 应用环境绑定权限
('app-env-binding:view', '应用环境绑定查看', 'environment', 'GET', '/api/v1/app-env-bindings*', '查看应用环境绑定', 'system', 1),
('app-env-binding:manage', '应用环境绑定管理', 'environment', 'POST,PUT,DELETE', '/api/v1/app-env-bindings*', '管理应用环境绑定', 'system', 1),

-- ConfigMap权限
('configmap:view', '查看ConfigMap', 'environment', 'GET', '/api/v1/config-maps*', '查看ConfigMap', 'system', 1),
('configmap:manage', '管理ConfigMap', 'environment', 'POST,PUT,DELETE', '/api/v1/config-maps*', '管理ConfigMap', 'system', 1),

-- Secret权限
('secret:view', '查看Secret', 'environment', 'GET', '/api/v1/secrets*', '查看Secret', 'system', 1),
('secret:manage', '管理Secret', 'environment', 'POST,PUT,DELETE', '/api/v1/secrets*', '管理Secret', 'system', 1),

-- 发布管理权限
('release:view', '查看发布', 'release', 'GET', '/api/v1/releases*', '查看发布工单', 'system', 1),
('release:create', '创建发布', 'release', 'POST', '/api/v1/releases*', '创建发布工单', 'system', 1),
('release:edit', '编辑发布', 'release', 'PUT', '/api/v1/releases*', '修改发布工单', 'system', 1),
('release:submit', '提交审批', 'release', 'POST', '/api/v1/releases/*/submit*', '提交发布工单审批', 'system', 1),
('release:execute', '执行发布', 'release', 'POST', '/api/v1/releases/*/execute*', '执行发布部署', 'system', 1),
('release:approve', '审批发布', 'release', 'POST', '/api/v1/releases/*/approve*', '审批通过发布工单', 'system', 1),
('release:reject', '拒绝发布', 'release', 'POST', '/api/v1/releases/*/reject*', '拒绝发布工单', 'system', 1),
('release:rollback', '回滚发布', 'release', 'POST', '/api/v1/releases/*/rollback*', '回滚发布', 'system', 1),
('release:canary', '金丝雀操作', 'release', 'POST', '/api/v1/releases/*/canary*', '确认或回滚金丝雀', 'system', 1),
('release:delete', '删除发布', 'release', 'DELETE', '/api/v1/releases*', '删除发布工单', 'system', 1),

-- 集群管理权限
('cluster:view', '查看集群', 'cluster', 'GET', '/api/v1/clusters*', '查看集群信息', 'system', 1),
('cluster:manage', '管理集群', 'cluster', 'POST,PUT,DELETE', '/api/v1/clusters*', '创建、编辑、删除集群', 'system', 1),

-- 节点管理权限
('node:view', '节点查看', 'cluster', 'GET', '/api/v1/nodes*', '查看节点信息', 'system', 1),
('node:manage', '节点管理', 'cluster', 'POST,PUT,DELETE', '/api/v1/nodes*', '管理节点', 'system', 1),

-- 命名空间管理权限
('namespace:view', '命名空间查看', 'cluster', 'GET', '/api/v1/namespaces*', '查看命名空间', 'system', 1),
('namespace:manage', '命名空间管理', 'cluster', 'POST,PUT,DELETE', '/api/v1/namespaces*', '管理命名空间', 'system', 1),

-- 资源管理权限
('resource:view', '查看资源', 'resource', 'GET', '/api/v1/resources*', '查看K8s资源', 'system', 1),
('resource:manage', '管理资源', 'resource', 'POST,PUT,DELETE', '/api/v1/resources*', '管理K8s资源', 'system', 1),

-- 监控权限
('monitor:view', '查看监控', 'monitor', 'GET', '/api/v1/monitors*', '查看监控数据', 'system', 1),
('monitor:manage', '管理监控', 'monitor', 'POST,PUT,DELETE', '/api/v1/monitors*', '配置监控规则', 'system', 1),

-- 监控指标
('metric:view', '查看指标', 'monitor', 'GET', '/api/v1/metrics*', '查看监控指标', 'system', 1),

-- 告警规则
('alert_rule:view', '查看告警规则', 'monitor', 'GET', '/api/v1/alert-rules*', '查看告警规则', 'system', 1),
('alert_rule:manage', '管理告警规则', 'monitor', 'POST,PUT,DELETE', '/api/v1/alert-rules*', '管理告警规则', 'system', 1),

-- 告警记录
('alert:view', '查看告警', 'monitor', 'GET', '/api/v1/alerts*', '查看告警记录', 'system', 1),
('alert:manage', '管理告警', 'monitor', 'POST,PUT,DELETE', '/api/v1/alerts*', '处理告警', 'system', 1),

-- 审计日志
('audit:view', '查看审计日志', 'audit', 'GET', '/api/v1/audit-logs*', '查看审计日志', 'system', 1),
('audit:export', '导出审计日志', 'audit', 'POST', '/api/v1/audit-logs/export*', '导出审计日志', 'system', 1),

-- 通知管理
('notification:view', '查看通知', 'notification', 'GET', '/api/v1/notifications*', '查看通知记录', 'system', 1),
('notification:send', '发送通知', 'notification', 'POST', '/api/v1/notifications*', '发送通知', 'system', 1),

-- 通知模板
('notification_template:view', '查看通知模板', 'notification', 'GET', '/api/v1/notification-templates*', '查看通知模板', 'system', 1),
('notification_template:manage', '管理通知模板', 'notification', 'POST,PUT,DELETE', '/api/v1/notification-templates*', '管理通知模板', 'system', 1),

-- 通知渠道
('notification_channel:view', '查看通知渠道', 'notification', 'GET', '/api/v1/notification-channels*', '查看通知渠道', 'system', 1),
('notification_channel:manage', '管理通知渠道', 'notification', 'POST,PUT,DELETE', '/api/v1/notification-channels*', '管理通知渠道', 'system', 1),

-- GitLab集成
('gitlab:projects', '查看GitLab项目', 'gitlab', 'GET', '/api/v1/gitlab/projects*', '查看GitLab项目', 'system', 1),
('gitlab:test', '测试GitLab连接', 'gitlab', 'POST', '/api/v1/gitlab/test*', '测试GitLab连接', 'system', 1),
('gitlab:webhooks', '管理GitLab Webhooks', 'gitlab', 'POST', '/api/v1/gitlab/webhooks*', '管理GitLab Webhooks', 'system', 1),
('gitlab:config', '配置GitLab客户端', 'gitlab', 'PUT', '/api/v1/gitlab/client*', '配置GitLab客户端', 'system', 1),

-- 系统设置
('settings:read', '查看系统设置', 'settings', 'GET', '/api/v1/settings*', '查看系统设置', 'system', 1),
('settings:write', '修改系统设置', 'settings', 'PUT', '/api/v1/settings*', '修改系统设置', 'system', 1),

-- 文件上传
('upload:file', '上传文件', 'upload', 'POST', '/api/v1/upload*', '上传文件', 'system', 1),

-- 成本治理
('cost:view', '查看成本', 'cost', 'GET', '/api/v1/costs*', '查看成本分析', 'system', 1),
('cost:manage', '管理成本', 'cost', 'POST,PUT,DELETE', '/api/v1/costs*', '配置成本策略', 'system', 1)

ON DUPLICATE KEY UPDATE name=VALUES(name);

-- ========================================
-- 5. 为角色分配权限
-- ========================================

-- 访客角色(GUEST)：只读权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'GUEST' AND p.code IN (
    'application:view', 'component:view', 'user:view', 'role:view', 'permission:view',
    'tenant:view', 'organization:view', 'project:view', 'project-member:view',
    'deploy:view', 'app_deployment:view',
    'pipeline:view', 'pipeline_run:view', 'artifact:view',
    'env:view', 'env-template:view', 'app-env-binding:view', 'configmap:view', 'secret:view',
    'release:view',
    'cluster:view', 'node:view', 'namespace:view', 'resource:view',
    'monitor:view', 'metric:view', 'alert_rule:view', 'alert:view',
    'audit:view',
    'notification:view', 'notification_template:view', 'notification_channel:view',
    'gitlab:projects',
    'cost:view'
);

-- 开发人员(DEVELOPER)：应用、组件、流水线的完整权限 + 其他只读
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'DEVELOPER' AND p.code IN (
    'application:view', 'application:create', 'application:update', 'application:delete',
    'component:view', 'component:manage',
    'pipeline:view', 'pipeline:create', 'pipeline:edit', 'pipeline:execute', 'pipeline:delete',
    'pipeline_run:view', 'pipeline_run:create',
    'artifact:view',
    'release:view', 'release:create',
    'tenant:view', 'organization:view', 'project:view', 'project-member:view',
    'deploy:view', 'app_deployment:view', 'app_deployment:deploy', 'app_deployment:restart',
    'env:view', 'env-template:view', 'app-env-binding:view',
    'cluster:view', 'node:view', 'namespace:view', 'resource:view',
    'monitor:view', 'metric:view', 'alert_rule:view', 'alert:view',
    'audit:view',
    'notification:view', 'notification_template:view', 'notification_channel:view',
    'gitlab:projects', 'gitlab:test',
    'settings:read',
    'upload:file',
    'cost:view',
    'user:view', 'role:view'
);

-- 运维人员(OPS)：部署、环境、集群、发布的完整权限 + 监控管理 + 其他只读
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'OPS' AND p.code IN (
    'deploy:view', 'deploy:execute', 'deploy:rollback', 'deploy:delete',
    'app_deployment:view', 'app_deployment:deploy', 'app_deployment:scale', 'app_deployment:restart', 'app_deployment:rollback',
    'env:view', 'env:manage',
    'env-template:view', 'env-template:manage',
    'app-env-binding:view', 'app-env-binding:manage',
    'configmap:view', 'configmap:manage',
    'secret:view', 'secret:manage',
    'cluster:view', 'cluster:manage',
    'node:view', 'node:manage',
    'namespace:view', 'namespace:manage',
    'resource:view', 'resource:manage',
    'monitor:view', 'monitor:manage',
    'metric:view',
    'alert_rule:view', 'alert_rule:manage',
    'alert:view', 'alert:manage',
    'release:view', 'release:create', 'release:edit', 'release:execute', 'release:approve', 'release:canary', 'release:delete',
    'pipeline:view', 'pipeline_run:view',
    'tenant:view', 'organization:view', 'project:view',
    'application:view', 'component:view', 'artifact:view',
    'audit:view',
    'notification:view', 'notification_template:view', 'notification_channel:view',
    'gitlab:projects',
    'settings:read', 'settings:write',
    'upload:file',
    'cost:view',
    'user:view', 'role:view'
);

-- 项目管理员(PROJECT_ADMIN)：除用户/角色/权限管理外的所有权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'PROJECT_ADMIN' AND p.code NOT IN (
    'user:create', 'user:edit', 'user:delete', 'user:assign_role', 'user:change_status',
    'role:manage', 'role:assign_perm',
    'permission:manage'
);

-- 超级管理员(SUPER_ADMIN)：所有权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'SUPER_ADMIN';

-- 6. 为现有未分配角色的用户分配访客角色
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, (SELECT id FROM roles WHERE code = 'GUEST' LIMIT 1) FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.id IS NULL;