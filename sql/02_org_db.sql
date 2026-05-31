-- 组织数据库 - 租户、组织和项目管理
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
USE org_db;

-- 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_code VARCHAR(64) NOT NULL UNIQUE COMMENT '租户编码',
    tenant_name VARCHAR(128) NOT NULL COMMENT '租户名称',
    contact_email VARCHAR(128) NULL COMMENT '联系邮箱',
    contact_phone VARCHAR(32) NULL COMMENT '联系电话',
    status TINYINT DEFAULT 1 COMMENT '状态 1-启用 0-禁用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    create_by BIGINT DEFAULT NULL COMMENT '创建人',
    update_by BIGINT DEFAULT NULL COMMENT '更新人',
    is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
    INDEX idx_tenant_code (tenant_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- 组织表
CREATE TABLE IF NOT EXISTS organizations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    parent_id BIGINT UNSIGNED DEFAULT NULL COMMENT '父组织ID',
    org_code VARCHAR(64) NOT NULL COMMENT '组织编码',
    org_name VARCHAR(128) NOT NULL COMMENT '组织名称',
    org_level INT DEFAULT 0 COMMENT '组织层级',
    org_path VARCHAR(512) DEFAULT '/' COMMENT '组织路径',
    description VARCHAR(255) NULL COMMENT '描述',
    status TINYINT DEFAULT 1 COMMENT '状态 1-启用 0-禁用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    create_by BIGINT DEFAULT NULL COMMENT '创建人',
    update_by BIGINT DEFAULT NULL COMMENT '更新人',
    is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
    UNIQUE KEY uk_tenant_org_code(tenant_id, org_code),
    KEY idx_parent_id(parent_id),
    KEY idx_tenant_id(tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表';

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    org_id BIGINT UNSIGNED DEFAULT NULL COMMENT '组织ID',
    project_code VARCHAR(64) NOT NULL UNIQUE COMMENT '项目编码',
    project_name VARCHAR(128) NOT NULL COMMENT '项目名称',
    owner_user_id BIGINT DEFAULT NULL COMMENT '负责人',
    description VARCHAR(255) NULL COMMENT '描述',
    visibility VARCHAR(32) DEFAULT 'private' COMMENT '可见性 private/internal/public',
    status TINYINT DEFAULT 1 COMMENT '状态 1-启用 0-禁用',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    create_by BIGINT DEFAULT NULL COMMENT '创建人',
    update_by BIGINT DEFAULT NULL COMMENT '更新人',
    is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
    KEY idx_tenant_id(tenant_id),
    KEY idx_org_id(org_id),
    KEY idx_owner_user_id(owner_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- 项目成员表
CREATE TABLE IF NOT EXISTS project_members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL COMMENT '项目ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    role_code VARCHAR(64) NOT NULL COMMENT '项目角色 owner/maintainer/developer/reporter',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_by BIGINT DEFAULT NULL COMMENT '创建人',
    UNIQUE KEY uk_project_user(project_id, user_id),
    KEY idx_user_id(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员表';

-- 插入默认租户
INSERT INTO tenants (tenant_code, tenant_name, contact_email, status)
VALUES ('default', '默认租户', 'admin@example.com', 1)
ON DUPLICATE KEY UPDATE tenant_name='默认租户';

-- 插入默认组织
INSERT INTO organizations (tenant_id, org_code, org_name, description, create_by, status)
VALUES (1, 'default', '默认组织', '系统默认组织', 0, 1)
ON DUPLICATE KEY UPDATE org_name='默认组织';