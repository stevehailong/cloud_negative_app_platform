-- 修正集群、环境、节点、命名空间权限路径配置
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
USE iam_db;

-- 修正集群权限路径
UPDATE permissions SET path = '/api/v1/clusters*' WHERE code IN ('cluster:view', 'cluster:manage');

-- 修正环境权限路径  
UPDATE permissions SET path = '/api/v1/environments*' WHERE code IN ('env:view', 'env:manage');

-- 修正节点权限路径
UPDATE permissions SET path = '/api/v1/nodes*' WHERE code IN ('node:view', 'node:manage');

-- 修正命名空间权限路径
UPDATE permissions SET path = '/api/v1/namespaces*' WHERE code IN ('namespace:view', 'namespace:manage');

-- 添加环境模板权限（如果不存在）
INSERT INTO permissions (code, name, resource_type, http_method, path, description, status) VALUES
('env-template:view', '环境模板查看', 'env-template', 'GET', '/api/v1/env-templates*', '查看环境模板', 1),
('env-template:manage', '环境模板管理', 'env-template', 'POST,PUT,DELETE', '/api/v1/env-templates*', '管理环境模板', 1),
('app-env-binding:view', '应用环境绑定查看', 'app-env-binding', 'GET', '/api/v1/app-env-bindings*', '查看应用环境绑定', 1),
('app-env-binding:manage', '应用环境绑定管理', 'app-env-binding', 'POST,PUT,DELETE', '/api/v1/app-env-bindings*', '管理应用环境绑定', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 分配环境模板权限给SUPER_ADMIN
INSERT INTO role_permissions (role_id, permission_id)
SELECT 9, id FROM permissions WHERE code LIKE 'env-template:%' OR code LIKE 'app-env-binding:%'
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

SELECT '权限路径修正完成' as message;
