-- 清理重复的部署记录脚本
-- 执行前请备份数据库!

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

USE deploy_db;

-- 1. 查看当前重复记录
SELECT 
    namespace, 
    workload_name, 
    COUNT(*) as count,
    GROUP_CONCAT(id ORDER BY update_time DESC) as ids
FROM app_deployments
GROUP BY namespace, workload_name
HAVING COUNT(*) > 1;

-- 2. 删除重复记录 (保留最新的)
-- 创建临时表存储要保留的记录
CREATE TEMPORARY TABLE keep_records AS
SELECT MAX(id) as id
FROM app_deployments
GROUP BY namespace, workload_name;

-- 查看将要删除的记录
SELECT * FROM app_deployments 
WHERE id NOT IN (SELECT id FROM keep_records)
ORDER BY namespace, workload_name, update_time;

-- 删除重复记录 (取消注释以执行)
-- DELETE FROM app_deployments 
-- WHERE id NOT IN (SELECT id FROM keep_records);

-- 3. 删除旧索引
ALTER TABLE app_deployments DROP INDEX IF EXISTS idx_namespace_workload;
ALTER TABLE app_deployments DROP INDEX IF EXISTS idx_app_env_workload;

-- 4. 创建新的唯一索引
ALTER TABLE app_deployments 
ADD UNIQUE INDEX uk_namespace_workload (namespace, workload_name);

-- 5. 创建查询索引
ALTER TABLE app_deployments 
ADD INDEX idx_app_env (app_id, env_id);

-- 6. 验证结果
SELECT 
    namespace, 
    workload_name, 
    COUNT(*) as count
FROM app_deployments
GROUP BY namespace, workload_name
HAVING COUNT(*) > 1;

-- 应该返回空结果

-- 7. 查看最终数据
SELECT 
    id,
    app_id,
    env_id,
    namespace,
    workload_name,
    deployment_status,
    update_time
FROM app_deployments
ORDER BY app_id, env_id, workload_name;
