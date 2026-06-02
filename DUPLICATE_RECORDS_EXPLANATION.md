# 重复记录问题说明

## 当前情况

根据之前的查询,数据库中有3条记录:

```
id=1: app_id=6, namespace=app-6,  workload_name=app-6
id=7: app_id=8, namespace=app-8,  workload_name=app-8         ← app-8 记录1
id=6: app_id=8, namespace=app-8,  workload_name=app-8-canary  ← app-8 记录2
```

## 判断是否重复

### ✅ 正常情况 (不是重复)
如果 app-8 的两条记录是:
- 记录1: `namespace=app-8, workload_name=app-8` (stable版本)
- 记录2: `namespace=app-8, workload_name=app-8-canary` (canary版本)

**这是正确的!** 因为:
- 同一个应用在同一个环境只有一个 namespace (`app-8`)
- 但可以有两个 workload: stable (`app-8`) 和 canary (`app-8-canary`)
- 这是金丝雀部署的标准设计

### ❌ 重复情况 (需要清理)
如果 app-8 的两条记录是:
- 记录1: `namespace=app-8, workload_name=app-8`
- 记录2: `namespace=app-8, workload_name=app-8` (完全相同!)

**这才是重复!** 需要删除其中一条。

## 如何检查

### 方式1: 使用 SQL 脚本
```bash
# 启动 MySQL (如果在 Docker 中)
docker start mysql

# 或启动 Docker Desktop

# 执行检查脚本
mysql -h 127.0.0.1 -u root -proot123456 deploy_db < /Users/hanhailong01/Downloads/my_cloud/check_and_cleanup.sql
```

### 方式2: 直接查询
```bash
mysql -h 127.0.0.1 -u root -proot123456 -e "
SELECT 
    id,
    namespace,
    workload_name,
    deployment_status,
    update_time
FROM deploy_db.app_deployments 
WHERE app_id = 8 AND env_id = 1
ORDER BY workload_name, update_time DESC;
"
```

## 如何清理

### 如果确认有重复:

```sql
-- 1. 备份
mysqldump -h 127.0.0.1 -u root -proot123456 deploy_db app_deployments > backup.sql

-- 2. 删除重复记录 (保留最新的)
DELETE d1 FROM app_deployments d1
INNER JOIN app_deployments d2 ON 
    d1.namespace = d2.namespace AND 
    d1.workload_name = d2.workload_name AND
    d1.app_id = d2.app_id AND
    d1.env_id = d2.env_id AND
    d1.update_time < d2.update_time
WHERE d1.app_id = 8 AND d1.env_id = 1;

-- 3. 验证
SELECT * FROM app_deployments WHERE app_id = 8 AND env_id = 1;
```

### 使用 Go 清理工具:

```bash
cd /Users/hanhailong01/Downloads/my_cloud/backend

# 修改连接字符串 (如果需要)
# 编辑 cmd/cleanup-duplicates/main.go 中的 dsn

# 运行清理
go run ./cmd/cleanup-duplicates/main.go
```

## 预期结果

清理后,app-8 应该有 **最多2条记录**:
1. `namespace=app-8, workload_name=app-8` (stable)
2. `namespace=app-8, workload_name=app-8-canary` (canary,可选)

如果只做滚动部署,可能只有1条 stable 记录。
如果做金丝雀部署,会有2条记录 (stable + canary)。

## 下一步

1. **启动 MySQL**
   ```bash
   # 如果在 Docker 中
   docker start mysql
   
   # 或启动 Docker Desktop
   ```

2. **执行检查脚本**
   ```bash
   mysql -h 127.0.0.1 -u root -proot123456 deploy_db < check_and_cleanup.sql
   ```

3. **如果确认有重复,取消注释清理 SQL 并重新执行**

4. **验证唯一索引已创建**
   ```sql
   SHOW INDEX FROM app_deployments WHERE Key_name = 'uk_namespace_workload';
   ```

## 总结

- **不是所有 app-8 的记录都是重复**
- **只有 (namespace, workload_name) 完全相同的才是重复**
- **stable 和 canary 是两个不同的 workload,不是重复**
