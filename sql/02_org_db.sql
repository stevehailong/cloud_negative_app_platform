-- 组织数据库 - 租户、组织和项目管理
USE org_db;

-- 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '租户名称',
    code VARCHAR(100) NOT NULL UNIQUE COMMENT '租户编码',
    description VARCHAR(500) COMMENT '租户描述',
    contact_name VARCHAR(100) COMMENT '联系人姓名',
    contact_phone VARCHAR(20) COMMENT '联系人电话',
    contact_email VARCHAR(100) COMMENT '联系人邮箱',
    expired_at TIMESTAMP NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    created_by VARCHAR(100) COMMENT '创建人',
    updated_by VARCHAR(100) COMMENT '更新人',
    status TINYINT DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
    INDEX idx_code (code),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- 组织表
CREATE TABLE IF NOT EXISTS organizations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    parent_id BIGINT UNSIGNED DEFAULT 0 COMMENT '父组织ID',
    name VARCHAR(100) NOT NULL COMMENT '组织名称',
    code VARCHAR(100) NOT NULL COMMENT '组织编码',
    description VARCHAR(500) COMMENT '组织描述',
    level INT DEFAULT 1 COMMENT '组织层级',
    sort INT DEFAULT 0 COMMENT '排序',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    created_by VARCHAR(100) COMMENT '创建人',
    updated_by VARCHAR(100) COMMENT '更新人',
    status TINYINT DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_parent_id (parent_id),
    INDEX idx_code (code),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表';

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '租户ID',
    org_id BIGINT UNSIGNED NOT NULL COMMENT '组织ID',
    name VARCHAR(100) NOT NULL COMMENT '项目名称',
    code VARCHAR(100) NOT NULL UNIQUE COMMENT '项目编码',
    description VARCHAR(500) COMMENT '项目描述',
    type VARCHAR(50) COMMENT '项目类型',
    owner VARCHAR(100) COMMENT '项目负责人',
    start_date DATE COMMENT '开始日期',
    end_date DATE COMMENT '结束日期',
    budget DECIMAL(15,2) COMMENT '预算',
    actual_cost DECIMAL(15,2) DEFAULT 0 COMMENT '实际成本',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    created_by VARCHAR(100) COMMENT '创建人',
    updated_by VARCHAR(100) COMMENT '更新人',
    status TINYINT DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_org_id (org_id),
    INDEX idx_code (code),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- 项目成员表
CREATE TABLE IF NOT EXISTS project_members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL COMMENT '项目ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    role VARCHAR(50) NOT NULL COMMENT '角色:owner,admin,developer,viewer',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    INDEX idx_project_id (project_id),
    INDEX idx_user_id (user_id),
    UNIQUE KEY uk_project_user (project_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员表';

-- 插入默认租户
INSERT INTO tenants (name, code, description, created_by, status) 
VALUES ('默认租户', 'default', '系统默认租户', 'system', 1);

-- 插入默认组织
INSERT INTO organizations (tenant_id, name, code, description, created_by, status) 
VALUES (1, '默认组织', 'default', '系统默认组织', 'system', 1);
