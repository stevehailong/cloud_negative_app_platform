#!/bin/bash

echo "创建测试数据并验证环境绑定修复"
echo "=================================="

# 1. 创建测试流水线
echo -e "\n1. 创建测试流水线..."
docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db <<EOF
INSERT INTO pipelines (app_id, pipeline_code, pipeline_name, repo_url, branch, status, create_time, update_time) 
VALUES (1, 'test-pipeline-001', '测试流水线', 'https://gitlab.example.com/test/app', 'main', 1, NOW(), NOW());
EOF

PIPELINE_ID=$(docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db -e "SELECT LAST_INSERT_ID()" -s -N)
echo "   ✅ 创建流水线 ID: $PIPELINE_ID"

# 2. 创建 pipeline run
echo -e "\n2. 创建流水线运行记录..."
docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db <<EOF
INSERT INTO pipeline_runs (pipeline_id, run_no, status, trigger_type, commit_id, commit_message, create_time, update_time) 
VALUES ($PIPELINE_ID, 'RUN-TEST-001', 'success', 'manual', 'abc123', 'test commit', NOW(), NOW());
EOF

RUN_ID=$(docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db -e "SELECT LAST_INSERT_ID()" -s -N)
echo "   ✅ 创建运行记录 ID: $RUN_ID"

# 3. 创建构建制品
echo -e "\n3. 创建构建制品..."
docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db <<EOF
INSERT INTO artifacts (pipeline_id, pipeline_run_id, artifact_type, artifact_version, repo_url, create_time) 
VALUES ($PIPELINE_ID, $RUN_ID, 'image', 'v1.0.0-test', 'registry.example.com/test-app:v1.0.0-test', NOW());
EOF

ARTIFACT_ID=$(docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db -e "SELECT LAST_INSERT_ID()" -s -N)
echo "   ✅ 创建制品 ID: $ARTIFACT_ID"

# 4. 验证数据
echo -e "\n4. 验证测试数据..."
echo "   流水线:"
docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db -e "SELECT id, app_id, pipeline_code FROM pipelines WHERE id=$PIPELINE_ID" 

echo -e "\n   制品:"
docker exec my-cloud-mysql mysql -uroot -proot123456 pipeline_db -e "SELECT id, artifact_type, artifact_version, repo_url FROM artifacts WHERE id=$ARTIFACT_ID"

echo -e "\n   应用环境绑定:"
docker exec my-cloud-mysql mysql -uroot -proot123456 env_db -e "SELECT b.id, b.app_id, b.env_id, e.env_name, e.env_type FROM app_env_bindings b JOIN environments e ON b.env_id = e.id WHERE b.app_id=1 AND b.is_deleted=0"

# 5. 检查当前 release 记录（用于对比）
echo -e "\n5. 当前 release 记录:"
RELEASE_COUNT=$(docker exec my-cloud-mysql mysql -uroot -proot123456 release_db -e "SELECT COUNT(*) FROM releases" -s -N)
echo "   总数: $RELEASE_COUNT"
if [ "$RELEASE_COUNT" -gt "0" ]; then
    echo "   最近的release:"
    docker exec my-cloud-mysql mysql -uroot -proot123456 release_db -e "SELECT id, app_id, env_id, release_version, status FROM releases ORDER BY id DESC LIMIT 3"
fi

echo -e "\n=================================="
echo "测试数据创建完成！"
echo ""
echo "流水线 ID: $PIPELINE_ID"
echo "制品 ID: $ARTIFACT_ID"
echo ""
echo "可以通过前端或API测试 DeployPipeline 功能"
echo "预期行为："
echo "  1. 查询应用ID=1的环境绑定"
echo "  2. 找到绑定的环境 ID=1 (dev-开发环境)"
echo "  3. 创建 release 时使用 envId=1"
echo "  4. Description 包含 '目标环境: dev-开发环境'"
