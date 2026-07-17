#!/bin/bash

#############################################
# PostgreSQL Database Restore Script
# 用于从备份恢复数据库
#############################################

set -euo pipefail

# 配置
BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"

# 数据库连接信息
DB_HOST="${DB_HOST:-postgres.production}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tanzanite}"
DB_USERNAME="${DB_USERNAME:-tanzanite}"
DB_PASSWORD="${DB_PASSWORD}"

# 日志函数
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2
}

# 使用帮助
usage() {
    cat <<EOF
Usage: $0 <backup_file>

Restore PostgreSQL database from a backup file.

Options:
    <backup_file>    Path to the backup file (*.sql.gz or *.sql)
    -h, --help       Show this help message

Example:
    $0 /backups/postgres/tanzanite_backup_20240101_120000.sql.gz
    $0 latest  # Restore from the most recent backup
EOF
    exit 1
}

# 检查参数
if [ $# -eq 0 ]; then
    usage
fi

BACKUP_FILE="$1"

# 如果指定 "latest"，使用最新的备份文件
if [ "${BACKUP_FILE}" = "latest" ]; then
    BACKUP_FILE=$(ls -t "${BACKUP_DIR}"/tanzanite_backup_*.sql.gz 2>/dev/null | head -1)
    if [ -z "${BACKUP_FILE}" ]; then
        error "No backup files found in ${BACKUP_DIR}"
        exit 1
    fi
    log "Using latest backup: ${BACKUP_FILE}"
fi

# 检查备份文件是否存在
if [ ! -f "${BACKUP_FILE}" ]; then
    error "Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

log "Starting database restore..."
log "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
log "Backup file: ${BACKUP_FILE}"

# 确认操作
read -p "⚠️  WARNING: This will OVERWRITE the database '${DB_NAME}'. Continue? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    log "Restore cancelled by user"
    exit 0
fi

# 设置密码
export PGPASSWORD="${DB_PASSWORD}"

# 检查数据库连接
log "Testing database connection..."
if ! psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d postgres -c "SELECT 1" &>/dev/null; then
    error "Cannot connect to database server"
    exit 1
fi
log "Database connection: OK"

# 终止活动连接
log "Terminating active connections..."
psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d postgres <<EOF
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '${DB_NAME}'
  AND pid <> pg_backend_pid();
EOF

# 删除并重建数据库
log "Dropping database..."
psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d postgres -c "DROP DATABASE IF EXISTS ${DB_NAME};"

log "Creating database..."
psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d postgres -c "CREATE DATABASE ${DB_NAME};"

# 恢复数据
log "Restoring data..."
if [[ "${BACKUP_FILE}" == *.gz ]]; then
    # 压缩文件
    if gunzip -c "${BACKUP_FILE}" | psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d "${DB_NAME}" 2>&1; then
        log "Restore completed successfully"
    else
        error "Restore failed!"
        exit 1
    fi
else
    # 未压缩文件
    if psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d "${DB_NAME}" < "${BACKUP_FILE}" 2>&1; then
        log "Restore completed successfully"
    else
        error "Restore failed!"
        exit 1
    fi
fi

# 验证恢复
log "Verifying restore..."
TABLE_COUNT=$(psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d "${DB_NAME}" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
log "Tables restored: ${TABLE_COUNT}"

# 运行ANALYZE以更新统计信息
log "Analyzing database..."
psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" -d "${DB_NAME}" -c "ANALYZE;"

log "Database restore completed successfully"

# 发送通知（可选）
if [ "${NOTIFY_SLACK:-false}" = "true" ] && [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
    curl -X POST "${SLACK_WEBHOOK_URL}" \
        -H 'Content-Type: application/json' \
        -d "{
            \"text\": \"✅ Database restore completed successfully\",
            \"attachments\": [{
                \"fields\": [
                    {\"title\": \"Database\", \"value\": \"${DB_NAME}\", \"short\": true},
                    {\"title\": \"Tables\", \"value\": \"${TABLE_COUNT}\", \"short\": true},
                    {\"title\": \"Backup File\", \"value\": \"$(basename ${BACKUP_FILE})\", \"short\": false}
                ]
            }]
        }" 2>/dev/null || true
fi

exit 0
