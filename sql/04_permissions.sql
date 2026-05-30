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
('app:view', '查看应用', 'application', 'GET', '/api/v1/applications/*', '查看应用列表和详情', 'system', 1),
('app:create', '创建应用', 'application', 'POST', '/api/v1/applications/', '创建新应用', 'system', 1),
('app:edit', '编辑应用', 'application', 'PUT', '/api/v1/applications/*', '修改应用信息', 'system', 1),
('app:delete', '删除应用', 'application', 'DELETE', '/api/v1/applications/*', '删除应用', 'system', 1),

-- 组件管理权限
('component:view', '查看组件', 'component', 'GET', '/api/v1/components/*', '查看组件信息', 'system', 1),
('component:manage', '管理组件', 'component', 'POST,PUT,DELETE', '/api/v1/components/*', '创建、编辑、删除组件', 'system', 1),

-- 用户管理权限
('user:view', '查看用户', 'user', 'GET', '/api/v1/users/*', '查看用户列表', 'system', 1),
('user:create', '创建用户', 'user', 'POST', '/api/v1/users/', '创建新用户', 'system', 1),
('user:edit', '编辑用户', 'user', 'PUT', '/api/v1/users/*', '编辑用户信息', 'system', 1),
('user:delete', '删除用户', 'user', 'DELETE', '/api/v1/users/*', '删除用户', 'system', 1),
('user:assign_role', '分配角色', 'user', 'POST', '/api/v1/users/assign-roles/', '为用户分配角色', 'system', 1),
('user:change_status', '修改用户状态', 'user', 'PUT', '/api/v1/users/*/status/', '启用或禁用用户', 'system', 1),

-- 角色管理权限
('role:view', '查看角色', 'role', 'GET', '/api/v1/roles/*', '查看角色列表', 'system', 1),
('role:manage', '管理角色', 'role', 'POST,PUT,DELETE', '/api/v1/roles/*', '创建、编辑、删除角色', 'system', 1),
('role:assign_perm', '分配权限', 'role', 'POST', '/api/v1/roles/*/permissions/', '为角色分配权限', 'system', 1),

-- 权限管理权限
('permission:view', '查看权限', 'permission', 'GET', '/api/v1/permissions/*', '查看权限列表', 'system', 1),
('permission:manage', '管理权限', 'permission', 'POST,PUT,DELETE', '/api/v1/permissions/*', '创建、编辑、删除权限', 'system', 1),

-- 部署权限
('deploy:view', '查看部署', 'deployment', 'GET', '/api/v1/deployments/*', '查看部署记录', 'system', 1),
('deploy:execute', '执行部署', 'deployment', 'POST', '/api/v1/deployments/', '创建和执行部署', 'system', 1),
('deploy:rollback', '回滚部署', 'deployment', 'POST', '/api/v1/deployments/*/rollback/', '回滚到历史版本', 'system', 1),
('deploy:delete', '删除部署', 'deployment', 'DELETE', '/api/v1/deployments/*', '删除部署记录', 'system', 1),

-- 流水线权限
('pipeline:view', '查看流水线', 'pipeline', 'GET', '/api/v1/pipelines/*', '查看流水线', 'system', 1),
('pipeline:create', '创建流水线', 'pipeline', 'POST', '/api/v1/pipelines/', '创建流水线', 'system', 1),
('pipeline:edit', '编辑流水线', 'pipeline', 'PUT', '/api/v1/pipelines/*', '修改流水线配置', 'system', 1),
('pipeline:execute', '执行流水线', 'pipeline', 'POST', '/api/v1/pipelines/*/run/', '触发流水线执行', 'system', 1),
('pipeline:delete', '删除流水线', 'pipeline', 'DELETE', '/api/v1/pipelines/*', '删除流水线', 'system', 1),

-- 环境管理权限
('env:view', '查看环境', 'environment', 'GET', '/api/v1/environments/*', '查看环境信息', 'system', 1),
('env:manage', '管理环境', 'environment', 'POST,PUT,DELETE', '/api/v1/environments/*', '创建、编辑、删除环境', 'system', 1),

