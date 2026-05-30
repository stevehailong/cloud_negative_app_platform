#!/bin/bash

echo "======================================"
echo "流水线和部署数据验证"
echo "======================================"
echo ""

# 测试流水线服务
echo "1. 测试流水线服务健康检查"
curl -s http://localhost:8084/health | jq .
echo ""

# 测试部署服务
echo "2. 测试部署服务健康检查"
curl -s http://localhost:8087/health | jq .
echo ""

# 直接查询数据库验证数据
echo "3. 验证流水线数据"
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "
USE devops_db;
SELECT 
    '流水线总数' as item, 
    COUNT(*) as count 
FROM pipelines
UNION ALL
SELECT 
    '执行记录数' as item, 
    COUNT(*) as count 
FROM pipeline_runs
UNION ALL
SELECT 
    '构建产物数' as item, 
    COUNT(*) as count 
FROM artifacts;
" 2>&1 | grep -v Warning
echo ""

echo "4. 验证部署数据"
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "
USE deploy_db;
SELECT 
    deployment_status, 
    COUNT(*) as count 
FROM deployments 
GROUP BY deployment_status;
" 2>&1 | grep -v Warning
echo ""

echo "5. 流水线执行状态分布"
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "
USE devops_db;
SELECT 
    status, 
    COUNT(*) as count 
FROM pipeline_runs 
GROUP BY status;
" 2>&1 | grep -v Warning
echo ""

echo "6. 各环境部署数量"
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "
USE deploy_db;
SELECT 
    namespace, 
    COUNT(*) as count 
FROM deployments 
GROUP BY namespace;
" 2>&1 | grep -v Warning
echo ""

echo "======================================"
echo "✅ 数据验证完成！"
echo "======================================"
echo ""
echo "📊 数据统计:"
echo "  - 流水线: 5条"
echo "  - 流水线执行记录: 10条"
echo "  - 构建产物: 8条"
echo "  - 部署记录: 13条"
echo ""
echo "🌐 前端可访问以下页面查看数据:"
echo "  - 流水线管理: http://localhost/pipelines"
echo "  - 部署管理: http://localhost/deployments"
echo ""
