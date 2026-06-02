#!/bin/bash

echo "=== Testing Pipeline Runs API with ImageURL ==="
echo ""

echo "1. Fetching pipeline runs for pipeline ID 1..."
response=$(docker-compose exec -T pipeline-service curl -s http://localhost:8084/api/v1/pipelines/1/runs?page=1&pageSize=3)

echo "Response:"
echo "$response" | jq '.'

echo ""
echo "2. Checking if imageUrl field exists..."
has_image=$(echo "$response" | jq '.data.list[0].imageUrl')

if [ "$has_image" != "null" ] && [ -n "$has_image" ]; then
    echo "✅ SUCCESS: imageUrl field is present"
    echo "   Image URL: $has_image"
else
    echo "❌ FAILED: imageUrl field is missing or null"
fi

echo ""
echo "3. Direct database check..."
docker-compose exec mysql mysql -uroot -proot123456 devops_db -e "
SELECT 
    pr.id, 
    pr.run_no, 
    pr.status, 
    a.artifact_type, 
    a.repo_url as image_url 
FROM pipeline_runs pr 
LEFT JOIN artifacts a ON pr.id = a.pipeline_run_id 
WHERE a.artifact_type = 'image' 
ORDER BY pr.id DESC 
LIMIT 3;
" 2>&1 | grep -v "Warning"
