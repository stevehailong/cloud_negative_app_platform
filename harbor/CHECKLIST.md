# Harbor 部署检查清单

## ✅ 准备工作（已完成）
- [x] Harbor安装包已下载 (v2.11.0)
- [x] 解压安装包
- [x] 创建harbor.yml配置文件
- [x] 创建数据目录 `/Users/hanhailong01/Downloads/my_cloud/harbor-data`

---

## 📝 待执行（按顺序）

### 第一步：准备Harbor
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo ./prepare
```
- [ ] 执行完成
- [ ] 生成了docker-compose.yml文件

---

### 第二步：加载镜像
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo docker load -i harbor.v2.11.0.tar.gz
```
- [ ] 执行完成
- [ ] 看到多个镜像加载完成的消息
- [ ] 预计耗时：5-10分钟

---

### 第三步：启动Harbor
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo docker-compose up -d
```
- [ ] 执行完成
- [ ] 所有容器启动成功

验证：
```bash
docker-compose ps
```
- [ ] 所有容器状态为 Up 或 Up (healthy)

---

### 第四步：访问Harbor
- [ ] 打开浏览器访问 http://localhost:8093
- [ ] 使用 admin / Harbor12345 登录成功
- [ ] 能看到Harbor管理界面

---

### 第五步：基础测试

**测试1：创建项目**
- [ ] 在Harbor UI中创建项目 `mycloud`
- [ ] 设置为私有项目

**测试2：Docker登录**
```bash
docker login localhost:8093 -u admin -p Harbor12345
```
- [ ] 登录成功

**测试3：推送镜像**
```bash
docker pull nginx:alpine
docker tag nginx:alpine localhost:8093/mycloud/nginx:test
docker push localhost:8093/mycloud/nginx:test
```
- [ ] 推送成功
- [ ] 在Harbor UI中能看到nginx镜像

**测试4：拉取镜像**
```bash
docker rmi localhost:8093/mycloud/nginx:test
docker pull localhost:8093/mycloud/nginx:test
```
- [ ] 拉取成功

---

## 🎯 完成标志

当以上所有项都打勾后，Harbor已成功部署并可以使用！

---

## ⚠️ 重要注意事项

1. **保留旧Registry**
   - 不要修改主docker-compose.yml中的registry配置
   - 让两个registry并行运行
   - localhost:5001 (旧) 和 localhost:8093 (Harbor) 同时工作

2. **逐步迁移**
   - 不要立即将所有应用切换到Harbor
   - 先迁移1-2个测试应用
   - 验证稳定后再迁移其他应用

3. **数据备份**
   - Harbor数据在 `/Users/hanhailong01/Downloads/my_cloud/harbor-data`
   - 定期备份此目录

---

## 📞 遇到问题？

查看详细指南：
```bash
cat /Users/hanhailong01/Downloads/my_cloud/docs/HARBOR_DEPLOYMENT_TEST.md
```

查看Harbor日志：
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose logs -f
```

---

## 开始执行

现在请按照清单从"第一步"开始执行！

第一个命令：
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
sudo ./prepare
```
