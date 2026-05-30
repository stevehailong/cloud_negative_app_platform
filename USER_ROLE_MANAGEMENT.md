# 用户注册和角色管理功能说明

## 功能概述

系统已实现完整的用户注册和角色管理功能，包括：

1. **用户注册** - 新用户可以自主注册账号
2. **用户管理** - 管理员可以查看和管理所有用户
3. **角色分配** - 管理员可以为用户分配角色
4. **用户状态管理** - 管理员可以启用/禁用用户账号

## 角色体系

系统预定义了 4 种角色：

| 角色ID | 角色名称 | 角色编码 | 说明 |
|--------|---------|---------|------|
| 1 | 超级管理员 | SUPER_ADMIN | 拥有所有权限 |
| 2 | 项目管理员 | PROJECT_ADMIN | 项目管理权限 |
| 3 | 开发人员 | DEVELOPER | 开发人员权限 |
| 4 | 运维人员 | OPS | 运维人员权限 |

**注意**：角色权限控制功能后续可以进一步细化。

## 使用指南

### 1. 用户注册

#### 访问注册页面

1. 打开浏览器访问: http://localhost
2. 在登录页面点击"立即注册"链接
3. 或直接访问: http://localhost/register

#### 填写注册信息

**必填字段**：
- **用户名**: 3-20个字符，只能包含字母、数字和下划线
- **邮箱**: 有效的邮箱地址（系统会验证格式和唯一性）
- **密码**: 至少6位字符
- **确认密码**: 必须与密码一致

**可选字段**：
- **真实姓名**: 用户的真实姓名
- **手机号**: 11位手机号码（格式验证）

#### 注册示例

```
用户名: zhangsan
邮箱: zhangsan@example.com
密码: 123456
确认密码: 123456
真实姓名: 张三
手机号: 13800138000
```

#### 注册后状态

- ✅ 用户账号创建成功，状态为"正常"
- ❌ **尚未分配角色**（需要管理员分配）
- ⚠️ 登录后可能无法访问某些功能，需等待管理员分配角色

### 2. 管理员操作

#### 访问用户管理页面

1. 使用管理员账号登录（admin / admin123）
2. 在左侧菜单点击"用户管理"
3. 或直接访问: http://localhost/users

#### 查看用户列表

用户列表显示信息：
- ID、用户名、邮箱、真实姓名
- 手机号、部门
- **已分配角色**（以标签形式显示）
- **账号状态**（正常/禁用）
- 创建时间

#### 搜索用户

在搜索框中输入关键词可搜索：
- 用户名
- 邮箱
- 真实姓名

回车或点击搜索按钮执行搜索。

#### 为用户分配角色

**步骤**：
1. 在用户列表中找到目标用户
2. 点击"分配角色"按钮
3. 在弹出的对话框中勾选要分配的角色
4. 可以同时分配多个角色
5. 点击"确定"保存

**示例**：为新注册的"张三"分配开发人员角色
```
用户名: zhangsan
真实姓名: 张三
选择角色: ☑ 开发人员 (DEVELOPER)
```

#### 启用/禁用用户

**禁用用户**：
1. 点击用户操作列的"禁用"按钮
2. 确认操作
3. 用户状态变为"禁用"，该用户将无法登录

**启用用户**：
1. 对已禁用的用户点击"启用"按钮
2. 确认操作
3. 用户状态恢复"正常"，可以正常登录

## API 接口说明

### 用户注册接口

**请求**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "zhangsan",
    "password": "123456",
    "email": "zhangsan@example.com",
    "realName": "张三",
    "phone": "13800138000"
  }'
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "注册成功"
  }
}
```

### 获取用户列表

**请求** (需要认证):
```bash
TOKEN="your_jwt_token"

curl -X GET "http://localhost:8080/api/v1/users?page=1&pageSize=10&keyword=zhang" \
  -H "Authorization: Bearer $TOKEN"
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 2,
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "realName": "张三",
        "phone": "13800138000",
        "status": 1,
        "createTime": "2026-05-28T10:30:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 为用户分配角色

**请求** (需要认证):
```bash
TOKEN="your_jwt_token"

curl -X POST http://localhost:8080/api/v1/users/assign-roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 2,
    "roleIds": [3]
  }'
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "角色分配成功"
  }
}
```

### 获取用户角色

**请求** (需要认证):
```bash
TOKEN="your_jwt_token"

curl -X GET http://localhost:8080/api/v1/users/2/roles \
  -H "Authorization: Bearer $TOKEN"
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 3,
      "name": "开发人员",
      "code": "DEVELOPER",
      "description": "开发人员权限"
    }
  ]
}
```

### 更新用户状态

**请求** (需要认证):
```bash
TOKEN="your_jwt_token"

# 禁用用户（status=0）
curl -X PUT http://localhost:8080/api/v1/users/2/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": 0}'

# 启用用户（status=1）
curl -X PUT http://localhost:8080/api/v1/users/2/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": 1}'
```

### 获取角色列表

