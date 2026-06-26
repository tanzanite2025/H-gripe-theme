#!/bin/bash

#############################################
# 本地开发环境快速启动脚本
# 用于新开发者快速搭建本地开发环境
#############################################

set -euo pipefail

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

# 显示标题
echo ""
echo "=========================================="
echo "  Tanzanite 本地开发环境设置"
echo "=========================================="
echo ""

# 检查操作系统
OS="$(uname -s)"
case "${OS}" in
    Linux*)     PLATFORM=Linux;;
    Darwin*)    PLATFORM=Mac;;
    MINGW*)     PLATFORM=Windows;;
    *)          PLATFORM="UNKNOWN:${OS}"
esac
log "检测到操作系统: ${PLATFORM}"
echo ""

#############################################
# Step 1: 检查必需工具
#############################################
log "Step 1: 检查必需工具..."

check_tool() {
    local tool=$1
    local install_cmd=$2
    
    if command -v "${tool}" &> /dev/null; then
        log "✓ ${tool} 已安装"
        return 0
    else
        warn "✗ ${tool} 未安装"
        if [ -n "${install_cmd}" ]; then
            info "  安装命令: ${install_cmd}"
        fi
        return 1
    fi
}

MISSING_TOOLS=0

# Go
if ! check_tool "go" "https://go.dev/doc/install"; then
    MISSING_TOOLS=$((MISSING_TOOLS + 1))
fi

# Docker
if ! check_tool "docker" "https://docs.docker.com/get-docker/"; then
    MISSING_TOOLS=$((MISSING_TOOLS + 1))
fi

# Docker Compose
if ! check_tool "docker-compose" "https://docs.docker.com/compose/install/"; then
    MISSING_TOOLS=$((MISSING_TOOLS + 1))
fi

# Make
if ! check_tool "make" ""; then
    warn "  Make未安装，某些命令可能无法使用"
fi

# Git
if ! check_tool "git" ""; then
    MISSING_TOOLS=$((MISSING_TOOLS + 1))
fi

if [ ${MISSING_TOOLS} -gt 0 ]; then
    error "缺少 ${MISSING_TOOLS} 个必需工具，请先安装后再运行此脚本"
    exit 1
fi

echo ""

#############################################
# Step 2: 创建环境配置文件
#############################################
log "Step 2: 创建环境配置文件..."

if [ ! -f ".env" ]; then
    log "创建 .env 文件..."
    cat > .env <<EOF
# 应用配置
APP_ENV=development
APP_PORT=8080
APP_DEBUG=true

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=tanzanite_dev
DB_USER=tanzanite
DB_PASSWORD=tanzanite_dev_password
DB_SSLMODE=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=$(openssl rand -base64 32)
JWT_EXPIRY=24h

# 文件存储配置
STORAGE_TYPE=local
STORAGE_PATH=./storage

# 支付网关配置（开发环境使用测试密钥）
STRIPE_SECRET_KEY=sk_test_your_key_here
STRIPE_PUBLISHABLE_KEY=pk_test_your_key_here

PAYPAL_CLIENT_ID=your_client_id_here
PAYPAL_SECRET=your_secret_here
PAYPAL_MODE=sandbox

ALIPAY_APP_ID=your_app_id_here
ALIPAY_PRIVATE_KEY=your_private_key_here
ALIPAY_PUBLIC_KEY=your_public_key_here

WECHAT_APP_ID=your_app_id_here
WECHAT_MCH_ID=your_mch_id_here
WECHAT_API_KEY=your_api_key_here

# CORS配置
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# 日志配置
LOG_LEVEL=debug
LOG_FORMAT=json
EOF
    log "✓ .env 文件已创建"
else
    log "✓ .env 文件已存在"
fi

echo ""

#############################################
# Step 3: 启动Docker服务
#############################################
log "Step 3: 启动Docker服务（PostgreSQL和Redis）..."

