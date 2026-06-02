# Harbor 快速开始指南

## 立即开始

### 选项A：自动部署（推荐）

**一键部署Harbor：**
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor
sudo ./setup-harbor.sh
```

等待5-10分钟后，访问 http://localhost:8093

---

### 选项B：手动部署

**步骤1：配置**
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
cp harbor.yml.tmpl harbor.yml
```

编辑 `harbor.yml`，修改以下内容：
```yaml
hostname: localhost
http:
  port: 8093
harbor_admin_password: Harbor12345
data_volume: /Users/hanhailong01/Downloads/my_cloud/harbor-data
```

**步骤2：安装**
```bash
sudo ./prepare
sudo ./install.sh --with-trivy
```

---

## 首次登录

1. 访问: http://localhost:8093
2. 用户名: `admin`
3. 密码: `Harbor12345`

---

## 快速测试

```bash
# 登录
docker login localhost:8093 -u admin -p Harbor12345

# 推送测试镜像
docker pull nginx:alpine
docker tag nginx:alpine localhost:8093/library/nginx:test
docker push localhost:8093/library/nginx:test
```

在Harbor UI中应该能看到这个镜像！

---

## 下一步

查看完整集成清单：
```bash
cat /Users/hanhailong01/Downloads/my_cloud/docs/HARBOR_INTEGRATION_CHECKLIST.md
```

---

## 常见问题

**Q: Harbor占用多少资源？**
A: 最少4GB内存，推荐8GB

**Q: 如何停止Harbor？**
A: `cd harbor/harbor && docker-compose stop`

**Q: 如何重启Harbor？**
A: `cd harbor/harbor && docker-compose restart`

**Q: 数据存储在哪里？**
A: `/Users/hanhailong01/Downloads/my_cloud/harbor-data`

---

## 需要帮助？

查看详细文档：
- 集成指南: `docs/HARBOR_INTEGRATION_GUIDE.md`
- 操作清单: `docs/HARBOR_INTEGRATION_CHECKLIST.md`
