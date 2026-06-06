-- 成本记录表：添加唯一约束，防止同一 namespace+date+source 重复
-- 同时清理已有重复数据（保留 ID 最大的一条）

-- 1. 清理已有重复
DELETE t1
FROM cost_records t1
INNER JOIN cost_records t2
WHERE t1.id < t2.id
  AND t1.cluster_id = t2.cluster_id
  AND t1.namespace  = t2.namespace
  AND t1.cost_date  = t2.cost_date
  AND t1.source     = t2.source;

-- 2. 添加唯一索引（如果不存在）
-- 注意：GORM AutoMigrate 会根据 model 的 uniqueIndex tag 自动创建，
-- 此 SQL 用于手动迁移场景
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_cost
  ON cost_records (cluster_id, namespace, cost_date, source);
