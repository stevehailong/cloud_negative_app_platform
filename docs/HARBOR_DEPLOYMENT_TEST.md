# Harbor 测试部署指南

## 阶段1：部署Harbor（当前阶段）

### 准备工作 ✅
- [x] Harbor安装包已下载
- [x] 配置文件已创建
- [x] 数据目录已创建

### 手动部署步骤

#### 步骤1：准备Harbor
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

# 运行prepare脚本（生成docker-compose配置）
sudo ./prepare
```

**预期输出**: 生成 `docker-compose.yml` 和其他配置文件

#### 步骤2：加载Harbor镜像
```bash
# 加载Harbor镜像到Docker
sudo docker load -i harbor.v2.11.0.tar.gz
```

**预期输出**: 加载多个Harbor相关镜像

**预计时间**: 5-10分钟

#### 步骤3：启动Harbor
```bash
# 启动所有Harbor服务
sudo docker-compose up -d
```

**预期输出**: 启动以下容器
- harbor-core
- harbor-portal
- harbor-db
- harbor-redis
- harbor-jobservice
- registry
- registryctl
- nginx

**预计时间**: 2-3分钟

#### 步骤4：检查Harbor状态
```bash
# 查看容器状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

**预期结果**: 所有容器状态为 `Up` 或 `Up (healthy)`

#### 步骤5：访问Harbor UI
1. 打开浏览器
2. 访问: http://localhost:8093
3. 登录:
   - 用户名: `admin`
   - 密码: `Harbor12345`

**预期结果**: 能够成功登录Harbor管理界面

---

## 阶段2：基础测试

### 测试1：Docker登录
```bash
docker login localhost:8093 -u admin -p Harbor12345
```

**预期输出**: `Login Succeeded`

### 测试2：创建项目
在Harbor UI中：
1. 点击"新建项目"
2. 项目名称: `mycloud`
3. 访问级别: 私有
4. 点击"确定"

### 测试3：推送测试镜像
```bash
# 拉取nginx测试镜像
docker pull nginx:alpine

# 重新打标签
docker tag nginx:alpine localhost:8093/mycloud/nginx:test

# 推送到Harbor
docker push localhost:8093/mycloud/nginx:test
```

**预期结果**: 
- 推送成功
- 在Harbor UI的mycloud项目中能看到nginx镜像

### 测试4：拉取镜像
```bash
# 删除本地镜像
docker rmi localhost:8093/mycloud/nginx:test

# 从Harbor拉取
docker pull localhost:8093/mycloud/nginx:test
```

**预期结果**: 拉取成功

### 测试5：漏洞扫描
在Harbor UI中：
1. 进入mycloud项目
2. 点击nginx镜像
3. 查看"漏洞"标签页

**预期结果**: 显示漏洞扫描结果（可能需要等待几分钟）

---

## 阶段3：与现有registry并行运行

### 配置说明

**当前状态**:
- 旧Registry: `localhost:5001` ✅ 继续运行
- Harbor: `localhost:8093` ✅ 并行运行

**过渡期策略**:
1. 新应用使用Harbor（localhost:8093）
2. 旧应用继续使用旧Registry（localhost:5001）
3. 逐步迁移镜像到Harbor
4. 验证所有功能正常后停用旧Registry

### 在docker-compose.yml中保持两者并存

**当前配置**:
```yaml
# 旧Registry（保留）
registry:
  image: registry:2
  container_name: my-cloud-registry
  ports:
    - "5001:5000"
  # ...

# Harbor（新增）
# 已通过独立的docker-compose在harbor目录中运行
```

**不需要修改主docker-compose.yml**，Harbor独立运行在自己的目录中。

---

## 阶段4：逐步迁移

### 迁移计划

#### 第1周：测试环境
- [ ] 选择1-2个测试应用
- [ ] 将其镜像推送到Harbor
- [ ] 更新这些应用的部署配置使用Harbor
- [ ] 验证功能正常

#### 第2周：开发环境
- [ ] 迁移所有开发环境应用
- [ ] 更新CI/CD Pipeline使用Harbor
- [ ] 培训团队使用Harbor UI

#### 第3周：生产环境
- [ ] 迁移生产环境镜像
- [ ] 配置镜像扫描策略
- [ ] 配置镜像清理策略

#### 第4周：收尾
- [ ] 验证所有应用都使用Harbor
- [ ] 停用旧Registry
- [ ] 清理旧Registry数据

---

## 常用命令

### Harbor管理
```bash
# 进入Harbor目录
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 重启Harbor
docker-compose restart

# 停止Harbor
docker-compose stop

# 启动Harbor
docker-compose start

# 完全停止并删除（保留数据）
docker-compose down
```

### 镜像操作
```bash
# 登录Harbor
docker login localhost:8093 -u admin -p Harbor12345

# 推送镜像
docker tag myimage:v1.0.0 localhost:8093/mycloud/myimage:v1.0.0
docker push localhost:8093/mycloud/myimage:v1.0.0

# 拉取镜像
docker pull localhost:8093/mycloud/myimage:v1.0.0

# 同时使用两个registry
docker tag myimage:v1.0.0 localhost:5001/myimage:v1.0.0  # 旧registry
docker tag myimage:v1.0.0 localhost:8093/mycloud/myimage:v1.0.0  # Harbor
```

---

## 故障排查

### 问题1：Harbor容器启动失败
```bash
# 查看日志
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose logs harbor-core

# 检查端口占用
lsof -i :8093

# 检查磁盘空间
df -h
```

### 问题2：无法推送镜像
```bash
# 检查是否登录
docker login localhost:8093

# 检查项目是否存在
# 访问Harbor UI查看

# 查看详细错误
docker push localhost:8093/mycloud/myimage:v1.0.0 --debug
```

### 问题3：漏洞扫描不工作
```bash
# 检查Trivy服务状态
docker-compose ps trivy

# 重启Trivy
docker-compose restart trivy

# 查看Trivy日志
docker-compose logs trivy
```

---

## 性能监控

### 资源使用
```bash
# 查看Harbor容器资源使用
docker stats $(docker ps --filter name=harbor -q)
```

### 存储使用
```bash
# 查看Harbor数据大小
du -sh /Users/hanhailong01/Downloads/my_cloud/harbor-data
```

---

## 下一步

1. ✅ **现在执行**: 运行"阶段1：部署Harbor"中的步骤
2. ⏳ **然后**: 完成"阶段2：基础测试"
3. ⏳ **接着**: 保持两个registry并行运行
4. ⏳ **最后**: 按照迁移计划逐步迁移

---

## 需要帮助？

如果遇到问题：
1. 查看Harbor日志：`docker-compose logs -f`
2. 检查Harbor状态：`docker-compose ps`
3. 重启Harbor：`docker-compose restart`
4. 查看本文档的"故障排查"部分

开始部署吧！ 🚀
