# 开发规范与经验总结 (Rules)

> 本文档记录项目开发过程中的规范要求和踩坑经验。  
> 每次编写代码前必须阅读本文件，避免重复犯错。

---

## 一、时间格式规范

### 规则

- 所有前端页面的时间展示**必须**使用 `@/utils/time.js` 中的统一格式化函数
- 时间格式统一为：`YYYY-MM-DD HH:mm:ss`（如 `2026-05-29 15:30:45`）
- 日期格式统一为：`YYYY-MM-DD`
- 耗时格式统一为：中文可读（如 `5分30秒`）

### 可用函数

| 函数 | 用途 | 示例输出 |
|------|------|---------|
| `formatTime(time)` | 完整时间 | `2026-05-29 15:30:45` |
| `formatDate(time)` | 仅日期 | `2026-05-29` |
| `formatDuration(seconds)` | 耗时秒数 | `5分30秒` |

### 用法

```vue
<script setup>
import { formatTime, formatDuration } from '@/utils/time'
</script>

<template>
  <!-- 正确：使用 template slot + formatTime -->
  <el-table-column label="创建时间" width="180">
    <template #default="{ row }">
      {{ formatTime(row.createTime) }}
    </template>
  </el-table-column>

  <!-- 错误：直接用 prop 展示原始时间戳 -->
  <!-- <el-table-column prop="createTime" label="创建时间" /> -->
</template>
```

### 踩坑记录

- **问题**：`<el-table-column prop="createTime">` 直接展示后端返回的 ISO 时间字符串（如 `2026-05-29T15:30:45+08:00`），用户体验差且格式不统一
- **根因**：新增页面时忘记导入和使用 formatTime
- **修复**：所有时间列必须用 `<template #default>` + `formatTime()` 包裹

---

## 二、编码规范

### 前端编码规范

1. **API 函数**：统一放在 `@/api/` 目录，按模块分文件
2. **组件命名**：PascalCase，文件名与组件名一致
3. **公共工具**：统一放 `@/utils/`，避免在组件内定义可复用函数
4. **状态展示**：使用统一的 helper 函数（如 `statusLabel()`、`statusType()`），避免在模板中写复杂条件
5. **字段命名**：后端返回的 JSON 字段为 **camelCase**（如 `gitBranch`、`startTime`），前端引用时必须使用 camelCase，**禁止使用 snake_case**（如 `git_branch`、`start_time`）

### 字段命名踩坑

- **问题**：前端模板中用 `row.git_branch`、`row.start_time` 等 snake_case 引用后端字段，导致显示为空
- **根因**：后端 Go struct 使用 `json:"gitBranch"` 标签，返回的是 camelCase，前端误用 snake_case
- **规则**：前端所有数据绑定统一使用 camelCase（与后端 JSON tag 一致）

### 后端编码规范

1. **服务间通信**：通过 HTTP API 调用，使用 Docker 内部 DNS（如 `http://release-service:8086`）
2. **数据库查询**：GORM 查询返回记录不存在时，**必须**返回 `nil` 而非空结构体指针
3. **错误处理**：`errors.Is(err, gorm.ErrRecordNotFound)` 判断记录不存在
4. **Git URL 格式**：统一为 `https://domain/group/project.git`，不使用冒号分隔（`domain:group/project`）

---

## 三、权限管理规范

### 新增功能模块时的权限检查清单

新增 API 模块后**必须**完成以下步骤，否则前端访问会返回 403：

1. **添加权限记录**：在 `sql/04_permissions.sql` 的 `permissions` 表中添加对应权限
2. **分配角色权限**：在同一 SQL 文件中为各角色分配新权限
3. **使用角色 code 匹配**：`SELECT r.id FROM roles r WHERE r.code = 'SUPER_ADMIN'`，不要硬编码角色 ID
4. **执行 SQL**：在运行中的 MySQL 实例中执行 INSERT 语句
5. **验证**：用 curl + token 验证 API 可访问

### 踩坑记录

- **问题**：新增发布管理模块后，前端访问 `/api/v1/releases` 返回 403
- **根因**：gateway 的 PermissionCheck 中间件要求 permissions 表有匹配的 path+method 记录，但新模块未添加权限数据
- **额外坑**：SQL 中使用了硬编码角色 ID（1-5），但实际数据库中角色 ID 是自增的（9-13），导致权限分配到不存在的角色
- **修复**：改为通过 `roles.code` 子查询获取角色 ID

