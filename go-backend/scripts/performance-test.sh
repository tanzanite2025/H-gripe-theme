#!/bin/bash

#############################################
# API Performance Testing Script
# 使用Apache Bench (ab) 和 wrk 进行性能测试
#############################################

set -euo pipefail

# 配置
API_URL="${API_URL:-http://localhost:8080}"
CONCURRENCY="${CONCURRENCY:-50}"
REQUESTS="${REQUESTS:-10000}"
DURATION="${DURATION:-30s}"
OUTPUT_DIR="${OUTPUT_DIR:-./test-results}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $*" >&2
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $*"
}

# 检查工具是否安装
check_tools() {
    local missing_tools=()
    
    if ! command -v ab &> /dev/null; then
        missing_tools+=("apache-bench")
    fi
    
    if ! command -v wrk &> /dev/null; then
        missing_tools+=("wrk")
    fi
    
    if ! command -v curl &> /dev/null; then
        missing_tools+=("curl")
    fi
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        error "Missing tools: ${missing_tools[*]}"
        echo ""
        echo "Install instructions:"
        echo "  Ubuntu/Debian: sudo apt-get install apache2-utils wrk curl"
        echo "  macOS: brew install wrk curl"
        echo "  Windows: Use WSL or install from respective websites"
        exit 1
    fi
}

# 创建输出目录
mkdir -p "${OUTPUT_DIR}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

log "==================================="
log "API Performance Testing"
log "==================================="
log "API URL: ${API_URL}"
log "Concurrency: ${CONCURRENCY}"
log "Requests: ${REQUESTS}"
log "Duration: ${DURATION}"
log "Output: ${OUTPUT_DIR}"
log "==================================="
echo ""

# 检查API是否可访问
log "Checking API availability..."
if curl -sf "${API_URL}/health" > /dev/null; then
    log "✓ API is accessible"
else
    error "✗ API is not accessible at ${API_URL}"
    exit 1
fi
echo ""

#############################################
# Test 1: Health Endpoint
#############################################
log "Test 1: Health Endpoint (/health)"
log "Running Apache Bench..."

ab -n "${REQUESTS}" -c "${CONCURRENCY}" \
   -g "${OUTPUT_DIR}/health_${TIMESTAMP}.tsv" \
   "${API_URL}/health" \
   > "${OUTPUT_DIR}/health_${TIMESTAMP}.txt" 2>&1

# 提取关键指标
HEALTH_RPS=$(grep "Requests per second:" "${OUTPUT_DIR}/health_${TIMESTAMP}.txt" | awk '{print $4}')
HEALTH_TIME=$(grep "Time per request:" "${OUTPUT_DIR}/health_${TIMESTAMP}.txt" | head -1 | awk '{print $4}')
HEALTH_FAILED=$(grep "Failed requests:" "${OUTPUT_DIR}/health_${TIMESTAMP}.txt" | awk '{print $3}')

log "  Requests/sec: ${HEALTH_RPS}"
log "  Time/request: ${HEALTH_TIME}ms"
log "  Failed: ${HEALTH_FAILED}"
echo ""

#############################################
# Test 2: API Endpoints with wrk
#############################################
log "Test 2: Mixed Workload (wrk)"

# 创建wrk Lua脚本
cat > "${OUTPUT_DIR}/mixed_workload.lua" <<'LUA'
-- Mixed workload script
wrk.method = "GET"
wrk.headers["Content-Type"] = "application/json"

-- 定义测试端点
local endpoints = {
    "/api/v1/health",
    "/api/v1/users?page=1&limit=10",
    "/api/v1/products?page=1&limit=20",
    "/api/v1/categories",
}

local counter = 1

request = function()
    local path = endpoints[counter]
    counter = counter + 1
    if counter > #endpoints then
        counter = 1
    end
    return wrk.format(nil, path)
end

response = function(status, headers, body)
    if status ~= 200 then
        print("Error: " .. status)
    end
end
LUA

log "Running wrk for ${DURATION}..."
wrk -t4 -c"${CONCURRENCY}" -d"${DURATION}" \
    -s "${OUTPUT_DIR}/mixed_workload.lua" \
    --latency \
    "${API_URL}" \
    > "${OUTPUT_DIR}/mixed_${TIMESTAMP}.txt" 2>&1

# 提取wrk结果
MIXED_RPS=$(grep "Requests/sec:" "${OUTPUT_DIR}/mixed_${TIMESTAMP}.txt" | awk '{print $2}')
MIXED_LATENCY_AVG=$(grep "Latency" "${OUTPUT_DIR}/mixed_${TIMESTAMP}.txt" | awk '{print $2}')
MIXED_LATENCY_P99=$(grep "99%" "${OUTPUT_DIR}/mixed_${TIMESTAMP}.txt" | awk '{print $2}')

log "  Requests/sec: ${MIXED_RPS}"
log "  Avg Latency: ${MIXED_LATENCY_AVG}"
log "  P99 Latency: ${MIXED_LATENCY_P99}"
echo ""

#############################################
# Test 3: POST Requests (User Registration)
#############################################
log "Test 3: POST Requests (User Registration Simulation)"

# 创建POST数据
cat > "${OUTPUT_DIR}/post_data.json" <<'JSON'
{
    "username": "testuser",
    "email": "test@example.com",
    "password": "TestPassword123!"
}
JSON

