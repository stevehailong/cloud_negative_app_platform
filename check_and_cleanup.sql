-- 检查和清理重复部署记录
-- 执行方式: mysql -h 127.0.0.1 -u root -proot123456 deploy_db < check_and_cleanup.sql

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

USE deploy_db;

-- 1. 查看所有 app-8 的记录
SELECT '=== 当前 app-8 的所有记录 ===' as '';
SELECT 
    id,
    app_id,
    env_id,
    namespace,
    workload_name,
    deployment_status,
    create_time,
    update_time
FROM app_deployments 
WHERE app_id = 8 AND env_id = 1
ORDER BY workload_name, update_time DESC;

-- 2. 检查是否有真正的重复 (namespace + workload_name 相同)
SELECT '\n=== 检查重复记录 (namespace + workload_name 相同) ===' as '';
SELECT 
    namespace,
    workload_name,
    COUNT(*) as count,
    GROUP_CONCAT(id ORDER BY update_time DESC) as ids,
    GROUP_CONCAT(deployment_status ORDER BY update_time DESC) as statuses
FROM app_deployments
WHERE app_id = 8 AND env_id = 1
GROUP BY namespace, workload_name
HAVING COUNT(*) > 1;

-- 3. 如果上面查询有结果,说明存在重复,执行清理
-- 保留最新的记录,删除旧的
-- 取消下面的注释来执行清理:

/*
DELETE d1 FROM app_deployments d1
INNER JOIN app_deployments d2 ON 
    d1.namespace = d2.namespace AND 
    d1.workload_name = d2.workload_name AND
    d1.app_id = d2.app_id AND
    d1.env_id = d2.env_id AND
    d1.update_time < d2.update_time
WHERE d1.app_id = 8 AND d1.env_id = 1;
*/

-- 4. 验证清理后的结果
SELECT '\n=== 清理后的记录 ===' as '';
SELECT 
    id,
    app_id,
    env_id,
    namespace,
    workload_name,
    deployment_status,
    update_time
FROM app_deployments 
WHERE app_id = 8 AND env_id = 1
ORDER BY workload_name;

-- 5. 确认唯一索引
SELECT '\n=== 检查唯一索引 ===' as '';
SHOW INDEX FROM app_deployments WHERE Key_name LIKE '%namespace%' OR Key_name LIKE '%workload%';
