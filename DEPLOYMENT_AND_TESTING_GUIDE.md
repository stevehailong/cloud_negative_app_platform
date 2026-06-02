# 部署和测试流程规范

## 📋 部署前检查清单

在对关键服务（gateway, pipeline-service, deploy-service 等）进行修改后，必须执行以下步骤：

### 1. 代码修改

- [ ] 完成代码修改
- [ ] 本地编译通过（Go: `go build`, 前端: `npm run build`）
- [ ] 代码审查通过

### 2. 构建和部署

```bash
# 重新构建服务镜像
docker-compose build <service-name>

# 重启服务
docker-compose restart <service-name>

# 或者完整重建
docker-compose up -d --build <service-name>
```

### 3. 部署后验证 ⚠️ **必须执行**

#### 快速健康检查（30秒）

```bash
./health_check.sh
```

这会检查：
- 所有服务是否正常运行
- 关键 API 是否返回预期字段
- 数据库连接是否正常

#### 完整集成测试（2分钟）

```bash
./test_pipeline_image_url.sh
```

包含：
- 数据库层测试（表结构、数据完整性）
- 服务层测试（容器状态、健康检查）
- API 层测试（接口响应、字段验证）
- 前端集成测试（页面可访问性）

### 4. 手动验证

如果自动化测试通过，进行手动验证：

1. **浏览器访问**: http://localhost:3000
2. **清除缓存**: Ctrl+Shift+R (或 Cmd+Shift+R)
3. **验证功能**:
   - 导航到相关页面
   - 执行关键操作
   - 检查数据显示是否正确

### 5. 日志检查

如果发现问题，检查服务日志：

```bash
# 查看服务日志
docker-compose logs --tail=50 <service-name>

# 实时监控日志
docker-compose logs -f <service-name>

# 查看错误日志
docker-compose logs <service-name> | grep -i error
```

## 🔍 常见问题排查

### 问题1: 前端显示旧数据

**症状**: 修改后端代码，但前端仍显示旧数据

**原因**:
- 浏览器缓存
- 前端静态资源未更新
- 后端镜像未重建

**解决**:
```bash
# 1. 强制重建前端
docker-compose build --no-cache frontend
docker-compose restart frontend

# 2. 清除浏览器缓存
# Chrome/Edge: Ctrl+Shift+Delete
# Firefox: Ctrl+Shift+Del

# 3. 验证后端是否重新构建
docker-compose images | grep <service-name>
# 检查 Created 时间是否是最近的
```

### 问题2: API 返回数据缺少字段

**症状**: API 响应中缺少预期字段（如 imageUrl）

**原因**:
- 后端代码未正确修改
- 服务未重启
- 数据库表结构问题

**排查**:
```bash
# 1. 检查服务是否使用最新镜像
docker-compose ps <service-name>

# 2. 测试 API 响应
docker-compose exec -T gateway curl -s 'http://<service>:port/api/endpoint' | jq

# 3. 检查数据库
docker-compose exec mysql mysql -uroot -proot123456 <database> -e "DESCRIBE <table>;"
```

### 问题3: 数据库表不存在

**症状**: 服务日志显示 "Table doesn't exist"

**原因**:
- SQL 迁移脚本未执行
- GORM AutoMigrate 失败

**解决**:
```bash
# 1. 手动执行 SQL 脚本
docker-compose exec -T mysql mysql -uroot -proot123456 < sql/xx_table_name.sql

# 2. 重启服务触发 AutoMigrate
docker-compose restart <service-name>

# 3. 验证表已创建
docker-compose exec mysql mysql -uroot -proot123456 <db> -e "SHOW TABLES;"
```

## 📝 测试脚本说明

### health_check.sh
快速健康检查脚本，部署后必须执行：
- 检查所有服务容器状态
- 验证关键 API 响应
- 检查数据库连接

**用法**:
```bash
./health_check.sh
```

### test_pipeline_image_url.sh
完整的集成测试脚本，用于验证流水线镜像地址功能：
- 数据库层: 表结构、数据完整性
- 服务层: 容器健康、服务可用性
- API 层: 接口功能、字段验证
- 前端层: 页面可访问性

**用法**:
```bash
./test_pipeline_image_url.sh
```

### test_app_env_binding.sh
应用环境绑定功能测试

### e2e_test.sh
端到端测试脚本

## 🚀 推荐部署流程

```bash
# 1. 修改代码后
git status
git diff

# 2. 重新构建和部署
docker-compose build <service-name>
docker-compose up -d <service-name>

# 3. 等待服务启动（约5秒）
sleep 5

# 4. 执行健康检查
./health_check.sh

# 5. 如果健康检查失败，查看日志
docker-compose logs --tail=50 <service-name>

# 6. 如果健康检查通过，执行完整测试
./test_pipeline_image_url.sh  # 或其他相关测试

# 7. 手动验证前端
# 打开浏览器，强制刷新（Ctrl+Shift+R）
# 验证功能是否正常

# 8. 确认无误后提交代码
git add .
git commit -m "fix: xxx"
git push
```

## ⚠️ 重要提示

1. **永远不要跳过测试**: 即使是"小改动"也可能影响其他功能
2. **使用自动化测试**: 手动测试容易遗漏问题
3. **检查日志**: 部署后立即检查服务日志，确保没有报错
4. **清除缓存**: 前端验证时务必清除浏览器缓存
5. **备份数据**: 修改数据库结构前先备份

## 📊 测试覆盖范围

| 功能模块 | 测试脚本 | 覆盖内容 |
|---------|---------|---------|
| 流水线镜像地址 | test_pipeline_image_url.sh | 数据库、API、前端显示 |
| 应用环境绑定 | test_app_env_binding.sh | CRUD 操作、数据关联 |
| 端到端流程 | e2e_test.sh | 完整业务流程 |
| 服务健康 | health_check.sh | 容器状态、基础功能 |

## 🔧 自定义测试

如需添加新的测试用例，参考现有脚本格式：

```bash
#!/bin/bash
# 测试: <功能描述>

echo "测试 <功能名称>..."

# 测试1: <测试内容>
if <测试命令>; then
    echo "✓ 测试1 通过"
else
    echo "✗ 测试1 失败"
    exit 1
fi

# 测试2: <测试内容>
# ...

echo "所有测试通过"
```

## 📚 相关文档

- [流水线镜像地址修复文档](PIPELINE_IMAGE_URL_FIX.md)
- [应用环境绑定实现文档](APP_ENV_BINDING_IMPLEMENTATION.md)
- [部署验证指南](DEPLOYMENT_VERIFICATION.md)
