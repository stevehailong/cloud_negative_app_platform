-- 环境管理服务数据库
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS env_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE env_db;

-- 环境表
CREATE TABLE IF NOT EXISTS environments (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  env_code VARCHAR(64) NOT NULL UNIQUE COMMENT '环境编码',
  env_name VARCHAR(128) NOT NULL COMMENT '环境名称',
  env_type VARCHAR(32) NOT NULL COMMENT '环境类型 dev/test/staging/prod/preview',
  cluster_id BIGINT NOT NULL COMMENT '集群ID',
  namespace VARCHAR(128) NOT NULL COMMENT '命名空间',
  project_id BIGINT NOT NULL COMMENT '项目ID',
  description VARCHAR(255) NULL COMMENT '描述',
  config_json JSON NULL COMMENT '环境配置JSON',
  status TINYINT DEFAULT 1 COMMENT '状态 1-启用 0-禁用',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  create_by BIGINT DEFAULT NULL COMMENT '创建人',
  update_by BIGINT DEFAULT NULL COMMENT '更新人',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
  KEY idx_cluster_id(cluster_id),
  KEY idx_project_id(project_id),
  KEY idx_namespace(namespace),
  KEY idx_env_type(env_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='环境表';

-- 环境模板表
CREATE TABLE IF NOT EXISTS env_templates (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  template_code VARCHAR(64) NOT NULL UNIQUE COMMENT '模板编码',
  template_name VARCHAR(128) NOT NULL COMMENT '模板名称',
  template_type VARCHAR(32) NOT NULL COMMENT '模板类型 helm/kustomize/yaml',
  repo_url VARCHAR(255) NULL COMMENT '模板仓库地址',
  chart_name VARCHAR(128) NULL COMMENT 'Chart名称',
  chart_version VARCHAR(64) NULL COMMENT 'Chart版本',
  values_yaml TEXT NULL COMMENT 'values.yaml内容',
  description VARCHAR(255) NULL COMMENT '描述',
  status TINYINT DEFAULT 1 COMMENT '状态 1-启用 0-禁用',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  create_by BIGINT DEFAULT NULL COMMENT '创建人',
  update_by BIGINT DEFAULT NULL COMMENT '更新人',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='环境模板表';

-- 应用环境绑定表
CREATE TABLE IF NOT EXISTS app_env_bindings (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  app_id BIGINT NOT NULL COMMENT '应用ID',
  env_id BIGINT NOT NULL COMMENT '环境ID',
  template_id BIGINT DEFAULT NULL COMMENT '模板ID',
  replicas INT DEFAULT 1 COMMENT '副本数',
  cpu_request VARCHAR(32) DEFAULT '100m' COMMENT 'CPU请求',
  cpu_limit VARCHAR(32) DEFAULT '500m' COMMENT 'CPU限制',
  memory_request VARCHAR(32) DEFAULT '128Mi' COMMENT '内存请求',
  memory_limit VARCHAR(32) DEFAULT '512Mi' COMMENT '内存限制',
  config_json JSON NULL COMMENT '配置JSON',
  status TINYINT DEFAULT 1 COMMENT '状态 1-启用 0-禁用',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  create_by BIGINT DEFAULT NULL COMMENT '创建人',
  update_by BIGINT DEFAULT NULL COMMENT '更新人',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
  UNIQUE KEY uk_app_env(app_id, env_id),
  KEY idx_env_id(env_id),
  KEY idx_template_id(template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用环境绑定表';

-- 初始化环境类型（示例数据）
INSERT INTO env_templates (template_code, template_name, template_type, description, status)
VALUES 
  ('basic-deployment', '基础部署模板', 'yaml', '最简单的Deployment+Service模板', 1),
  ('helm-app', 'Helm应用模板', 'helm', '标准Helm Chart应用模板', 1)
ON DUPLICATE KEY UPDATE template_name=template_name;
