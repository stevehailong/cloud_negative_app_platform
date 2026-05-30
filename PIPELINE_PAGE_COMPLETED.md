# 流水线管理页面开发完成

## 功能概述

流水线管理页面已完成开发，提供完整的CI/CD流水线管理功能。

## 新增文件

### 后端API接口
- `frontend/src/api/pipeline.js` - 流水线相关API接口

### 前端页面组件
- `frontend/src/views/pipeline/PipelineList.vue` - 流水线列表主页面
- `frontend/src/views/pipeline/components/PipelineRuns.vue` - 执行记录组件

## 功能特性

### 1. 流水线列表管理

**功能点**:
- ✅ 流水线列表展示（支持分页）
- ✅ 按名称、类型、状态筛选
- ✅ 流水线启用/禁用状态切换
- ✅ 流水线新建、编辑、删除
- ✅ 一键触发流水线执行
- ✅ 查看最近执行状态

**数据展示**:
- 流水线名称（带图标）
- 流水线编码
- 流水线类型（构建/CI-CD/测试）
- CI工具（Jenkins/GitLab CI/GitHub Actions）
- 最近执行状态和时间
- 创建时间
- 启用状态

### 2. 流水线创建/编辑

**表单字段**:
- 流水线名称（必填）
- 流水线编码（必填，创建后不可修改）
- 关联应用（下拉选择）
- 流水线类型（构建/CI-CD/测试）
- CI工具（Jenkins/GitLab CI/GitHub Actions）
- 配置信息（JSON格式）
- 启用状态（开关）

**表单验证**:
- 所有必填项验证
- JSON配置格式验证

### 3. 执行记录管理

**功能点**:
- ✅ 执行记录列表展示
- ✅ 按状态、Git分支筛选
- ✅ 查看执行详情
- ✅ 查看执行日志（外部链接）
- ✅ 停止运行中的执行

**数据展示**:
- 执行编号
- 执行状态（成功/失败/运行中/等待中）
- 触发方式（手动/Webhook/定时/API）
- Git分支和Commit
- 执行时间
- 执行耗时

### 4. UI/UX特性

**页面布局**:
- 清晰的页面头部（标题+操作按钮）
- 筛选卡片（条件筛选）
- 数据表格（支持排序）
- 分页组件
- 抽屉式执行记录

**交互设计**:
- 状态标签彩色区分
- 图标化操作按钮
- 确认对话框（删除/触发）
- 加载状态提示
- 成功/失败消息提示

**响应式**:
- 固定操作列
- 最小宽度自适应
- 表格横向滚动

## 数据示例

### 流水线列表示例

```
| 流水线名称        | 编码                     | 类型  | CI工具    | 最近执行 | 状态 |
|------------------|--------------------------|-------|-----------|----------|------|
| 前端构建流水线    | PIPE-FRONTEND-001        | 构建  | jenkins   | ✅ 2小时前 | 启用 |
| 后端服务流水线    | PIPE-BACKEND-001         | CI/CD | jenkins   | 🔄 10分钟前| 启用 |
| 全栈应用流水线    | PIPE-FULLSTACK-001       | CI/CD | jenkins   | ✅ 4小时前 | 启用 |
| 数据处理流水线    | PIPE-DATA-001            | 构建  | gitlab-ci | ✅ 30分钟前| 启用 |
| 移动端构建流水线  | PIPE-MOBILE-001          | 构建  | jenkins   | ❌ 2天前  | 禁用 |
```

### 执行记录示例

```
| 执行编号                        | 状态  | 触发方式 | Git分支 | 执行时间            | 耗时   |
|--------------------------------|-------|----------|---------|---------------------|--------|
| PIPE-FRONTEND-001-20260528-001 | 成功  | 手动触发 | main    | 2026-05-28 10:00:00 | 5分0秒 |
| PIPE-FRONTEND-001-20260528-002 | 成功  | 手动触发 | main    | 2026-05-28 05:00:00 | 4分0秒 |
| PIPE-FRONTEND-001-20260527-001 | 失败  | Webhook  | develop | 2026-05-27 10:00:00 | 2分0秒 |
```

## API接口

### 流水线管理

```javascript
// 获取流水线列表
GET /api/v1/pipelines?page=1&pageSize=10&pipeline_name=xxx&pipeline_type=build

// 获取流水线详情
GET /api/v1/pipelines/:id

// 创建流水线
POST /api/v1/pipelines
{
  "pipeline_name": "前端构建流水线",
  "pipeline_code": "PIPE-FRONTEND-001",
  "app_id": 1,
  "pipeline_type": "build",
  "ci_tool": "jenkins",
  "config_json": {...},
  "enabled": 1
}

// 更新流水线
PUT /api/v1/pipelines/:id

// 删除流水线
DELETE /api/v1/pipelines/:id

// 触发执行
POST /api/v1/pipelines/:id/trigger
{
  "trigger_type": "manual"
}
```

