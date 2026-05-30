-- 添加节点和命名空间管理权限
USE iam_db;

-- 添加节点管理权限
INSERT INTO permissions (code, name, resource_type, http_method, path, description, status) VALUES
('node:view', '节点查看', 'node', 'GET', '/api/v1/nodes*', '查看集群节点', 1),
('node:manage', '节点管理', 'node', 'POST,PUT,DELETE', '/api/v1/nodes*', '管理集群节点', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 添加命名空间管理权限
INSERT INTO permissions (code, name, resource_type, http_method, path, description, status) VALUES
('namespace:view', '命名空间查看', 'namespace', 'GET', '/api/v1/namespaces*', '查看命名空间', 1),
('namespace:manage', '命名空间管理', 'namespace', 'POST,PUT,DELETE', '/api/v1/namespaces*', '管理命名空间', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 分配给SUPER_ADMIN角色（role_id=9）
INSERT INTO role_permissions (role_id, permission_id)
SELECT 9, id FROM permissions WHERE code IN ('node:view', 'node:manage', 'namespace:view', 'namespace:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 分配给PROJECT_ADMIN角色（role_id=10）  
INSERT INTO role_permissions (role_id, permission_id)
SELECT 10, id FROM permissions WHERE code IN ('node:view', 'namespace:view')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 分配给DEVELOPER角色（role_id=11）
INSERT INTO role_permissions (role_id, permission_id)
SELECT 11, id FROM permissions WHERE code IN ('node:view', 'namespace:view')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 分配给OPS角色（role_id=12）
INSERT INTO role_permissions (role_id, permission_id)
SELECT 12, id FROM permissions WHERE code IN ('node:view', 'node:manage', 'namespace:view', 'namespace:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 分配给GUEST角色（role_id=13）
INSERT INTO role_permissions (role_id, permission_id)
SELECT 13, id FROM permissions WHERE code IN ('node:view', 'namespace:view')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

SELECT '权限添加完成' as message;
