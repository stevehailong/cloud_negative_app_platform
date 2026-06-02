# Harbor 部署命令清单

## 📝 在终端中依次执行以下命令

### 步骤1: 准备Harbor
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo ./prepare
```
**预期**: 生成配置文件，显示 "✔ ----Harbor has been installed and started successfully.----"

---

### 步骤2: 加载镜像
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo docker load -i harbor.v2.11.0.tar.gz
```
**预期**: 加载多个镜像，大约5-10分钟

---

### 步骤3: 启动Harbor
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo docker-compose up -d
```
**预期**: 创建并启动所有Harbor容器

---

### 步骤4: 检查状态
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose ps
```
**预期**: 所有容器状态为 `Up` 或 `Up (healthy)`

---

### 步骤5: 访问Harbor
打开浏览器访问: **http://localhost:8093**
- 用户名: `admin`
- 密码: `Harbor12345`

---

## 🧪 基础测试

### 1. 创建项目
在Harbor UI中创建项目 `mycloud`

### 2. Docker登录
```bash
docker login localhost:8093 -u admin -p Harbor12345
```

### 3. 推送测试镜像
```bash
docker pull nginx:alpine
docker tag nginx:alpine localhost:8093/mycloud/nginx:test
docker push localhost:8093/mycloud/nginx:test
```

### 4. 验证
在Harbor UI中查看 mycloud 项目，应该能看到 nginx:test 镜像

---

## ✅ 完成标志

- [ ] Harbor UI 可以访问
- [ ] 可以登录
- [ ] 可以创建项目
- [ ] 可以推送镜像
- [ ] 可以在UI中看到镜像

---

## 🔧 常用命令

```bash
# 进入Harbor目录
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 重启
docker-compose restart

# 停止
docker-compose stop

# 启动
docker-compose start
```

---

## 📞 遇到问题？

查看日志：
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose logs -f harbor-core
```

检查端口：
```bash
lsof -i :8093
```

---

## 下一步

部署完成后，查看迁移计划：
```bash
cat /Users/hanhailong01/Downloads/my_cloud/docs/HARBOR_MIGRATION_PLAN.md
```
