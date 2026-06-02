# 应用环境绑定功能验证清单

## 后端验证

### 1. API接口测试

#### ✓ 环境列表API
```bash
curl -X GET 'http://localhost:8080/api/v1/environments?page=1&pageSize=10' \
  -H "Authorization: Bearer {YOUR_TOKEN}"
```
**预期结果**：
- 返回环境列表
- 每个环境包含：envName, envType, namespace, clusterName

#### ✓ 创建绑定API
```bash
curl -X POST 'http://localhost:8080/api/v1/app-env-bindings' \
  -H "Authorization: Bearer {YOUR_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "appId": 1,
    "envId": 1,
    "replicas": 1,
    "cpuRequest": "100m",
    "cpuLimit": "500m",
    "memoryRequest": "128Mi",
    "memoryLimit": "512Mi",
    "configJson": "{}"
  }'
```
**预期结果**：
- code: 0
- message: "创建成功"
- 返回创建的绑定对象

#### ✓ 查询绑定列表API
```bash
curl -X GET 'http://localhost:8080/api/v1/app-env-bindings?applicationId=1&page=1&pageSize=10' \
  -H "Authorization: Bearer {YOUR_TOKEN}"
```
**预期结果**：
- 返回绑定列表
- 每项包含：envName, envType, namespace, clusterName, replicas, cpuLimit等

#### ✓ 删除绑定API
```bash
curl -X DELETE 'http://localhost:8080/api/v1/app-env-bindings/1' \
  -H "Authorization: Bearer {YOUR_TOKEN}"
```
**预期结果**：
- code: 0
- message: "删除成功"

### 2. 数据库验证

```sql
-- 检查绑定表结构
DESC env_db.app_env_bindings;

-- 查询绑定数据
SELECT * FROM env_db.app_env_bindings WHERE is_deleted = 0;

-- 关联查询环境和集群信息
SELECT 
    aeb.id,
    aeb.app_id,
    aeb.env_id,
    e.env_name,
    e.env_type,
    e.namespace,
    c.cluster_name,
    aeb.replicas,
    aeb.cpu_limit,
    aeb.memory_limit
FROM env_db.app_env_bindings aeb
LEFT JOIN env_db.environments e ON aeb.env_id = e.id
LEFT JOIN infra_db.clusters c ON e.cluster_id = c.id
WHERE aeb.is_deleted = 0;
```

## 前端验证

### 1. 页面访问
- [ ] 访问 http://localhost
- [ ] 清空浏览器缓存（Cmd+Shift+R）
- [ ] 登录系统

### 2. 应用详情页
- [ ] 进入"应用管理"
- [ ] 选择一个应用
- [ ] 点击"详情"按钮
- [ ] 找到"环境绑定"卡片

### 3. 绑定环境功能
- [ ] 点击【绑定环境】按钮
- [ ] 对话框正常弹出
- [ ] 环境下拉框加载数据
- [ ] 选择一个环境
- [ ] 环境信息预览显示（集群、命名空间）
- [ ] 配置副本数、CPU、内存
- [ ] 点击【确定】
- [ ] 提示"绑定成功"
- [ ] 绑定列表自动刷新

### 4. 绑定列表显示
- [ ] 表格显示绑定的环境
- [ ] 显示环境名称
- [ ] 显示环境类型（dev/test/prod等）
- [ ] 显示命名空间
- [ ] 显示集群名称
- [ ] 显示配置状态
- [ ] 显示创建时间

### 5. 解绑功能
- [ ] 点击【解绑】按钮
- [ ] 弹出确认对话框
- [ ] 确认后解绑成功
- [ ] 列表自动刷新

### 6. 配置功能
- [ ] 点击【配置】按钮
- [ ] 跳转到环境配置页面

## 边界情况测试

### 1. 重复绑定检查
- [ ] 尝试绑定已经绑定过的环境
- [ ] 预期：提示"该应用已绑定此环境"

### 2. 空环境列表
- [ ] 如果没有可用环境
- [ ] 预期：环境下拉框显示"暂无数据"

### 3. 资源配置验证
- [ ] 副本数：1-100
- [ ] CPU/内存格式正确（如 100m, 512Mi）

### 4. 网络异常处理
- [ ] 模拟网络错误
- [ ] 预期：显示友好的错误提示

## 集成测试

### 1. 与部署流程集成
- [ ] 在应用部署页面
- [ ] 环境筛选是否生效
- [ ] 部署时是否使用绑定的配置

### 2. 与命名空间隔离集成
- [ ] 部署到不同环境
- [ ] 验证是否使用正确的namespace
- [ ] 验证是否部署到正确的cluster

## 性能测试

### 1. 查询性能
- [ ] 绑定列表查询响应时间 < 100ms
- [ ] 关联查询（环境+集群）响应时间 < 200ms

### 2. 并发测试
- [ ] 多用户同时绑定环境
- [ ] 验证数据一致性

## 文档验证

- [ ] APP_ENV_BINDING_GUIDE.md 完整准确
- [ ] APP_ENV_BINDING_IMPLEMENTATION.md 总结全面
- [ ] test_app_env_binding.sh 可正常执行

## 最终验证

### 完整流程验证
1. [ ] 创建应用
2. [ ] 创建环境（指定命名空间和集群）
3. [ ] 在应用详情页绑定环境
4. [ ] 配置资源限制
5. [ ] 验证绑定列表显示正确
6. [ ] （可选）部署应用到绑定的环境
7. [ ] 验证部署使用正确的命名空间

## 验证结果

| 功能模块 | 测试状态 | 备注 |
|---------|---------|------|
| 后端API | ⏳ 待验证 | env-service重新构建中 |
| 数据库 | ✅ 通过 | 表结构正确 |
| 前端UI | ⏳ 待验证 | 需要浏览器测试 |
| 文档 | ✅ 完成 | 已创建完整文档 |
| 集成 | ⏳ 待验证 | 需要完整流程测试 |

## 问题记录

| 问题 | 状态 | 解决方案 |
|------|------|---------|
| clusterName为null | ⏳ 修复中 | 重新构建env-service |
| 前端缓存 | ✅ 已知 | 清空浏览器缓存 |
| 跨库查询 | ✅ 已解决 | 添加clusterDB连接 |

---

**验证负责人**：_________
**验证日期**：2026-06-01
**签字**：_________
