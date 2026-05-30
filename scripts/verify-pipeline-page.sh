#!/bin/bash

echo "=========================================="
echo "流水线页面部署验证"
echo "=========================================="
echo ""

# 1. 检查前端容器
echo "1. 检查前端容器状态"
FRONTEND_STATUS=$(docker ps --filter "name=my-cloud-frontend" --format "{{.Status}}")
if [ ! -z "$FRONTEND_STATUS" ]; then
    echo "   ✅ 前端容器运行中: $FRONTEND_STATUS"
else
    echo "   ❌ 前端容器未运行"
    exit 1
fi
echo ""

# 2. 检查PipelineList文件
echo "2. 检查PipelineList组件文件"
PIPELINE_FILES=$(docker exec my-cloud-frontend ls /usr/share/nginx/html/assets/ | grep -i pipeline | wc -l)
if [ "$PIPELINE_FILES" -gt 0 ]; then
    echo "   ✅ 找到 $PIPELINE_FILES 个PipelineList文件"
    docker exec my-cloud-frontend ls -lh /usr/share/nginx/html/assets/ | grep -i pipeline
else
    echo "   ❌ PipelineList文件不存在"
fi
echo ""

# 3. 检查文件大小
echo "3. 检查PipelineList.js文件大小"
PIPELINE_JS_SIZE=$(docker exec my-cloud-frontend ls -lh /usr/share/nginx/html/assets/ | grep "PipelineList.*\.js" | awk '{print $5}')
echo "   文件大小: $PIPELINE_JS_SIZE"
if [ "$PIPELINE_JS_SIZE" != "379" ]; then
    echo "   ✅ 文件已更新（不是旧的占位文件）"
else
    echo "   ⚠️  可能是旧的占位文件"
fi
echo ""

# 4. 检查前端页面
echo "4. 检查前端页面可访问性"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost/)
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✅ 前端页面可访问 (HTTP $HTTP_CODE)"
else
    echo "   ❌ 前端页面异常 (HTTP $HTTP_CODE)"
fi
echo ""

# 5. 检查pipelines路由
echo "5. 检查/pipelines路由"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost/pipelines)
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✅ /pipelines路由正常 (HTTP $HTTP_CODE)"
else
    echo "   ❌ /pipelines路由异常 (HTTP $HTTP_CODE)"
fi
echo ""

# 6. 检查Gateway
echo "6. 检查Gateway健康状态"
GATEWAY_STATUS=$(curl -s http://localhost:8080/health | jq -r '.data.status' 2>/dev/null)
if [ "$GATEWAY_STATUS" = "ok" ]; then
    echo "   ✅ Gateway正常"
else
    echo "   ❌ Gateway异常"
fi
echo ""

# 7. 检查Pipeline Service
echo "7. 检查Pipeline Service"
PIPELINE_SVC=$(docker ps --filter "name=my-cloud-pipeline-service" --format "{{.Status}}")
if [ ! -z "$PIPELINE_SVC" ]; then
    echo "   ✅ Pipeline Service运行中: $PIPELINE_SVC"
else
    echo "   ❌ Pipeline Service未运行"
fi
echo ""

# 8. 检查数据库数据
echo "8. 检查数据库中的流水线数据"
PIPELINE_COUNT=$(docker exec my-cloud-mysql mysql -uroot -proot123456 -e "USE devops_db; SELECT COUNT(*) as count FROM pipelines;" 2>/dev/null | grep -v count | grep -v Warning)
if [ "$PIPELINE_COUNT" -gt 0 ]; then
    echo "   ✅ 数据库中有 $PIPELINE_COUNT 条流水线数据"
else
    echo "   ❌ 数据库中没有流水线数据"
fi
echo ""

echo "=========================================="
echo "验证完成"
echo "=========================================="
echo ""
echo "📝 访问地址:"
echo "   主页: http://localhost"
echo "   流水线: http://localhost/pipelines"
echo ""
echo "💡 提示:"
echo "   1. 如果看不到流水线页面内容，请确保已登录"
echo "   2. 打开浏览器开发者工具查看Network请求"
echo "   3. 检查是否有JavaScript错误"
echo ""
