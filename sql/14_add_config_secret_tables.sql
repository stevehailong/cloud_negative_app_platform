-- 添加ConfigMap和Secret管理表
USE env_db;

-- ConfigMap配置表
CREATE TABLE IF NOT EXISTS config_maps (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  name VARCHAR(128) NOT NULL COMMENT 'ConfigMap名称',
  env_id BIGINT NOT NULL COMMENT '环境ID',
  namespace VARCHAR(128) NOT NULL COMMENT '命名空间',
  data JSON NOT NULL COMMENT '配置数据(key-value pairs)',
  description VARCHAR(255) NULL COMMENT '描述',
  sync_status VARCHAR(32) DEFAULT 'pending' COMMENT '同步状态: pending/synced/failed',
  sync_message TEXT NULL COMMENT '同步消息',
  last_sync_time DATETIME NULL COMMENT '最后同步时间',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  create_by BIGINT DEFAULT NULL COMMENT '创建人',
  update_by BIGINT DEFAULT NULL COMMENT '更新人',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
  UNIQUE KEY uk_env_name(env_id, name, is_deleted),
  KEY idx_env_id(env_id),
  KEY idx_namespace(namespace),
  KEY idx_sync_status(sync_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ConfigMap配置表';

-- Secret密钥表
CREATE TABLE IF NOT EXISTS secrets (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  name VARCHAR(128) NOT NULL COMMENT 'Secret名称',
  env_id BIGINT NOT NULL COMMENT '环境ID',
  namespace VARCHAR(128) NOT NULL COMMENT '命名空间',
  secret_type VARCHAR(32) DEFAULT 'Opaque' COMMENT 'Secret类型: Opaque/TLS/DockerConfigJson',
  data JSON NOT NULL COMMENT '密钥数据(key-value pairs, base64编码)',
  description VARCHAR(255) NULL COMMENT '描述',
  sync_status VARCHAR(32) DEFAULT 'pending' COMMENT '同步状态: pending/synced/failed',
  sync_message TEXT NULL COMMENT '同步消息',
  last_sync_time DATETIME NULL COMMENT '最后同步时间',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  create_by BIGINT DEFAULT NULL COMMENT '创建人',
  update_by BIGINT DEFAULT NULL COMMENT '更新人',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
  UNIQUE KEY uk_env_name(env_id, name, is_deleted),
  KEY idx_env_id(env_id),
  KEY idx_namespace(namespace),
  KEY idx_secret_type(secret_type),
  KEY idx_sync_status(sync_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Secret密钥表';

-- 添加ConfigMap和Secret相关权限
USE iam_db;

INSERT INTO permissions (code, name, resource_type, http_method, path, description, status)
VALUES 
('configmap:view', '查看ConfigMap', 'configmap', 'GET', '/api/v1/config-maps*', '查看ConfigMap列表和详情', 1),
('configmap:manage', '管理ConfigMap', 'configmap', 'POST,PUT,DELETE', '/api/v1/config-maps*', '创建/更新/删除ConfigMap', 1),
('secret:view', '查看Secret', 'secret', 'GET', '/api/v1/secrets*', '查看Secret列表和详情', 1),
('secret:manage', '管理Secret', 'secret', 'POST,PUT,DELETE', '/api/v1/secrets*', '创建/更新/删除Secret', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 为SUPER_ADMIN分配ConfigMap和Secret权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'SUPER_ADMIN'
AND p.code IN ('configmap:view', 'configmap:manage', 'secret:view', 'secret:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为PROJECT_MANAGER分配ConfigMap和Secret权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'PROJECT_MANAGER'
AND p.code IN ('configmap:view', 'configmap:manage', 'secret:view', 'secret:manage')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为DEVELOPER分配查看权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'DEVELOPER'
AND p.code IN ('configmap:view', 'secret:view')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);