### 执行记录

```javascript
// 获取执行记录列表
GET /api/v1/pipeline-runs?pipeline_id=1&page=1&pageSize=10&status=success

// 获取执行详情
GET /api/v1/pipeline-runs/:id

// 停止执行
POST /api/v1/pipeline-runs/:id/stop
```

## 测试数据

项目已包含5条流水线和10条执行记录的测试数据：

### 流水线数据
- 前端构建流水线 (build, jenkins)
- 后端服务流水线 (ci-cd, jenkins)
- 全栈应用流水线 (ci-cd, jenkins)
- 数据处理流水线 (build, gitlab-ci)
- 移动端构建流水线 (build, jenkins, 已禁用)

### 执行记录
- 包含成功、失败、运行中状态
- 涵盖手动触发、Webhook、定时触发
- 记录Git分支和Commit信息
- 包含执行时长统计

## 使用说明

### 访问页面

```
http://localhost/pipelines
```

### 基本操作流程

1. **查看流水线列表**
   - 访问流水线页面
   - 查看所有已配置的流水线
   - 使用筛选条件快速定位

2. **创建流水线**
   - 点击"新建流水线"按钮
   - 填写流水线信息
   - 选择关联应用
   - 配置流水线类型和CI工具
   - 输入JSON配置
   - 保存

3. **执行流水线**
   - 在流水线列表中找到目标流水线
   - 点击"执行"按钮
   - 确认触发执行
   - 查看执行状态

4. **查看执行记录**
   - 点击流水线的"记录"按钮
   - 查看所有历史执行记录
   - 可按状态、分支筛选
   - 点击"详情"查看详细信息
   - 点击"日志"查看执行日志

5. **管理流水线**
   - 编辑：修改流水线配置
   - 启用/禁用：通过开关快速切换状态
   - 删除：删除不需要的流水线

## 技术栈

### 前端框架
- Vue 3 (Composition API)
- Element Plus
- Vue Router
- Pinia
- Axios

### 工具库
- dayjs - 时间处理
- nprogress - 进度条

### 代码规范
- ESLint
- Prettier

## 部署说明

### 开发模式

```bash
cd frontend
npm install
npm run dev
```

### 生产构建

```bash
cd frontend
npm run build
```

构建产物在 `frontend/dist` 目录。

### Docker部署

```bash
# 重新构建前端镜像
docker-compose build frontend

# 重启前端容器
docker-compose restart frontend
```

## 注意事项

1. **JSON配置格式**
   - 配置信息必须是有效的JSON格式
   - 示例配置：
   ```json
   {
     "stages": [
       {
         "name": "构建",
         "steps": ["build", "test"]
       },
       {
         "name": "部署",
         "steps": ["deploy"]
       }
     ],
     "triggers": ["manual", "webhook"]
   }
   ```

2. **流水线编码规则**
   - 格式：PIPE-{应用}-{序号}
   - 示例：PIPE-FRONTEND-001
   - 创建后不可修改

3. **执行日志**
   - 日志地址存储在`log_url`字段
   - 点击"日志"按钮会在新标签页打开
   - 需要确保CI工具的日志地址可访问

4. **状态说明**
   - 成功(success): 执行成功完成
   - 失败(failed): 执行过程中出错
   - 运行中(running): 正在执行
   - 等待中(pending): 等待执行

## 后续优化建议

### 功能增强
1. 流水线配置可视化编辑器
2. 执行日志实时查看（WebSocket）
3. 执行阶段进度显示
4. 构建产物管理和下载
5. 流水线执行统计图表
6. 流水线模板功能
7. 批量操作支持

### 性能优化
1. 列表虚拟滚动
2. 图表懒加载
3. 执行记录分页优化
4. 缓存策略

### 体验优化
1. 拖拽排序
2. 快捷键支持
3. 搜索历史
4. 收藏/置顶功能

## 相关文档

- [Pipeline Service API文档](../docs/pipeline-service-api.md)
- [测试数据说明](../docs/test-data-guide.md)
- [部署指南](../README.md)

---

**开发完成时间**: 2026-05-28  
**页面路径**: `/pipelines`  
**状态**: ✅ 完成并可用
