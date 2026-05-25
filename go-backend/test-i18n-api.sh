#!/bin/bash

# 博客多语言功能 API 测试脚本
# 使用方法: ./test-i18n-api.sh [base_url]
# 示例: ./test-i18n-api.sh http://localhost:9000

BASE_URL="${1:-http://localhost:9000}"

echo "=========================================="
echo "博客多语言功能 API 测试"
echo "Base URL: $BASE_URL"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    
    echo -e "${YELLOW}测试: $name${NC}"
    echo "请求: $method $endpoint"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        echo "$body"
    fi
    
    echo ""
    echo "------------------------------------------"
    echo ""
}

# 1. 测试健康检查
test_api "健康检查" "GET" "/health"

# 2. 测试获取语言列表
test_api "获取支持的语言列表" "GET" "/api/v1/i18n/languages"

# 3. 测试语言检测
test_api "检测用户语言偏好" "GET" "/api/v1/i18n/detect"

# 4. 测试设置语言
test_api "设置用户语言偏好为中文" "POST" "/api/v1/i18n/set-language" '{"locale":"zh"}'

# 5. 测试获取文章翻译（假设文章ID为1）
test_api "获取文章翻译版本 (ID=1)" "GET" "/api/v1/i18n/translations/1"

# 6. 测试 Sitemap 索引
echo -e "${YELLOW}测试: Sitemap 索引${NC}"
echo "请求: GET /sitemap.xml"
curl -s "$BASE_URL/sitemap.xml" > /tmp/sitemap-index.xml
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 成功${NC}"
    echo "文件已保存到: /tmp/sitemap-index.xml"
    head -n 20 /tmp/sitemap-index.xml
else
    echo -e "${RED}✗ 失败${NC}"
fi
echo ""
echo "------------------------------------------"
echo ""

# 7. 测试 Hreflang Sitemap
echo -e "${YELLOW}测试: Hreflang Sitemap${NC}"
echo "请求: GET /sitemap-hreflang.xml"
curl -s "$BASE_URL/sitemap-hreflang.xml" > /tmp/sitemap-hreflang.xml
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 成功${NC}"
    echo "文件已保存到: /tmp/sitemap-hreflang.xml"
    head -n 30 /tmp/sitemap-hreflang.xml
else
    echo -e "${RED}✗ 失败${NC}"
fi
echo ""
echo "------------------------------------------"
echo ""

# 8. 测试单语言 Sitemap
echo -e "${YELLOW}测试: 英文 Sitemap${NC}"
echo "请求: GET /sitemap-en.xml"
curl -s "$BASE_URL/sitemap-en.xml" > /tmp/sitemap-en.xml
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 成功${NC}"
    echo "文件已保存到: /tmp/sitemap-en.xml"
    head -n 20 /tmp/sitemap-en.xml
else
    echo -e "${RED}✗ 失败${NC}"
fi
echo ""
echo "------------------------------------------"
echo ""

# 9. 测试中文 Sitemap
echo -e "${YELLOW}测试: 中文 Sitemap${NC}"
echo "请求: GET /sitemap-zh.xml"
curl -s "$BASE_URL/sitemap-zh.xml" > /tmp/sitemap-zh.xml
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 成功${NC}"
    echo "文件已保存到: /tmp/sitemap-zh.xml"
    head -n 20 /tmp/sitemap-zh.xml
else
    echo -e "${RED}✗ 失败${NC}"
fi
echo ""
echo "------------------------------------------"
echo ""

echo "=========================================="
echo "测试完成！"
echo "=========================================="
echo ""
echo "生成的文件:"
echo "  - /tmp/sitemap-index.xml"
echo "  - /tmp/sitemap-hreflang.xml"
echo "  - /tmp/sitemap-en.xml"
echo "  - /tmp/sitemap-zh.xml"
echo ""
echo "提示: 使用 'cat /tmp/sitemap-*.xml' 查看完整内容"
