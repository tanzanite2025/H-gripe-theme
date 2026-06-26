#!/bin/bash

#############################################
# 数据库迁移管理脚本
# 使用golang-migrate进行数据库版本管理
#############################################

set -euo pipefail

# 配置
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./migrations}"
DB_URL="${DATABASE_URL:-postgresql://tanzanite:tanzanite_dev_password@localhost:5432/tanzanite_dev?sslmode=disable}"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

# 检查migrate工具是否安装
check_migrate() {
    if ! command -v migrate &> /dev/null; then
        error "migrate工具未安装"
        echo ""
        echo "安装方法："
        echo "  macOS: brew install golang-migrate"
        echo "  Linux: https://github.com/golang-migrate/migrate/releases"
        echo "  或使用Go: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
}

# 使用说明
usage() {
    cat <<EOF
数据库迁移管理工具

用法: $0 <命令> [参数]

命令:
    create <name>       创建新的迁移文件
    up [n]             应用所有（或指定数量）待执行的迁移
    down [n]           回滚所有（或指定数量）已执行的迁移
    goto <version>     迁移到指定版本
    force <version>    强制设置迁移版本（用于修复脏状态）
    version            显示当前迁移版本
    status             显示迁移状态
    drop               删除所有数据（危险操作！）
    help               显示此帮助信息

示例:
    $0 create add_users_table
    $0 up
    $0 up 1
    $0 down 1
    $0 version
    $0 status

环境变量:
    DATABASE_URL       数据库连接URL（默认：本地开发数据库）
    MIGRATIONS_DIR     迁移文件目录（默认：./migrations）
EOF
    exit 1
}

# 创建新的迁移文件
create_migration() {
    local name=$1
    if [ -z "${name}" ]; then
        error "请提供迁移名称"
        echo "用法: $0 create <name>"
        exit 1
    fi
    
    log "创建迁移文件: ${name}"
    migrate create -ext sql -dir "${MIGRATIONS_DIR}" -seq "${name}"
    log "迁移文件已创建在: ${MIGRATIONS_DIR}"
}

# 应用迁移
migrate_up() {
    local steps=${1:-}
    
    if [ -z "${steps}" ]; then
        log "应用所有待执行的迁移..."
        migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" up
    else
        log "应用 ${steps} 个迁移..."
        migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" up "${steps}"
    fi
    
    log "迁移完成"
    show_version
}

# 回滚迁移
migrate_down() {
    local steps=${1:-}
    
    warn "⚠️  警告：此操作将回滚数据库更改"
    read -p "确认继续？(yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log "操作已取消"
        exit 0
    fi
    
    if [ -z "${steps}" ]; then
        log "回滚所有迁移..."
        migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" down
    else
        log "回滚 ${steps} 个迁移..."
        migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" down "${steps}"
    fi
    
    log "回滚完成"
    show_version
}

# 迁移到指定版本
migrate_goto() {
    local version=$1
    if [ -z "${version}" ]; then
        error "请提供目标版本号"
        exit 1
    fi
    
    log "迁移到版本: ${version}"
    migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" goto "${version}"
    log "迁移完成"
}

# 强制设置版本（修复脏状态）
migrate_force() {
    local version=$1
    if [ -z "${version}" ]; then
        error "请提供版本号"
        exit 1
    fi
    
    warn "⚠️  强制设置迁移版本可能导致数据不一致"
    read -p "确认继续？(yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log "操作已取消"
        exit 0
    fi
    
    log "强制设置版本: ${version}"
    migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" force "${version}"
    log "版本已设置"
}

# 显示当前版本
show_version() {
    log "当前迁移版本:"
    migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" version
}

# 显示迁移状态
show_status() {
    log "迁移状态:"
    echo ""
    echo "数据库: ${DB_URL}"
    echo "迁移目录: ${MIGRATIONS_DIR}"
    echo ""
    
    # 显示当前版本
    local version
    version=$(migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" version 2>&1 | grep -oE '[0-9]+' | head -1 || echo "0")
    echo "当前版本: ${version}"
    
    # 列出所有迁移文件
    echo ""
    echo "可用的迁移文件:"
    if [ -d "${MIGRATIONS_DIR}" ]; then
        ls -la "${MIGRATIONS_DIR}"/*.sql 2>/dev/null || echo "  无迁移文件"
    else
        warn "迁移目录不存在: ${MIGRATIONS_DIR}"
    fi
}

# 删除所有数据（危险操作）
migrate_drop() {
    error "⚠️  危险操作：此操作将删除数据库中的所有数据！"
    read -p "请输入 'DELETE ALL DATA' 以确认: " -r
    if [ "${REPLY}" != "DELETE ALL DATA" ]; then
        log "操作已取消"
        exit 0
    fi
    
    log "删除所有数据..."
    migrate -path "${MIGRATIONS_DIR}" -database "${DB_URL}" drop -f
    log "数据已删除"
}

# 检查migrate工具
check_migrate

# 处理命令
case "${1:-help}" in
    create)
        create_migration "${2:-}"
        ;;
    up)
        migrate_up "${2:-}"
        ;;
    down)
        migrate_down "${2:-}"
        ;;
    goto)
        migrate_goto "${2:-}"
        ;;
    force)
        migrate_force "${2:-}"
        ;;
    version)
        show_version
        ;;
    status)
        show_status
        ;;
    drop)
        migrate_drop
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        error "未知命令: $1"
        usage
        ;;
esac

exit 0
