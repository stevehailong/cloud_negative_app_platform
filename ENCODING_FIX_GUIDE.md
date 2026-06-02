# 监控页面集群名称乱码问题修复指南

生成时间: 2026-06-01

## 问题描述

访问 http://localhost/monitors 监控页面时，集群名称显示为乱码：
```
œœ¬å¢œ°Kuberneteséšç¾¤
```

## 根本原因

### 问题1: 数据库字符编码错误 ❌
数据在插入数据库时使用了错误的字符集连接（latin1），导致UTF-8中文被错误地存储为双重编码的数据。

**检查方法**:
```bash
docker exec my-cloud-mysql mysql -uroot -proot123456 infra_db \
  -e "SELECT id, cluster_name, HEX(cluster_name) FROM clusters;"
```

发现数据的HEX值为:
```
C3A6C593C2ACC3A5C593C2B04B756265726E65746573...
```
这是典型的UTF-8被当作Latin1存储的特征。

### 问题2: 前端字段名映射 ✅ (已修复)
前端代码在处理集群数据时，字段名优先级设置不正确。

## 修复方案

### 1. 修复后端响应头 ✅

**文件**: `backend/cmd/cluster-service/main.go`

添加中间件确保所有HTTP响应包含正确的Content-Type:

```go
// 初始化 Gin 路由
r := gin.Default()

// 添加中间件确保正确的Content-Type
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
    c.Next()
})

// 注册路由
router.RegisterRoutes(r, clusterHandler)
```

### 2. 修复数据库中的损坏数据 ✅

**执行命令**:
```bash
docker exec my-cloud-mysql mysql -uroot -proot123456 infra_db \
  --default-character-set=utf8mb4 \
  -e "UPDATE clusters SET cluster_name = '本地Kubernetes集群' WHERE id = 1;"
```

**验证修复**:
```bash
docker exec my-cloud-mysql mysql -uroot -proot123456 infra_db \
  --default-character-set=utf8mb4 \
  -e "SELECT id, cluster_name FROM clusters;"
```

应该显示:
```
id	cluster_name
1	本地Kubernetes集群
```

### 3. 修复前端字段映射 ✅

**文件**: `frontend/src/views/monitor/MonitorDashboard.vue:292-326`

```javascript
targetOptions.value = (data.list || []).map(item => {
  // 根据不同类型选择正确的名称字段
  let name = ''
  if (metricsQuery.type === 'app') {
    name = item.appName || item.name
  } else if (metricsQuery.type === 'environment') {
    name = item.envName || item.name
  } else if (metricsQuery.type === 'cluster') {
    name = item.clusterName || item.name  // 正确使用 clusterName
  }
  return {
    id: item.id,
    name: name
  }
})
```

### 4. 修复应用指标格式化 ✅

**文件**: `frontend/src/views/monitor/MonitorDashboard.vue:328-374`

添加数值格式化函数:
```javascript
// 格式化数值，保留小数点后1位
const formatNumber = (val) => {
  if (val === null || val === undefined || val === '--') return '--'
  if (typeof val === 'number') {
    return val.toFixed(1)
  }
  return val
}

metrics.value = {
  cpu: formatNumber(data.cpu),
  memory: formatNumber(data.memory),
  qps: data.qps !== undefined && data.qps !== null ? Math.round(data.qps) : '--',
  errorRate: formatNumber(data.errorRate),
  // ...
}
```

## 部署步骤

### 1. 重新构建和部署 cluster-service
```bash
cd /Users/hanhailong01/Downloads/my_cloud
docker-compose build cluster-service
docker-compose up -d cluster-service
```

### 2. 重新构建和部署 frontend
```bash
docker-compose build frontend
docker-compose up -d frontend
```

### 3. 验证修复
```bash
# 测试 cluster-service API
curl -s 'http://localhost:8088/api/v1/clusters?page=1&pageSize=10' | python3 -m json.tool

# 应该看到正确的中文集群名称
```

## 验证结果

### API 测试
```bash
$ curl -s 'http://localhost:8088/api/v1/clusters?page=1&pageSize=10' | \
  python3 -c "import sys, json; d=json.load(sys.stdin); \
  [print(f'ID:{c[\"id\"]}, 名称:{c[\"clusterName\"]}') for c in d['data']['list']]"

ID:1, 名称:本地Kubernetes集群  ✅
```

### 前端页面
访问 http://localhost/monitors:
1. 选择"监控对象" = "集群"
2. 集群名称下拉框应显示: **本地Kubernetes集群** ✅
3. 选择"监控对象" = "应用"
4. 指标卡片应显示格式化的数值 ✅

## 预防措施

### 1. 数据库连接字符集
确保所有服务的数据库连接字符串都包含 `charset=utf8mb4`:

```go
dsn := "root:root123456@tcp(mysql:3306)/infra_db?charset=utf8mb4&parseTime=True&loc=Local"
```

### 2. MySQL 客户端字符集
在使用 mysql 命令行工具时，始终指定字符集:
```bash
mysql --default-character-set=utf8mb4 ...
```

### 3. 响应头设置
所有 Gin 服务都应该添加 Content-Type 中间件:
```go
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
    c.Next()
})
```

## 后续检查清单

- [x] cluster-service 已添加 Content-Type 中间件
- [x] 数据库中的集群名称已修复
- [x] 前端字段映射已修复
- [x] 前端数值格式化已修复
- [x] API 返回正确的中文
- [ ] 检查其他表是否也有编码问题
- [ ] 检查其他服务是否需要添加 Content-Type 中间件

## 常见问题

### Q: 刷新浏览器后还是显示乱码？
A: 请使用硬刷新 (Cmd+Shift+R 或 Ctrl+Shift+R)，或者清除浏览器缓存。

### Q: 其他中文字段是否也受影响？
A: 可能。建议检查所有包含中文的表，使用相同的方法修复。

### Q: 如何检查其他表的编码问题？
A: 
```bash
# 查看表中的中文数据HEX值
docker exec my-cloud-mysql mysql -uroot -proot123456 <database_name> \
  -e "SELECT id, <column_name>, HEX(<column_name>) FROM <table_name>;"
  
# 如果HEX值以 C3A6, C3A5 等开头，说明存在编码问题
```

## 总结

问题的根本原因是数据库字符编码配置不当，导致UTF-8中文被错误存储。通过：
1. 添加正确的 Content-Type 响应头
2. 修复数据库中已损坏的数据
3. 修复前端字段映射和格式化逻辑

现在监控页面可以正确显示集群名称和指标数据。
