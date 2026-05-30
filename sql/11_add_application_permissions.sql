-- 添加应用相关权限
USE iam_db;

-- 应用管理权限
INSERT INTO permissions (code, name, resource_type, http_method, path, description, status)
VALUES 
('application:view', '查看应用', 'application', 'GET', '/api/v1/applications*', '查看应用列表和详情', 1),
('application:create', '创建应用', 'application', 'POST', '/api/v1/applications', '创建新应用', 1),
('application:update', '更新应用', 'application', 'PUT', '/api/v1/applications*', '更新应用信息', 1),
('application:delete', '删除应用', 'application', 'DELETE', '/api/v1/applications*', '删除应用', 1),
('component:view', '查看组件', 'component', 'GET', '/api/v1/components*', '查看组件列表', 1),
('component:manage', '管理组件', 'component', 'POST,PUT,DELETE', '/api/v1/components*', '创建/更新/删除组件', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 为SUPER_ADMIN分配所有应用权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'SUPER_ADMIN'
AND p.code IN ('application:view', 'application:create', 'application:update', 'application:delete', 'component:view', 'component:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为DEVELOPER分配应用查看和组件查看权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'DEVELOPER'
AND p.code IN ('application:view', 'component:view')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为PROJECT_MANAGER分配所有应用权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'PROJECT_MANAGER'
AND p.code IN ('application:view', 'application:create', 'application:update', 'application:delete', 'component:view', 'component:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);
