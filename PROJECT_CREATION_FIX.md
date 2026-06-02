# 项目创建功能修复文档

## 修复时间
2026-06-02

## 问题描述

### 现象
- 用户在前端创建项目后,点击"确定"按钮,项目列表中不显示新创建的项目
- 数据库中实际已创建项目记录,但前端查询不到

### 根本原因

1. **数据库连接问题**
   - project-service 只连接了 `org_db` 数据库
   - 权限检查需要查询 `iam_db.user_roles` 和 `iam_db.roles` 表
   - 导致权限检查失败,报错: `Table 'org_db.user_roles' doesn't exist`

2. **项目可见性问题**
   - 新建项目默认 `visibility='private'` (私有)
   - 创建时未设置 `owner_user_id`,值为 NULL
   - 未自动添加创建者为项目成员
   - 导致非管理员用户看不到自己创建的项目

3. **权限过滤逻辑**
   - `ListAccessible` 方法的查询条件:
     - 项目是公开的 OR
     - 用户是项目所有者 OR  
     - 用户是项目成员
   - 三个条件都不满足时,项目不会出现在列表中

## 修复方案

### 1. 添加 iam_db 数据库连接

**修改文件**: `backend/cmd/project-service/main.go`

```go
// 连接到iam_db用于权限检查
iamDSN := "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
iamDB, err := database.InitDB(iamDSN, database.DefaultConnectionPoolConfig())
if err != nil {
    log.Fatal("连接iam_db失败:", err)
}

// 注册路由时传递两个数据库连接
router.RegisterRoutes(r, db, iamDB)
```

**修改文件**: `backend/internal/project/router/router.go`
- 更新 `RegisterRoutes` 函数签名接收 `iamDB` 参数

**修改文件**: `backend/internal/project/handler/project_handler.go`
- 更新 `NewProjectHandler` 函数接收 `iamDB` 参数

**修改文件**: `backend/internal/project/repository/project_repository.go`
- 在 `ProjectRepository` 结构体中添加 `iamDB` 字段
- 更新 `IsAdmin` 方法使用 `iamDB` 查询权限表

### 2. 自动设置项目所有者

**修改文件**: `backend/internal/project/handler/project_handler.go`

在 `CreateProject` 方法中添加:

```go
// 获取当前用户ID
userIDVal, _ := c.Get("userId")
userID, _ := userIDVal.(uint)

// 如果未指定owner，则设置为当前创建用户
ownerUserID := req.OwnerUserID
if ownerUserID == nil && userID > 0 {
    ownerUserID = &userID
}
```

### 3. 自动添加创建者为项目成员

**修改文件**: `backend/internal/project/handler/project_handler.go`

在项目创建成功后添加:

```go
// 自动将创建者添加为项目成员(owner角色)
if userID > 0 {
    member := &model.ProjectMember{
        ProjectID:  project.ID,
        UserID:     userID,
        RoleCode:   "owner",
        CreateTime: time.Now(),
        CreateBy:   &userID,
    }
    if err := h.projectRepo.AddMember(member); err != nil {
        // 添加成员失败不影响项目创建,只记录日志
        c.JSON(http.StatusOK, gin.H{
            "code":    0, 
            "message": "项目创建成功，但添加成员失败", 
            "data":    project,
        })
        return
    }
}
```

### 4. 修改默认可见性为公开

**修改文件**: `frontend/src/views/project/ProjectList.vue`

```javascript
const form = reactive({
  id: null,
  projectCode: '',
  projectName: '',
  tenantId: null,
  visibility: 'public',  // 从 'private' 改为 'public'
  description: '',
  status: 1
})

const resetForm = () => {
  // ...
  form.visibility = 'public'  // 从 'private' 改为 'public'
  // ...
}
```

### 5. 修复历史数据

```sql
-- 更新已有项目的可见性为公开
UPDATE org_db.projects SET visibility='public' WHERE id IN (1,2);

-- 更新已有项目的所有者
UPDATE org_db.projects SET owner_user_id=1 WHERE id IN (1,2);

-- 为已有项目添加成员记录
INSERT INTO org_db.project_members (project_id, user_id, role_code, create_time, create_by) 
VALUES 
  (1, 1, 'owner', NOW(), 1),
  (2, 1, 'owner', NOW(), 1);
```

## 验证结果

### 数据库验证

1. **项目表**:
```
id | project_code    | project_name | owner_user_id | visibility
1  | project-ai-001  | AI智能体      | 1             | public
2  | ai-project-001  | AI项目        | 1             | public
```

2. **项目成员表**:
```
id | project_id | user_id | username | role_code
1  | 1          | 1       | admin    | owner
2  | 2          | 1       | admin    | owner
```

### 功能验证

✅ 用户创建项目时:
- 自动设置 `owner_user_id` 为当前用户
- 自动添加创建者为项目成员,角色为 'owner'
- 默认可见性为 'public',所有用户可见

✅ 项目列表查询:
- 不再出现 `user_roles` 表不存在的错误
- 管理员可以看到所有项目
- 普通用户可以看到公开项目和自己参与的项目
- 项目所有者可以看到自己的私有项目

✅ 前端显示:
- 创建项目后立即在列表中显示
- 已有项目正常显示

## 影响范围

### 修改的文件
- `backend/cmd/project-service/main.go`
- `backend/internal/project/router/router.go`
- `backend/internal/project/handler/project_handler.go`
- `backend/internal/project/repository/project_repository.go`
- `frontend/src/views/project/ProjectList.vue`

### 需要重启的服务
- `my-cloud-project-service`
- `my-cloud-frontend`

## 部署步骤

```bash
# 1. 重新构建服务
cd /Users/hanhailong01/Downloads/my_cloud
docker-compose build project-service frontend

# 2. 重启服务
docker-compose restart project-service frontend

# 3. 验证服务启动
docker logs my-cloud-project-service --tail 20
docker logs my-cloud-frontend --tail 20

# 4. 修复历史数据(可选,如果有需要)
# 参见上述SQL语句
```

## 后续优化建议

1. **权限管理优化**
   - 添加更细粒度的项目角色(owner, maintainer, developer, reporter)
   - 实现基于角色的操作权限控制

2. **项目成员管理**
   - 在项目详情页增加成员管理功能
   - 支持添加/移除成员,修改成员角色

3. **项目可见性控制**
   - 在界面上提供可见性选择(public/internal/private)
   - internal: 同租户内可见
   - private: 仅成员可见

4. **审计日志**
   - 记录项目创建、成员变更等操作日志
   - 便于追踪和审计

5. **单元测试**
   - 为项目创建功能添加单元测试
   - 覆盖权限检查、成员添加等关键逻辑
