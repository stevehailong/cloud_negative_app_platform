-- 应用数据库 - 应用和组件管理
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
USE app_db;

-- 应用表
CREATE TABLE IF NOT EXISTS applications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '应用名称',
    code VARCHAR(100) NOT NULL UNIQUE COMMENT '应用编码',
    project_id BIGINT UNSIGNED NOT NULL COMMENT '项目ID',
    description VARCHAR(500) COMMENT '应用描述',
    type VARCHAR(50) COMMENT '应用类型:web,api,job,function',
    language VARCHAR(50) COMMENT '开发语言',
    framework VARCHAR(50) COMMENT '开发框架',
    repo_url VARCHAR(500) COMMENT '代码仓库地址',
    repo_branch VARCHAR(100) COMMENT '默认分支',
    build_tool VARCHAR(50) COMMENT '构建工具',
    build_path VARCHAR(200) COMMENT '构建路径',
    docker_file VARCHAR(200) COMMENT 'Dockerfile路径',
    health_check TEXT COMMENT '健康检查配置',
    labels TEXT COMMENT '标签(JSON)',
    owner VARCHAR(100) COMMENT '负责人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    created_by VARCHAR(100) COMMENT '创建人',
    updated_by VARCHAR(100) COMMENT '更新人',
    status TINYINT DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
    INDEX idx_project_id (project_id),
    INDEX idx_code (code),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用表';

-- 组件表
CREATE TABLE IF NOT EXISTS components (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    application_id BIGINT UNSIGNED NOT NULL COMMENT '应用ID',
    name VARCHAR(100) NOT NULL COMMENT '组件名称',
    type VARCHAR(50) COMMENT '组件类型:frontend,backend,database,cache',
    version VARCHAR(50) COMMENT '版本',
    image VARCHAR(500) COMMENT '镜像地址',
    port INT COMMENT '端口',
    replicas INT DEFAULT 1 COMMENT '副本数',
    cpu VARCHAR(20) COMMENT 'CPU限制',
    memory VARCHAR(20) COMMENT '内存限制',
    env_vars TEXT COMMENT '环境变量(JSON)',
    config_maps TEXT COMMENT '配置映射(JSON)',
    secrets TEXT COMMENT '密钥(JSON)',
    volumes TEXT COMMENT '存储卷(JSON)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    created_by VARCHAR(100) COMMENT '创建人',
    updated_by VARCHAR(100) COMMENT '更新人',
    status TINYINT DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
    INDEX idx_application_id (application_id),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组件表';

-- 应用依赖表
CREATE TABLE IF NOT EXISTS app_dependencies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    application_id BIGINT UNSIGNED NOT NULL COMMENT '应用ID',
    depend_app_id BIGINT UNSIGNED NOT NULL COMMENT '依赖应用ID',
    depend_type VARCHAR(50) COMMENT '依赖类型:strong,weak',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_application_id (application_id),
    INDEX idx_depend_app_id (depend_app_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用依赖表';