# 检查docker-compose.yml是否存在
if [ ! -f "../docker-compose.yml" ]; then
    error "docker-compose.yml 文件不存在"
    exit 1
fi

# 启动服务
log "启动PostgreSQL和Redis..."
cd ..
docker-compose up -d postgres redis

# 等待服务就绪
log "等待数据库服务就绪..."
sleep 5

# 检查服务状态
if docker-compose ps | grep -q "postgres.*Up"; then
    log "✓ PostgreSQL 已启动"
else
    error "PostgreSQL 启动失败"
    docker-compose logs postgres
    exit 1
fi

if docker-compose ps | grep -q "redis.*Up"; then
    log "✓ Redis 已启动"
else
    error "Redis 启动失败"
    docker-compose logs redis
    exit 1
fi

cd go-backend

echo ""

#############################################
# Step 4: 安装Go依赖
#############################################
log "Step 4: 安装Go依赖..."

log "运行 go mod download..."
if go mod download; then
    log "✓ Go依赖安装成功"
else
    error "Go依赖安装失败"
    exit 1
fi

# 安装开发工具
log "安装开发工具..."

install_dev_tool() {
    local tool=$1
    local package=$2
    
    if ! command -v "${tool}" &> /dev/null; then
        log "安装 ${tool}..."
        go install "${package}" || warn "安装 ${tool} 失败"
    else
        log "✓ ${tool} 已安装"
    fi
}

install_dev_tool "swag" "github.com/swaggo/swag/cmd/swag@latest"
install_dev_tool "golangci-lint" "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
install_dev_tool "air" "github.com/cosmtrek/air@latest"

echo ""

#############################################
# Step 5: 数据库迁移
#############################################
log "Step 5: 运行数据库迁移..."

# 等待数据库完全就绪
sleep 3

log "创建数据库表..."
if make migrate-up 2>/dev/null || go run cmd/server/main.go migrate; then
    log "✓ 数据库迁移完成"
else
    warn "数据库迁移失败，请手动运行"
fi

echo ""

#############################################
# Step 6: 生成Swagger文档
#############################################
log "Step 6: 生成API文档..."

if command -v swag &> /dev/null; then
    log "生成Swagger文档..."
    if swag init -g cmd/server/main.go; then
        log "✓ Swagger文档已生成"
    else
        warn "Swagger文档生成失败"
    fi
else
    warn "swag未安装，跳过API文档生成"
fi

echo ""

#############################################
# Step 7: 运行测试
#############################################
log "Step 7: 运行测试..."

log "运行单元测试..."
if go test ./... -short; then
    log "✓ 测试通过"
else
    warn "部分测试失败，请检查"
fi

echo ""

#############################################
# 完成
#############################################
echo ""
echo "=========================================="
log "✓ 开发环境设置完成！"
echo "=========================================="
echo ""
info "下一步操作："
echo ""
echo "1. 启动开发服务器（热重载）："
echo "   ${GREEN}make dev${NC} 或 ${GREEN}air${NC}"
echo ""
echo "2. 启动生产模式："
echo "   ${GREEN}make run${NC}"
echo ""
echo "3. 查看API文档："
echo "   ${GREEN}http://localhost:8080/swagger/index.html${NC}"
echo ""
echo "4. 查看健康状态："
echo "   ${GREEN}curl http://localhost:8080/health${NC}"
echo ""
echo "5. 运行测试："
echo "   ${GREEN}make test${NC}"
echo ""
echo "6. 查看可用命令："
echo "   ${GREEN}make help${NC}"
echo ""
info "数据库连接信息："
echo "  Host: localhost:5432"
echo "  Database: tanzanite_dev"
echo "  User: tanzanite"
echo "  Password: tanzanite_dev_password"
echo ""
info "Redis连接信息："
echo "  Host: localhost:6379"
echo "  Database: 0"
echo ""
echo "=========================================="
echo ""

exit 0
