# ⚠️ 重要提示：正确的服务部署方法

## 问题说明

在本次修复流水线镜像地址显示问题时，发现了一个关键的部署问题：

### ❌ 错误的部署方式
```bash
docker-compose build service-name
docker-compose restart service-name  # ❌ 错误！restart 不会使用新镜像
```

**问题**: `docker-compose restart` 只是重启现有容器，**不会**使用新构建的镜像！即使你刚刚 `build` 了新镜像，容器仍然使用旧的镜像。

### ✅ 正确的部署方式
```bash
docker-compose build service-name
docker-compose up -d service-name  # ✅ 正确！会重新创建容器
```

**原因**: `docker-compose up -d` 会检测到镜像已更新，自动重新创建容器。

## 验证方法

### 检查容器和镜像创建时间

```bash
# 查看镜像创建时间
docker images my_cloud-pipeline-service --format "{{.CreatedAt}}"

# 查看容器创建时间
docker inspect my-cloud-pipeline-service | grep Created | head -1
```

**正常情况**: 容器创建时间应该**晚于或等于**镜像创建时间

**异常情况**: 如果容器创建时间**早于**镜像创建时间，说明容器使用的是旧镜像！

### 示例

```bash
$ docker images my_cloud-pipeline-service --format "{{.CreatedAt}}"
2026-06-01 21:46:07 +0800 CST  # 镜像是今天 21:46 创建的

$ docker inspect my-cloud-pipeline-service | grep Created | head -1
"Created": "2026-05-30T06:52:37.496535587Z",  # ❌ 容器还是 5月30日的旧版本！
```

这种情况下，即使代码修改并重新构建了镜像，容器仍然运行的是旧代码。

## 完整的部署流程

### 1. 单个服务部署

```bash
# 方案A: 使用 docker-compose up (推荐)
docker-compose build <service-name>
docker-compose up -d <service-name>

# 方案B: 使用 make 命令
make deploy-service SERVICE=<service-name>

# 方案C: 手动重新创建
docker-compose stop <service-name>
docker-compose rm -f <service-name>
docker-compose build <service-name>
docker-compose up -d <service-name>
```

### 2. 所有服务部署

```bash
docker-compose build
docker-compose up -d
```

### 3. 强制重建（忽略缓存）

```bash
docker-compose build --no-cache <service-name>
docker-compose up -d <service-name>
```

## 本次问题的根本原因

### 时间线

1. **21:40** - 第一次修复，修改了 `ListPipelineRuns` 方法
2. **21:40** - 使用 `docker-compose build && restart` 部署
3. **21:40** - ❌ 用户反馈仍然没有显示
4. **21:46** - 第二次修复，修改了 `ListAllPipelineRuns` 方法（正确的接口）
5. **21:46** - 再次使用 `docker-compose build && restart` 部署
6. **21:46** - ❌ 用户反馈仍然没有显示
7. **21:52** - 检查发现容器创建时间是 5月30日，镜像是 21:46
8. **21:53** - 使用 `docker-compose up -d` 重新创建容器
9. **21:53** - ✅ 容器使用新镜像，问题解决

### 教训

1. **必须使用 `docker-compose up -d` 而不是 `restart`**
2. 部署后必须验证容器和镜像时间是否匹配
3. 如果多次部署仍无效，检查容器是否真的使用了新镜像

## 更新的 Makefile

已更新 Makefile 中的 `deploy-service` 命令，使用正确的方式：

```makefile
deploy-service:
	@echo "📦 部署 $(SERVICE)..."
	@docker-compose build $(SERVICE)
	@docker-compose up -d $(SERVICE)  # ✅ 使用 up -d 而不是 restart
	@echo "⏳ 等待服务启动..."
	@sleep 5
	@./health_check.sh
	@echo "✅ $(SERVICE) 部署完成"
```

## 快速参考

| 命令 | 是否使用新镜像 | 适用场景 |
|------|----------------|----------|
| `docker-compose restart` | ❌ 否 | 只是重启服务，不更新代码 |
| `docker-compose up -d` | ✅ 是 | 代码更新后部署 |
| `docker-compose up -d --force-recreate` | ✅ 是 | 强制重新创建容器 |
| `docker-compose stop && rm && up -d` | ✅ 是 | 完全清理后重建 |

## 检查清单

部署后务必检查：

- [ ] 容器创建时间 ≥ 镜像创建时间
- [ ] 服务日志没有错误
- [ ] 健康检查通过
- [ ] API 返回预期字段
- [ ] 前端功能正常

## 相关命令

```bash
# 查看所有服务的镜像和容器时间
for service in $(docker-compose ps --services); do
  echo "=== $service ==="
  docker images my_cloud-$service --format "镜像: {{.CreatedAt}}"
  docker inspect my-cloud-$service | grep -A1 Created | head -2
  echo ""
done

# 一键重新创建所有服务（慎用）
docker-compose up -d --force-recreate

# 查看哪些容器需要重新创建
docker-compose ps --all
```

---

**总结**: 永远使用 `docker-compose up -d` 而不是 `restart` 来部署代码更新！