**请求** (需要认证):
```bash
TOKEN="your_jwt_token"

curl -X GET http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $TOKEN"
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "超级管理员",
      "code": "SUPER_ADMIN",
      "description": "拥有所有权限"
    },
    {
      "id": 2,
      "name": "项目管理员",
      "code": "PROJECT_ADMIN",
      "description": "项目管理权限"
    },
    {
      "id": 3,
      "name": "开发人员",
      "code": "DEVELOPER",
      "description": "开发人员权限"
    },
    {
      "id": 4,
      "name": "运维人员",
      "code": "OPS",
      "description": "运维人员权限"
    }
  ]
}
```

## 完整测试流程

### 1. 测试用户注册

```bash
# 注册新用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "lisi",
    "password": "123456",
    "email": "lisi@example.com",
    "realName": "李四"
  }'
```

### 2. 尝试用新用户登录

```bash
# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "lisi",
    "password": "123456"
  }'
```

**预期结果**：登录成功，获得 token

### 3. 管理员登录并查看新用户

```bash
# 管理员登录
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

echo "Admin Token: $ADMIN_TOKEN"

# 查看用户列表
curl -X GET "http://localhost:8080/api/v1/users?page=1&pageSize=10" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

### 4. 为新用户分配角色

```bash
# 获取新用户的ID（假设是2）
USER_ID=2

# 分配开发人员角色（角色ID=3）
curl -X POST http://localhost:8080/api/v1/users/assign-roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"userId\": $USER_ID,
    \"roleIds\": [3]
  }" | jq .
```

### 5. 验证角色分配

```bash
# 查看用户的角色
curl -X GET "http://localhost:8080/api/v1/users/$USER_ID/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

### 6. 测试禁用用户

```bash
# 禁用用户
curl -X PUT "http://localhost:8080/api/v1/users/$USER_ID/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": 0}' | jq .

# 尝试用被禁用的用户登录（应该失败）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"lisi","password":"123456"}' | jq .
```

**预期结果**：登录失败，提示"用户已被禁用"

### 7. 重新启用用户

```bash
# 启用用户
curl -X PUT "http://localhost:8080/api/v1/users/$USER_ID/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": 1}' | jq .

# 再次尝试登录（应该成功）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"lisi","password":"123456"}' | jq .
```

## 前端页面访问

### 注册页面
- URL: http://localhost/register
- 功能: 新用户自主注册

### 登录页面
- URL: http://localhost/login
- 功能: 用户登录，底部有"立即注册"链接

### 用户管理页面
- URL: http://localhost/users
- 功能: 查看用户列表、分配角色、启用/禁用用户
- 权限: 需要登录后访问（建议使用管理员账号）

## 数据库表结构

### users 表（用户信息）
```sql
SELECT id, username, email, real_name, status, created_at 
FROM iam_db.users;
```

### roles 表（角色定义）
```sql
SELECT * FROM iam_db.roles;
```

### user_roles 表（用户角色关联）
```sql
SELECT ur.*, u.username, r.name as role_name
FROM iam_db.user_roles ur
LEFT JOIN iam_db.users u ON ur.user_id = u.id
LEFT JOIN iam_db.roles r ON ur.role_id = r.id;
```

## 注意事项

1. **新注册用户无角色**: 用户注册后默认没有角色，需要管理员手动分配
2. **邮箱唯一性**: 每个邮箱只能注册一次
3. **用户名唯一性**: 用户名不能重复
4. **密码安全**: 密码使用 bcrypt 加密存储，不可逆
5. **Token有效期**: JWT token 有效期为 24 小时（86400秒）
6. **权限控制**: 当前所有认证用户都可以访问用户管理接口，建议后续添加角色权限验证

## 后续改进建议

1. **邮箱验证**: 注册时发送验证邮件确认邮箱
2. **审批流程**: 新用户注册后需要管理员审批才能激活
3. **密码强度**: 增加密码强度要求（大小写、数字、特殊字符）
4. **权限细化**: 为不同角色配置具体的接口访问权限
5. **操作日志**: 记录用户的操作历史
6. **批量操作**: 支持批量分配角色、批量禁用用户
7. **用户详情页**: 展示用户的详细信息和活动记录
8. **角色管理页**: 支持创建、编辑、删除自定义角色

## 文件清单

### 后端文件
- `backend/internal/auth/handler/auth_handler.go` - 认证和用户管理Handler
- `backend/internal/auth/service/auth_service.go` - 认证和用户管理Service
- `backend/internal/auth/repository/user_repository.go` - 用户数据访问层
- `backend/internal/auth/repository/role_repository.go` - 角色数据访问层（新增方法）
- `backend/internal/auth/router/router.go` - 路由配置

### 前端文件
- `frontend/src/views/auth/Register.vue` - 注册页面（新建）
- `frontend/src/views/user/UserManagement.vue` - 用户管理页面（新建）
- `frontend/src/api/user.js` - 用户管理API接口（新建）
- `frontend/src/api/auth.js` - 认证API接口（已存在）
- `frontend/src/router/index.js` - 路由配置（更新）
- `frontend/src/views/Login.vue` - 登录页面（添加注册链接）
- `frontend/src/layouts/MainLayout.vue` - 主布局（添加用户管理菜单）

## 状态

✅ **功能已完成并部署**

- ✅ 用户注册页面
- ✅ 用户管理页面
- ✅ 角色分配功能
- ✅ 用户状态管理
- ✅ 后端接口实现
- ✅ 前端页面实现
- ✅ Docker 容器已重启并加载新代码

可以立即访问 http://localhost 进行测试！
