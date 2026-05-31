-- 系统设置表
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
USE iam_db;

CREATE TABLE IF NOT EXISTS system_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    setting_group VARCHAR(50) NOT NULL COMMENT '设置分组: basic/security/notification/integration',
    setting_key VARCHAR(100) NOT NULL COMMENT '设置键',
    setting_value TEXT COMMENT '设置值',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_group_key (setting_group, setting_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

-- 插入默认基本设置
INSERT IGNORE INTO system_settings (setting_group, setting_key, setting_value) VALUES
('basic', 'platformName', '云原生应用研发交付平台'),
('basic', 'platformShortName', 'My Cloud'),
('basic', 'platformLogo', ''),
('basic', 'contactEmail', 'support@example.com'),
('basic', 'supportPhone', '400-xxx-xxxx'),
('basic', 'icp', '');

-- 插入默认安全设置
INSERT IGNORE INTO system_settings (setting_group, setting_key, setting_value) VALUES
('security', 'sessionTimeout', '30'),
('security', 'passwordMinLength', '8'),
('security', 'passwordComplexity', '["lowercase","number"]'),
('security', 'loginLockEnabled', 'true'),
('security', 'loginLockAttempts', '5'),
('security', 'loginLockDuration', '30'),
('security', 'apiRateLimitEnabled', 'true'),
('security', 'apiRateLimit', '1000'),
('security', 'ipWhitelist', '');

-- 添加设置管理权限
INSERT IGNORE INTO permissions (code, name, resource_type, http_method, path, description, status, created_at, updated_at) VALUES
('settings:read', '查看系统设置', 'settings', 'GET', '/api/v1/settings*', '查看系统设置', 1, NOW(), NOW()),
('settings:write', '修改系统设置', 'settings', 'PUT', '/api/v1/settings*', '修改系统设置', 1, NOW(), NOW()),
('upload:file', '上传文件', 'upload', 'POST', '/api/v1/upload*', '上传文件', 1, NOW(), NOW());

-- 为 SUPER_ADMIN 和 OPS 角色分配设置权限
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r, permissions p
WHERE r.code IN ('SUPER_ADMIN', 'OPS')
AND p.code IN ('settings:read', 'settings:write', 'upload:file');

-- 为 DEVELOPER 角色分配设置只读和上传权限
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r, permissions p
WHERE r.code = 'DEVELOPER'
AND p.code IN ('settings:read', 'upload:file');
