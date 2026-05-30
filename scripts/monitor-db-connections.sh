#!/bin/bash
# 数据库连接监控脚本

echo "=== MySQL 连接状态监控 ==="
echo ""

# 获取当前连接数
echo "1. 当前连接统计:"
docker exec -i my-cloud-mysql mysql -uroot -proot123456 -e "
SELECT 
    COUNT(*) as total_connections,
    SUM(CASE WHEN command = 'Sleep' THEN 1 ELSE 0 END) as sleeping,
    SUM(CASE WHEN command != 'Sleep' THEN 1 ELSE 0 END) as active
FROM information_schema.processlist;
" 2>/dev/null | tail -n +2

echo ""
echo "2. 每个数据库的连接数:"
docker exec -i my-cloud-mysql mysql -uroot -proot123456 -e "
SELECT 
    db as database_name,
    COUNT(*) as connections
FROM information_schema.processlist
WHERE db IS NOT NULL
GROUP BY db
ORDER BY connections DESC;
" 2>/dev/null | tail -n +2

echo ""
echo "3. 每个用户的连接数:"
docker exec -i my-cloud-mysql mysql -uroot -proot123456 -e "
SELECT 
    user,
    host,
    COUNT(*) as connections
FROM information_schema.processlist
GROUP BY user, host
ORDER BY connections DESC;
" 2>/dev/null | tail -n +2

echo ""
echo "4. 连接数配置:"
docker exec -i my-cloud-mysql mysql -uroot -proot123456 -e "
SHOW VARIABLES WHERE Variable_name IN ('max_connections', 'max_connect_errors', 'wait_timeout', 'interactive_timeout');
" 2>/dev/null | tail -n +2

echo ""
echo "5. 连接统计信息:"
docker exec -i my-cloud-mysql mysql -uroot -proot123456 -e "
SHOW STATUS WHERE Variable_name IN ('Threads_connected', 'Threads_running', 'Max_used_connections', 'Aborted_connects', 'Aborted_clients');
" 2>/dev/null | tail -n +2

echo ""
echo "=== 监控完成 ==="
