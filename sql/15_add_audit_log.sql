-- 添加审计日志表
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS audit_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE audit_db;

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  user_id BIGINT NOT NULL COMMENT '用户ID',
  username VARCHAR(100) NOT NULL COMMENT '用户名',
  action VARCHAR(50) NOT NULL COMMENT '操作类型: create/update/delete/view',
  resource_type VARCHAR(50) NOT NULL COMMENT '资源类型: application/cluster/environment等',
  resource_id BIGINT DEFAULT NULL COMMENT '资源ID',
  resource_name VARCHAR(255) DEFAULT NULL COMMENT '资源名称',
  method VARCHAR(10) NOT NULL COMMENT 'HTTP方法: GET/POST/PUT/DELETE',
  path VARCHAR(500) NOT NULL COMMENT '请求路径',
  ip_address VARCHAR(50) DEFAULT NULL COMMENT 'IP地址',
  user_agent TEXT DEFAULT NULL COMMENT '用户代理',
  request_body TEXT DEFAULT NULL COMMENT '请求体(敏感信息需脱敏)',
  response_code INT DEFAULT NULL COMMENT '响应码',
  response_message VARCHAR(500) DEFAULT NULL COMMENT '响应消息',
  duration_ms INT DEFAULT NULL COMMENT '请求耗时(毫秒)',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  KEY idx_user_id(user_id),
  KEY idx_username(username),
  KEY idx_action(action),
  KEY idx_resource_type(resource_type),
  KEY idx_resource_id(resource_id),
  KEY idx_create_time(create_time),
  KEY idx_path(path(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审计日志表';

-- 添加审计日志查看权限
USE iam_db;

INSERT INTO permissions (code, name, resource_type, http_method, path, description, status)
VALUES 
('audit:view', '查看审计日志', 'audit', 'GET', '/api/v1/audit-logs*', '查看审计日志列表和详情', 1),
('audit:export', '导出审计日志', 'audit', 'POST', '/api/v1/audit-logs/export', '导出审计日志', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 为SUPER_ADMIN分配审计日志权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'SUPER_ADMIN'
AND p.code IN ('audit:view', 'audit:export')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为ADMIN分配审计日志查看权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'ADMIN'
AND p.code = 'audit:view'
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);