### 权限 path 匹配规则

| 模式 | 示例 | 匹配范围 |
|------|------|---------|
| 精确匹配 | `/api/v1/releases` | 仅匹配该路径 |
| 尾部通配 | `/api/v1/releases/*` | 匹配 `/api/v1/releases` 及其所有子路径 |
| 参数通配 | `/api/v1/releases/:id/approve` | 匹配任意 ID 值 |

---

## 四、服务间通信与 502 问题

### 常见 502 原因

| 原因 | 表现 | 排查方式 |
|------|------|---------|
| 服务未启动 | 网关转发超时 | `docker compose ps` 检查服务状态 |
| 服务启动失败 | 容器反复重启 | `docker compose logs <service>` 查日志 |
| 端口配置错误 | 连接被拒绝 | 检查 docker-compose.yml 端口映射 |
| 服务间 DNS 解析失败 | 无法连接目标 | 确认 docker-compose 网络配置 |
| 编译错误导致镜像构建失败 | 旧容器运行旧代码 | `docker compose build <service>` 重新构建 |

### 防止 502 的开发习惯

1. **修改后端代码后**必须先在本地编译通过：`cd backend && go build ./cmd/<service>/`
2. **重建服务**时使用 `docker compose up -d --build <service>`
3. **新增服务间调用**时确认目标服务已在 docker-compose.yml 中定义且端口正确
4. **网关路由**新增后需确认 `backend/internal/gateway/router/router.go` 中代理地址正确

### 踩坑记录

- **问题**：pipeline-service 调用 release-service 时偶发 502
- **根因**：release-service 重启后 pipeline-service 使用了缓存的连接，TCP 连接已断开
- **建议**：服务间 HTTP 调用使用短连接或设置合理的超时和重试

---

## 五、Webhook 集成规范

### GitLab Webhook 配置

1. **URL 格式**：`http://<gateway-host>/hooks/gitlab`
2. **项目匹配逻辑**：从 webhook payload 的 `project.path_with_namespace` 提取**项目名**（最后一段），非完整路径
3. **应用 repoUrl 格式**：必须是标准 URL 格式 `https://domain/group/project.git`

### 踩坑记录

- **问题**：Webhook 回调后无法匹配到对应流水线
- **根因**：GitLab API 按 project name 搜索，代码却传入了完整路径 `group/project`
- **修复**：从路径中提取最后一段作为项目名称搜索

---

## 六、数据库规范

### 模型字段新增

1. 新增字段后需**同时更新**：
   - Go model struct
   - 数据库 migration 或 AutoMigrate
   - 前端 API 请求/响应结构体
2. GORM AutoMigrate 仅添加列不删除列，字段改名需手动 migration

### Repository 层

- `GetByXxx()` 方法在记录不存在时**必须返回 nil**，不能返回空结构体指针
- 错误示例：`return &model, err`（model 是栈分配，永远非 nil）
- 正确示例：
  ```go
  func (r *Repo) GetByCode(code string) (*Model, error) {
      var m Model
      err := r.db.Where("code = ?", code).First(&m).Error
      if err != nil {
          return nil, err
      }
      return &m, nil
  }
  ```

---

## 七、前端路由与菜单

### 新增页面检查清单

1. 在 `src/views/<module>/` 创建 Vue 组件
2. 在 `src/router/index.js` 添加路由配置
3. 在 `src/layouts/MainLayout.vue` 添加菜单项和图标
4. 确认 gateway 已配置对应的代理路由
5. 确认 permissions 表有对应权限记录

---

## 八、构建与部署

### 开发流程

```bash
# 1. 后端编译验证（修改后必做）
cd backend && go build ./cmd/<service>/

# 2. 前端构建验证（修改后必做）
cd frontend && npm run build

# 3. 重建并部署单个服务
docker compose up -d --build <service-name>

# 4. 查看日志确认启动正常
docker compose logs <service-name> --tail 20
```

### 全量重建

```bash
docker compose build && docker compose up -d
```

---

> 文档版本：v1.0  
> 创建日期：2026-05-29  
> 维护说明：每次遇到新的规范问题或踩坑经验，及时补充到对应章节。
