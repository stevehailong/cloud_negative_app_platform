# 流水线权限问题已解决

## ❌ 原问题

访问流水线API时返回：
```json
{
    "code": 40301,
    "message": "无权限访问此资源",
    "requestId": "..."
}
```

## ✅ 解决方案

已为以下角色添加流水线权限：

### 1. SUPER_ADMIN（超级管理员）
- ✅ pipeline:view - 查看流水线
- ✅ pipeline:create - 创建流水线
- ✅ pipeline:edit - 编辑流水线
- ✅ pipeline:execute - 执行流水线
- ✅ pipeline:delete - 删除流水线

### 2. PROJECT_ADMIN（项目管理员）
- ✅ pipeline:view - 查看流水线
- ✅ pipeline:create - 创建流水线
- ✅ pipeline:edit - 编辑流水线
- ✅ pipeline:execute - 执行流水线
- ✅ pipeline:delete - 删除流水线

### 3. DEVELOPER（开发者）
- ✅ pipeline:view - 查看流水线
- ✅ pipeline:execute - 执行流水线

### 4. OPS（运维人员）
- ✅ pipeline:view - 查看流水线
- ✅ pipeline:create - 创建流水线
- ✅ pipeline:edit - 编辑流水线
- ✅ pipeline:execute - 执行流水线
- ✅ pipeline:delete - 删除流水线

## 🔄 使其生效的步骤

### 方式1: 重新登录（推荐）

1. **退出当前账号**
   - 点击右上角用户头像
   - 选择"退出登录"

2. **重新登录**
   - 使用管理员账号登录
   - 系统会重新加载权限

3. **访问流水线页面**
   - 点击侧边栏"流水线"菜单
   - 或直接访问 http://localhost/pipelines

### 方式2: 清除Token并刷新

1. **打开浏览器开发者工具**（F12）

2. **进入Application或存储标签**

3. **清除localStorage**
   ```javascript
   // 在Console中执行
   localStorage.clear()
   ```

4. **刷新页面**
   - 会自动跳转到登录页
   - 重新登录即可

### 方式3: 使用新的无痕窗口

1. 打开浏览器无痕模式
2. 访问 http://localhost
3. 登录后访问流水线页面

## 🧪 验证权限

### 测试API访问

```bash
# 1. 重新登录获取新token
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq .

# 2. 使用新token访问流水线API
TOKEN="新获取的token"
curl http://localhost/api/v1/pipelines \
  -H "Authorization: Bearer $TOKEN" \
  | jq .

# 预期返回：包含5条流水线数据
```

### 检查角色权限

```sql
-- 查看超级管理员的流水线权限
SELECT 
    r.name as role_name,
    p.code as permission_code,
    p.name as permission_name,
    p.path as api_path
FROM role_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.code = 'SUPER_ADMIN' 
  AND p.code LIKE 'pipeline:%'
ORDER BY p.code;
```

## 📋 权限说明

### pipeline:view
- **说明**: 查看流水线列表和详情
- **API路径**: GET /api/v1/pipelines/*
- **页面功能**: 查看流水线列表、搜索、分页

### pipeline:create
- **说明**: 创建新的流水线
- **API路径**: POST /api/v1/pipelines/
- **页面功能**: 新建流水线按钮和表单

### pipeline:edit
- **说明**: 编辑现有流水线
- **API路径**: PUT /api/v1/pipelines/*
- **页面功能**: 编辑按钮和表单

### pipeline:execute
- **说明**: 触发流水线执行
- **API路径**: POST /api/v1/pipelines/*/run/
- **页面功能**: 执行按钮

### pipeline:delete
- **说明**: 删除流水线
- **API路径**: DELETE /api/v1/pipelines/*
- **页面功能**: 删除按钮

## 🎯 不同角色的权限对比

| 功能 | 超级管理员 | 项目管理员 | 开发者 | 运维 | 访客 |
|------|-----------|-----------|--------|------|------|
| 查看流水线 | ✅ | ✅ | ✅ | ✅ | ❌ |
| 创建流水线 | ✅ | ✅ | ❌ | ✅ | ❌ |
| 编辑流水线 | ✅ | ✅ | ❌ | ✅ | ❌ |
| 执行流水线 | ✅ | ✅ | ✅ | ✅ | ❌ |
| 删除流水线 | ✅ | ✅ | ❌ | ✅ | ❌ |

## 🔍 故障排查

### 问题1: 重新登录后仍然无权限

**检查用户角色**:
```sql
-- 查看用户的角色
SELECT u.username, r.code as role_code, r.name as role_name
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.username = 'admin';
```

**解决**: 确保用户有正确的角色分配

### 问题2: API返回401而不是403

**原因**: Token过期或无效

**解决**: 
1. 清除localStorage
2. 重新登录
3. 获取新token

### 问题3: 部分功能无权限

**原因**: 用户角色权限不足

**解决**: 
1. 检查用户角色
2. 联系管理员提升权限
3. 或使用超级管理员账号

## 📝 添加自定义角色权限

如果需要为其他角色添加流水线权限：

```sql
-- 示例：给ID为99的角色添加查看权限
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT 99, id, NOW()
FROM permissions
WHERE code = 'pipeline:view';

-- 添加多个权限
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT 99, id, NOW()
FROM permissions
WHERE code IN ('pipeline:view', 'pipeline:execute');
```

## ✅ 验收清单

完成以下步骤确认问题已解决：

- [ ] 执行了权限添加SQL
- [ ] 退出登录
- [ ] 重新登录
- [ ] 访问流水线页面
- [ ] 能看到流水线列表
- [ ] 搜索功能正常
- [ ] 分页功能正常
- [ ] 无权限错误消失

## 🎊 预期结果

重新登录后，访问 http://localhost/pipelines 应该能看到：

```
┌─────────────────────────────────────────────────────┐
│  流水线管理                                          │
│  CI/CD流水线配置与执行                               │
├─────────────────────────────────────────────────────┤
│  流水线名称: [___________]  [查询] [重置]          │
├─────────────────────────────────────────────────────┤
│ 名称          │编码          │类型 │工具    │状态  │
├─────────────────────────────────────────────────────┤
│ 前端构建流水线 │PIPE-FRONT-001│build│jenkins │启用  │
│ 后端服务流水线 │PIPE-BACK-001 │ci-cd│jenkins │启用  │
│ ...                                                  │
└─────────────────────────────────────────────────────┘
```

不再出现"无权限访问此资源"的错误！

---

**问题状态**: ✅ 已解决  
**需要操作**: 重新登录  
**解决时间**: 2026-05-28 20:50
