#!/bin/bash
# SQL文件字符集检查脚本
# 用途：确保所有SQL文件都设置了UTF-8字符集

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "  SQL文件字符集检查"
echo "========================================="
echo ""

# 查找所有SQL文件
SQL_FILES=$(find . -name "*.sql" -type f)

MISSING_CHARSET=()
TOTAL_FILES=0

for file in $SQL_FILES; do
    TOTAL_FILES=$((TOTAL_FILES + 1))
    
    # 检查前10行是否包含字符集设置
    if ! head -10 "$file" | grep -q "SET NAMES utf8mb4\|SET CHARACTER SET utf8mb4"; then
        MISSING_CHARSET+=("$file")
        echo -e "${RED}✗${NC} $file - 缺少字符集设置"
    else
        echo -e "${GREEN}✓${NC} $file"
    fi
done

echo ""
echo "========================================="
echo "  检查结果"
echo "========================================="
echo "总文件数: $TOTAL_FILES"
echo "缺少字符集设置: ${#MISSING_CHARSET[@]}"

if [ ${#MISSING_CHARSET[@]} -eq 0 ]; then
    echo -e "${GREEN}所有SQL文件都已正确设置UTF-8字符集！${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}以下文件需要添加字符集设置：${NC}"
    for file in "${MISSING_CHARSET[@]}"; do
        echo "  - $file"
    done
    echo ""
    echo -e "${YELLOW}请在文件开头添加以下语句：${NC}"
    echo "  SET NAMES utf8mb4;"
    echo "  SET CHARACTER SET utf8mb4;"
    exit 1
fi
