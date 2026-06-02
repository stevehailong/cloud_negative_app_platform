#!/bin/bash

echo "修复应用8的部署记录不同步问题"
echo "================================"

echo -e "\n当前问题："
echo "- 数据库记录显示 app-8 在 my-app-dev namespace 运行中(4副本)"
echo "- 但K8s集群中实际不存在这个 Deployment"
echo "- 导致扩缩容操作失败"

echo -e "\n解决方案："
echo "1. 删除数据库中的无效部署记录"
echo "2. 通过发布管理重新部署应用"

read -p "是否继续清理无效记录? (y/n): " confirm
if [ "$confirm" != "y" ]; then
    echo "已取消"
    exit 0
fi

echo -e "\n执行清理..."
docker exec my-cloud-mysql mysql -uroot -proot123456 deploy_db <<EOF
-- 删除无效的部署记录
DELETE FROM app_deployments WHERE id = 9;

-- 验证
SELECT COUNT(*) as remaining_deployments 
FROM app_deployments 
WHERE app_id = 8;
EOF

echo -e "\n✅ 清理完成"
echo ""
echo "下一步操作："
echo "1. 在前端进入【发布管理】"
echo "2. 找到最新的成功发布记录（如 REL-8-1780325813）"
echo "3. 点击【重新部署】或创建新的发布"
echo "4. 这将重新在K8s集群中创建 Deployment 资源"
echo ""
echo "或者直接调用API部署："
echo "  curl -X POST http://localhost:8080/api/v1/releases/{release_id}/execute \\"
echo "       -H \"Authorization: Bearer YOUR_TOKEN\""
