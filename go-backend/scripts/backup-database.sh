#!/bin/bash
# Legacy DATABASE-ONLY export. The independent maintenance console provides
# consistent database + attachments and the local upload/recovery loop.
set -euo pipefail
umask 077
BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DB_HOST="${DB_HOST:-postgres.production}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-commerce_platform}"
DB_USERNAME="${DB_USERNAME:-commerce_platform}"
export PGPASSWORD="${DB_PASSWORD:?DB_PASSWORD is required}"
[[ "$RETENTION_DAYS" =~ ^[0-9]+$ ]] || { echo 'Invalid retention days' >&2; exit 1; }
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/commerce_platform_backup_$(date -u +%Y%m%d_%H%M%S)_$$.sql.gz"
BACKUP_PARTIAL="$BACKUP_FILE.partial"
trap 'rm -f -- "$BACKUP_PARTIAL"' EXIT
# stderr remains separate; verbose logs must never enter SQL stdout.
pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USERNAME" -d "$DB_NAME" \
  --format=plain --no-owner --no-privileges | gzip > "$BACKUP_PARTIAL"
test -s "$BACKUP_PARTIAL"
gzip -t "$BACKUP_PARTIAL"
mv -- "$BACKUP_PARTIAL" "$BACKUP_FILE"
sha256sum "$BACKUP_FILE" > "$BACKUP_FILE.sha256"
find "$BACKUP_DIR" -maxdepth 1 -type f \( -name 'commerce_platform_backup_*.sql.gz' -o -name 'commerce_platform_backup_*.sql.gz.sha256' \) -mtime "+$RETENTION_DAYS" -delete
echo "Database-only export created: $BACKUP_FILE"
echo 'Checksum and gzip verified. SQL restore, business and attachment integrity have NOT been verified.'
