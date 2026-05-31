-- 添加项目成员管理权限
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
USE iam_db;

-- 项目成员权限
INSERT INTO permissions (code, name, resource_type, http_method, path, description, status)
VALUES 
('project-member:view', '查看项目成员', 'project-member', 'GET', '/api/v1/project-members*', '查看项目成员列表', 1),
('project-member:manage', '管理项目成员', 'project-member', 'POST,DELETE', '/api/v1/project-members*', '添加/移除项目成员', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 为SUPER_ADMIN分配项目成员权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'SUPER_ADMIN'
AND p.code IN ('project-member:view', 'project-member:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为PROJECT_MANAGER分配项目成员权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'PROJECT_MANAGER'
AND p.code IN ('project-member:view', 'project-member:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为DEVELOPER分配查看项目成员权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'DEVELOPER'
AND p.code = 'project-member:view'
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);
