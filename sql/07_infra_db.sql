-- 集群管理服务数据库
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS infra_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE infra_db;

-- 集群表
CREATE TABLE IF NOT EXISTS clusters (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  cluster_code VARCHAR(64) NOT NULL UNIQUE COMMENT '集群编码',
  cluster_name VARCHAR(128) NOT NULL COMMENT '集群名称',
  cluster_type VARCHAR(32) NOT NULL COMMENT '集群类型 kubernetes/docker-swarm',
  api_server VARCHAR(255) NOT NULL COMMENT 'API Server地址',
  kubeconfig TEXT NULL COMMENT 'kubeconfig配置',
  version VARCHAR(64) NULL COMMENT 'K8s版本',
  region VARCHAR(64) NULL COMMENT '所属区域',
  zone VARCHAR(64) NULL COMMENT '所属可用区',
  description VARCHAR(255) NULL COMMENT '描述',
  status TINYINT DEFAULT 1 COMMENT '状态 1-正常 0-异常',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  create_by BIGINT DEFAULT NULL COMMENT '创建人',
  update_by BIGINT DEFAULT NULL COMMENT '更新人',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
  KEY idx_cluster_type(cluster_type),
  KEY idx_region(region)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群表';

-- 节点表
CREATE TABLE IF NOT EXISTS cluster_nodes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  cluster_id BIGINT NOT NULL COMMENT '集群ID',
  node_name VARCHAR(128) NOT NULL COMMENT '节点名称',
  node_ip VARCHAR(64) NOT NULL COMMENT '节点IP',
  node_role VARCHAR(32) NOT NULL COMMENT '节点角色 master/worker',
  cpu_cores INT DEFAULT 0 COMMENT 'CPU核数',
  memory_gb INT DEFAULT 0 COMMENT '内存GB',
  disk_gb INT DEFAULT 0 COMMENT '磁盘GB',
  os_image VARCHAR(128) NULL COMMENT '操作系统',
  container_runtime VARCHAR(64) NULL COMMENT '容器运行时',
  kubelet_version VARCHAR(64) NULL COMMENT 'Kubelet版本',
  status TINYINT DEFAULT 1 COMMENT '状态 1-Ready 0-NotReady',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
  KEY idx_cluster_id(cluster_id),
  KEY idx_node_role(node_role),
  FOREIGN KEY (cluster_id) REFERENCES clusters(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群节点表';

-- 命名空间表
CREATE TABLE IF NOT EXISTS namespaces (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  cluster_id BIGINT NOT NULL COMMENT '集群ID',
  namespace_name VARCHAR(128) NOT NULL COMMENT '命名空间名称',
  project_id BIGINT NOT NULL COMMENT '项目ID',
  resource_quota_json JSON NULL COMMENT '资源配额JSON',
  limit_range_json JSON NULL COMMENT '资源限制JSON',
  description VARCHAR(255) NULL COMMENT '描述',
  status TINYINT DEFAULT 1 COMMENT '状态 1-Active 0-Terminating',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  create_by BIGINT DEFAULT NULL COMMENT '创建人',
  update_by BIGINT DEFAULT NULL COMMENT '更新人',
  is_deleted TINYINT DEFAULT 0 COMMENT '是否删除',
  UNIQUE KEY uk_cluster_namespace(cluster_id, namespace_name),
  KEY idx_project_id(project_id),
  FOREIGN KEY (cluster_id) REFERENCES clusters(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='命名空间表';

-- 初始化示例集群
INSERT INTO clusters (cluster_code, cluster_name, cluster_type, api_server, version, region, description, status)
VALUES 
  ('local-k8s', '本地Kubernetes集群', 'kubernetes', 'https://kubernetes.default.svc', 'v1.28.0', 'local', '开发测试集群', 1)
ON DUPLICATE KEY UPDATE cluster_name=cluster_name;
