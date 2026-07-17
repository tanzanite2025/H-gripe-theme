#!/bin/bash

#############################################
# PostgreSQL Database Backup Script
# 用于生产环境的数据库备份
#############################################

set -euo pipefail

# 配置
BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="tanzanite_backup_${TIMESTAMP}.sql.gz"

# 数据库连接信息（从环境变量或Kubernetes Secret获取）
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

# 创建备份目录
mkdir -p "${BACKUP_DIR}"

log "Starting database backup..."
log "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
log "Backup file: ${BACKUP_FILE}"

# 执行备份
export PGPASSWORD="${DB_PASSWORD}"

if pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USERNAME}" \
    -d "${DB_NAME}" \
    --format=plain \
    --no-owner \
    --no-privileges \
    --verbose \
    2>&1 | gzip > "${BACKUP_DIR}/${BACKUP_FILE}"; then
    
    log "Backup completed successfully"
    
    # 计算文件大小
    BACKUP_SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_FILE}" | cut -f1)
    log "Backup size: ${BACKUP_SIZE}"
    
    # 验证备份文件
    if [ -s "${BACKUP_DIR}/${BACKUP_FILE}" ]; then
        log "Backup file verification: OK"
    else
        error "Backup file is empty!"
        exit 1
    fi
else
    error "Backup failed!"
    exit 1
fi

# 清理旧备份
log "Cleaning up old backups (retention: ${RETENTION_DAYS} days)..."
find "${BACKUP_DIR}" -name "tanzanite_backup_*.sql.gz" -type f -mtime +${RETENTION_DAYS} -delete
log "Old backups cleaned up"

# 列出当前备份
log "Current backups:"
ls -lh "${BACKUP_DIR}"/tanzanite_backup_*.sql.gz 2>/dev/null || log "No backups found"

# 可选：上传到云存储（S3）
if [ "${UPLOAD_TO_S3:-false}" = "true" ]; then
    log "Uploading backup to S3..."
    if command -v aws &> /dev/null; then
        aws s3 cp "${BACKUP_DIR}/${BACKUP_FILE}" "s3://${S3_BUCKET:-tanzanite-backups}/postgres/" \
            --storage-class STANDARD_IA
        log "Backup uploaded to S3"
    else
        error "AWS CLI not found, skipping S3 upload"
    fi
fi

log "Backup process completed successfully"

# 发送通知（可选）
if [ "${NOTIFY_SLACK:-false}" = "true" ] && [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
    curl -X POST "${SLACK_WEBHOOK_URL}" \
        -H 'Content-Type: application/json' \
        -d "{
            \"text\": \"✅ Database backup completed successfully\",
            \"attachments\": [{
                \"fields\": [
                    {\"title\": \"Database\", \"value\": \"${DB_NAME}\", \"short\": true},
                    {\"title\": \"Size\", \"value\": \"${BACKUP_SIZE}\", \"short\": true},
                    {\"title\": \"Timestamp\", \"value\": \"${TIMESTAMP}\", \"short\": true}
                ]
            }]
        }" 2>/dev/null || true
fi

exit 0