# 注意：这个测试可能会创建实际数据，谨慎使用
warn "Skipping POST test to avoid creating test data"
warn "To enable, uncomment the ab command below"

# ab -n 100 -c 10 \
#    -p "${OUTPUT_DIR}/post_data.json" \
#    -T "application/json" \
#    "${API_URL}/api/v1/auth/register" \
#    > "${OUTPUT_DIR}/post_${TIMESTAMP}.txt" 2>&1

echo ""

#############################################
# Test 4: Database-Heavy Endpoints
#############################################
log "Test 4: Database Query Performance"

log "Testing product search..."
ab -n 1000 -c 20 \
   "${API_URL}/api/v1/products/search?q=test" \
   > "${OUTPUT_DIR}/db_search_${TIMESTAMP}.txt" 2>&1

DB_SEARCH_RPS=$(grep "Requests per second:" "${OUTPUT_DIR}/db_search_${TIMESTAMP}.txt" | awk '{print $4}')
DB_SEARCH_TIME=$(grep "Time per request:" "${OUTPUT_DIR}/db_search_${TIMESTAMP}.txt" | head -1 | awk '{print $4}')

log "  Search Requests/sec: ${DB_SEARCH_RPS}"
log "  Search Time/request: ${DB_SEARCH_TIME}ms"
echo ""

#############################################
# Test 5: Static Content
#############################################
log "Test 5: Static Content Performance"

if curl -sf "${API_URL}/static/logo.png" > /dev/null 2>&1; then
    ab -n 10000 -c 100 \
       "${API_URL}/static/logo.png" \
       > "${OUTPUT_DIR}/static_${TIMESTAMP}.txt" 2>&1
    
    STATIC_RPS=$(grep "Requests per second:" "${OUTPUT_DIR}/static_${TIMESTAMP}.txt" | awk '{print $4}')
    log "  Static Requests/sec: ${STATIC_RPS}"
else
    warn "Static content endpoint not available, skipping test"
fi
echo ""

#############################################
# Generate Summary Report
#############################################
log "Generating summary report..."

cat > "${OUTPUT_DIR}/summary_${TIMESTAMP}.md" <<EOF
# Performance Test Report
**Date:** $(date)
**API URL:** ${API_URL}
**Test Configuration:**
- Concurrency: ${CONCURRENCY}
- Total Requests: ${REQUESTS}
- Duration: ${DURATION}

## Results Summary

### 1. Health Endpoint
- **Requests/sec:** ${HEALTH_RPS}
- **Time/request:** ${HEALTH_TIME}ms
- **Failed requests:** ${HEALTH_FAILED}

### 2. Mixed Workload (wrk)
- **Requests/sec:** ${MIXED_RPS}
- **Average Latency:** ${MIXED_LATENCY_AVG}
- **P99 Latency:** ${MIXED_LATENCY_P99}

### 3. Database Queries
- **Search Requests/sec:** ${DB_SEARCH_RPS}
- **Search Time/request:** ${DB_SEARCH_TIME}ms

### 4. Static Content
- **Requests/sec:** ${STATIC_RPS:-N/A}

## Performance Benchmarks

| Endpoint Type | Target RPS | Actual RPS | Status |
|--------------|-----------|-----------|--------|
| Health Check | > 5000 | ${HEALTH_RPS} | $([ $(echo "${HEALTH_RPS} > 5000" | bc -l) -eq 1 ] && echo "✅ PASS" || echo "❌ FAIL") |
| API Endpoints | > 1000 | ${MIXED_RPS} | $([ $(echo "${MIXED_RPS} > 1000" | bc -l) -eq 1 ] && echo "✅ PASS" || echo "❌ FAIL") |
| DB Queries | > 500 | ${DB_SEARCH_RPS} | $([ $(echo "${DB_SEARCH_RPS} > 500" | bc -l) -eq 1 ] && echo "✅ PASS" || echo "❌ FAIL") |

## Recommendations

$(if [ $(echo "${HEALTH_RPS} < 5000" | bc -l) -eq 1 ]; then
    echo "- ⚠️ Health endpoint performance is below target. Consider optimizing health check logic."
fi)

$(if [ $(echo "${MIXED_RPS} < 1000" | bc -l) -eq 1 ]; then
    echo "- ⚠️ API endpoint performance is below target. Review database queries and caching strategy."
fi)

$(if [ $(echo "${DB_SEARCH_RPS} < 500" | bc -l) -eq 1 ]; then
    echo "- ⚠️ Database query performance is below target. Add database indexes and optimize queries."
fi)

## Detailed Results

Full test results are available in:
- Health Endpoint: \`health_${TIMESTAMP}.txt\`
- Mixed Workload: \`mixed_${TIMESTAMP}.txt\`
- Database Queries: \`db_search_${TIMESTAMP}.txt\`
- Static Content: \`static_${TIMESTAMP}.txt\`

EOF

log "Summary report saved to: ${OUTPUT_DIR}/summary_${TIMESTAMP}.md"
cat "${OUTPUT_DIR}/summary_${TIMESTAMP}.md"

log ""
log "==================================="
log "Performance testing completed!"
log "Results saved to: ${OUTPUT_DIR}"
log "==================================="

exit 0