-- 集群管理权限
('cluster:view', '查看集群', 'cluster', 'GET', '/api/v1/clusters/*', '查看集群信息', 'system', 1),
('cluster:manage', '管理集群', 'cluster', 'POST,PUT,DELETE', '/api/v1/clusters/*', '创建、编辑、删除集群', 'system', 1),

-- 项目管理权限
('project:view', '查看项目', 'project', 'GET', '/api/v1/projects/*', '查看项目信息', 'system', 1),
('project:manage', '管理项目', 'project', 'POST,PUT,DELETE', '/api/v1/projects/*', '创建、编辑、删除项目', 'system', 1),

-- 监控权限
('monitor:view', '查看监控', 'monitor', 'GET', '/api/v1/monitors/*', '查看监控数据', 'system', 1),
('monitor:manage', '管理监控', 'monitor', 'POST,PUT,DELETE', '/api/v1/monitors/*', '配置监控规则', 'system', 1),

-- 发布管理权限
('release:view', '查看发布', 'release', 'GET', '/api/v1/releases/*', '查看发布工单', 'system', 1),
('release:create', '创建发布', 'release', 'POST', '/api/v1/releases/', '创建发布工单', 'system', 1),
('release:edit', '编辑发布', 'release', 'PUT', '/api/v1/releases/*', '修改发布工单', 'system', 1),
('release:submit', '提交审批', 'release', 'POST', '/api/v1/releases/*/submit/', '提交发布工单审批', 'system', 1),
('release:execute', '执行发布', 'release', 'POST', '/api/v1/releases/*/execute/', '执行发布部署', 'system', 1),
('release:approve', '审批发布', 'release', 'POST', '/api/v1/releases/*/approve/', '审批通过发布工单', 'system', 1),
('release:reject', '拒绝发布', 'release', 'POST', '/api/v1/releases/*/reject/', '拒绝发布工单', 'system', 1),
('release:rollback', '回滚发布', 'release', 'POST', '/api/v1/releases/*/rollback/', '回滚发布', 'system', 1),
('release:canary', '金丝雀操作', 'release', 'POST', '/api/v1/releases/*/canary/*', '确认或回滚金丝雀', 'system', 1),
('release:delete', '删除发布', 'release', 'DELETE', '/api/v1/releases/*', '删除发布工单', 'system', 1)

ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 5. 为角色分配权限
-- 注意：角色ID可能因自增而不同，以下使用子查询按code匹配

-- 访客角色(GUEST)：只读权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'GUEST' AND p.code IN (
    'app:view', 'component:view', 'user:view', 'role:view', 'permission:view',
    'deploy:view', 'pipeline:view', 'env:view', 'cluster:view', 'project:view', 'monitor:view', 'release:view'
);

-- 开发人员(DEVELOPER)：应用、组件、流水线的完整权限 + 发布查看/创建 + 其他只读
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'DEVELOPER' AND p.code IN (
    'app:view', 'app:create', 'app:edit', 'app:delete',
    'component:view', 'component:manage',
    'pipeline:view', 'pipeline:create', 'pipeline:edit', 'pipeline:execute', 'pipeline:delete',
    'release:view', 'release:create',
    'deploy:view', 'env:view', 'cluster:view', 'project:view', 'monitor:view',
    'user:view', 'role:view'
);

-- 运维人员(OPS)：部署、环境、集群、发布的完整权限 + 其他只读
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'OPS' AND p.code IN (
    'deploy:view', 'deploy:execute', 'deploy:rollback', 'deploy:delete',
    'env:view', 'env:manage',
    'cluster:view', 'cluster:manage',
    'monitor:view', 'monitor:manage',
    'release:view', 'release:create', 'release:edit', 'release:execute', 'release:approve', 'release:canary', 'release:delete',
    'app:view', 'component:view', 'pipeline:view', 'project:view',
    'user:view', 'role:view'
);

-- 项目管理员(PROJECT_ADMIN)：除用户和角色管理外的所有权限
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
SELECT u.id, 5 FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.id IS NULL;
