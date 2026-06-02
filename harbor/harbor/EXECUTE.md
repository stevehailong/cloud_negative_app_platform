# 🚀 Harbor 部署 - 立即执行

## ✅ 准备工作已完成

- [x] Harbor安装包已下载 (633MB)
- [x] harbor.yml配置已创建
- [x] 数据目录已创建
- [x] 部署脚本已准备

---

## 📝 执行步骤

### 方式1：自动化脚本（推荐） ⭐

**在终端中复制粘贴并执行：**

```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor && ./deploy.sh
```

这个脚本会自动完成所有步骤。

---

### 方式2：使用Harbor官方安装脚本

```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo ./install.sh --with-trivy
```

这是Harbor官方的一键安装，包含漏洞扫描功能。

---

### 方式3：手动分步执行

```bash
# 进入目录
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

# 步骤1: 准备配置
sudo ./prepare

# 步骤2: 加载镜像（5-10分钟）
sudo docker load -i harbor.v2.11.0.tar.gz

# 步骤3: 启动服务
sudo docker-compose up -d

# 步骤4: 检查状态
docker-compose ps
```

---

## 🧪 安装完成后的测试

### 1. 访问Harbor

打开浏览器: **http://localhost:8093**

登录:
- 用户名: `admin`
- 密码: `Harbor12345`

### 2. 创建项目

在Harbor UI中创建项目 `mycloud`

### 3. 推送测试镜像

```bash
# 登录Harbor
docker login localhost:8093 -u admin -p Harbor12345

# 拉取测试镜像
docker pull nginx:alpine

# 打标签
docker tag nginx:alpine localhost:8093/mycloud/nginx:test

# 推送到Harbor
docker push localhost:8093/mycloud/nginx:test
```

### 4. 验证

在Harbor UI中查看 `mycloud` 项目，应该能看到 `nginx:test` 镜像。

---

## ⏰ 预计时间

- 准备配置: 1分钟
- 加载镜像: 5-10分钟
- 启动服务: 2-3分钟
- **总计: 10-15分钟**

---

## 🔍 检查安装状态

```bash
# 查看容器状态
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose ps

# 查看日志
docker-compose logs -f
```

**预期结果**: 所有容器状态为 `Up` 或 `Up (healthy)`

---

## 📞 遇到问题？

### 端口被占用
```bash
lsof -i :8093
# 停止占用端口的进程或修改harbor.yml中的端口
```

### Docker未运行
```bash
# 启动Docker Desktop
open -a Docker
```

### 查看错误日志
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose logs harbor-core
```

---

## 🎯 推荐执行

**立即在你的终端执行：**

```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo ./install.sh --with-trivy
```

然后等待10分钟，访问 http://localhost:8093

---

## 📚 更多信息

- 详细命令: `cat COMMANDS.md`
- 检查清单: `cat ../CHECKLIST.md`
- 迁移计划: `cat ../../docs/HARBOR_MIGRATION_PLAN.md`

---

开始执行吧！🚀
