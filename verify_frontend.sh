#!/bin/bash

echo "================================"
echo "前端功能验证脚本"
echo "================================"
echo ""

# 检查frontend容器状态
echo "1. 检查frontend容器状态"
echo "--------------------------------------"
docker ps | grep frontend
echo ""

# 检查ApplicationDetail.vue中是否还有showBindEnv的实现
echo "2. 检查ApplicationDetail.vue代码"
echo "--------------------------------------"
if grep -q "showBindEnv" /Users/hanhailong01/Downloads/my_cloud/frontend/src/views/application/ApplicationDetail.vue; then
    echo "✓ showBindEnv函数存在"
    echo ""
    echo "函数实现预览："
    grep -A 10 "const showBindEnv" /Users/hanhailong01/Downloads/my_cloud/frontend/src/views/application/ApplicationDetail.vue | head -15
else
    echo "✗ showBindEnv函数不存在"
fi
echo ""

# 检查是否有bindEnvDialogVisible
echo "3. 检查环境绑定对话框"
echo "--------------------------------------"
if grep -q "bindEnvDialogVisible" /Users/hanhailong01/Downloads/my_cloud/frontend/src/views/application/ApplicationDetail.vue; then
    echo "✓ 环境绑定对话框已定义"
else
    echo "✗ 环境绑定对话框未定义"
fi
echo ""

# 检查后端API
echo "4. 检查后端API"
echo "--------------------------------------"
echo "测试环境列表API："
docker exec my-cloud-env-service wget -qO- 'http://localhost:8085/api/v1/environments?page=1&pageSize=1' 2>&1 | jq -r '.data.list[0].clusterName // "API调用失败"'

echo ""
echo "测试绑定列表API："
docker exec my-cloud-env-service wget -qO- 'http://localhost:8085/api/v1/app-env-bindings?page=1&pageSize=1' 2>&1 | jq -r '.code // "API调用失败"'

echo ""
echo "================================"
echo "如果上面的检查都通过，说明后端正常"
echo "请尝试以下操作清理浏览器缓存："
echo ""
echo "方法1: 开发者工具 (F12) → Network标签 → 勾选 Disable cache"
echo "方法2: 无痕模式访问 http://localhost"
echo "方法3: 完全关闭浏览器重新打开"
echo ""
echo "如果还是不行，请提供："
echo "1. 浏览器控制台(Console)的错误信息"
echo "2. Network标签中对应API的返回内容"
echo "================================"
