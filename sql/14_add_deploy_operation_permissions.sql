-- 添加部署操作权限（scale/restart）
USE iam_db;

-- 新增 deployment:restart 权限
INSERT IGNORE INTO permissions (code, name, resource_type, path, http_method, description, created_by, status)
VALUES ('deployment:restart', '重启部署', 'api', '/api/v1/deployments/*/restart/', 'POST', '重启部署实例', 'system', 1);

-- 确保 deployment:scale 权限存在
INSERT IGNORE INTO permissions (code, name, resource_type, path, http_method, description, created_by, status)
VALUES ('deployment:scale', '扩缩容', 'api', '/api/v1/deployments/*/scale/', 'POST', '部署实例扩缩容', 'system', 1);

-- 给 SUPER_ADMIN 角色添加权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.code = 'SUPER_ADMIN' AND p.code IN ('deployment:restart', 'deployment:scale');

-- 给 OPS 角色添加权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.code = 'OPS' AND p.code IN ('deployment:restart', 'deployment:scale');

-- 给 DEVELOPER 角色添加权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.code = 'DEVELOPER' AND p.code IN ('deployment:restart', 'deployment:scale');
