-- 添加命名空间隔离的唯一性约束
-- 确保在同一个集群中，一个namespace只能属于一个环境

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

USE env_db;

-- 添加组合唯一索引：cluster_id + namespace
-- 这确保了在同一个集群中，每个namespace只能被一个环境使用
ALTER TABLE environments 
ADD UNIQUE KEY uk_cluster_namespace (cluster_id, namespace)
COMMENT '集群+命名空间唯一约束，确保命名空间隔离';

-- 说明：
-- 1. 一个环境 = 一个命名空间（在特定集群中）
-- 2. 不同集群可以有相同名称的namespace（但这通常不推荐）
-- 3. 应用部署时会自动使用环境定义的namespace
-- 4. 这是企业级Kubernetes多租户隔离的标准方案
