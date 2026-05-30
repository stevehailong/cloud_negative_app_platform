-- GitLab集成权限
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
USE iam_db;

-- 添加GitLab相关权限
INSERT INTO permissions (code, name, resource_type, http_method, path, description, created_by, status) VALUES
('gitlab:test', '测试GitLab连接', 'gitlab', 'POST', '/api/v1/gitlab/test*', '测试GitLab连接是否正常', 'system', 1),
('gitlab:projects', '查看GitLab项目', 'gitlab', 'GET', '/api/v1/gitlab/projects*', '获取GitLab项目列表和分支', 'system', 1),
('gitlab:webhooks', '管理GitLab Webhooks', 'gitlab', 'POST', '/api/v1/gitlab/webhooks*', '创建GitLab Webhook', 'system', 1),
('gitlab:config', '配置GitLab客户端', 'gitlab', 'PUT', '/api/v1/gitlab/client*', '动态更新GitLab配置', 'system', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 超级管理员拥有所有GitLab权限（使用code查找角色，兼容不同ID）
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM permissions p, roles r WHERE p.code LIKE 'gitlab:%' AND r.code = 'SUPER_ADMIN';

-- 项目管理员拥有查看和测试权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM permissions p, roles r WHERE p.code IN ('gitlab:test', 'gitlab:projects') AND r.code = 'PROJECT_ADMIN';

-- 开发人员拥有查看和测试权限
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM permissions p, roles r WHERE p.code IN ('gitlab:test', 'gitlab:projects') AND r.code = 'DEVELOPER';
