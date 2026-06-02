-- 新的部署管理架构设计

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 1. app_deployments 表（主记录表）
-- 每个应用在每个环境中只有一条记录
CREATE TABLE app_deployments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    app_id BIGINT NOT NULL COMMENT '应用ID',
    env_id BIGINT NOT NULL COMMENT '环境ID',
    cluster_id BIGINT NOT NULL COMMENT '集群ID',
    namespace VARCHAR(255) NOT NULL COMMENT 'K8s命名空间',
    workload_name VARCHAR(255) NOT NULL COMMENT '工作负载名称',
    workload_type VARCHAR(50) DEFAULT 'deployment' COMMENT '工作负载类型',
    
    -- 当前运行状态
    current_version VARCHAR(255) COMMENT '当前版本号',
    current_image VARCHAR(500) COMMENT '当前镜像',
    desired_replicas INT DEFAULT 1 COMMENT '期望副本数',
    available_replicas INT DEFAULT 0 COMMENT '可用副本数',
    deployment_status VARCHAR(50) COMMENT '部署状态: running, stopped, failed',
    
    -- 最后一次部署信息
    last_deploy_id BIGINT COMMENT '最后一次部署历史记录ID',
    last_deploy_time DATETIME COMMENT '最后部署时间',
    last_deploy_user_id BIGINT COMMENT '最后部署人',
    
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_namespace_workload (namespace, workload_name),
    KEY idx_namespace (namespace),
    KEY idx_workload_name (workload_name),
    KEY idx_app_env (app_id, env_id)
) COMMENT='应用部署主记录表';

-- 2. deployment_history 表（历史记录表）
-- 记录每次部署的详细信息
CREATE TABLE deployment_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    app_deployment_id BIGINT NOT NULL COMMENT '关联的应用部署ID',
    release_id BIGINT COMMENT '关联的发布工单ID',
    
    -- 部署详情
    version VARCHAR(255) COMMENT '版本号',
    image_url VARCHAR(500) COMMENT '镜像地址',
    replicas INT COMMENT '副本数',
    deployment_type VARCHAR(50) COMMENT '部署类型: create, update, rollback, restart, scale',
    
    -- 执行信息
    operator_user_id BIGINT COMMENT '操作人',
    start_time DATETIME COMMENT '开始时间',
    end_time DATETIME COMMENT '结束时间',
    duration INT COMMENT '耗时(秒)',
    status VARCHAR(50) COMMENT '状态: success, failed, progressing',
    failure_reason TEXT COMMENT '失败原因',
    
    -- 变更记录
    changes JSON COMMENT '变更内容: {"image": "old->new", "replicas": "3->5"}',
    
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    KEY idx_app_deployment (app_deployment_id),
    KEY idx_release (release_id),
    KEY idx_operator (operator_user_id),
    KEY idx_create_time (create_time)
) COMMENT='部署历史记录表';

-- 3. 迁移当前数据示例
-- 从 deployments 表迁移到新表结构

-- 创建主记录
INSERT INTO app_deployments (app_id, env_id, cluster_id, namespace, workload_name, 
    current_version, current_image, desired_replicas, available_replicas, 
    deployment_status, last_deploy_time)
SELECT 
    8 as app_id,
    1 as env_id,
    cluster_id,
    namespace,
    workload_name,
    image_version as current_version,
    image_version as current_image,
    desired_replicas,
    available_replicas,
    deployment_status,
    update_time as last_deploy_time
FROM deployments 
WHERE namespace = 'app-8' 
ORDER BY id DESC 
LIMIT 1;

-- 创建历史记录
INSERT INTO deployment_history (app_deployment_id, release_id, version, image_url, 
    replicas, deployment_type, start_time, end_time, status)
SELECT 
    LAST_INSERT_ID() as app_deployment_id,
    release_id,
    image_version as version,
    image_version as image_url,
    desired_replicas as replicas,
    'update' as deployment_type,
    start_time,
    end_time,
    deployment_status as status
FROM deployments 
WHERE namespace = 'app-8'
ORDER BY id ASC;
