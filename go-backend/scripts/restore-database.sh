#!/bin/bash
# Legacy SQL validation into a NEW isolated database only. Whole-storefront
# overwrite uses the independent paired database/media recovery executor.
set -euo pipefail
umask 077
if [ "$#" -ne 1 ] || [ ! -f "$1" ]; then
  echo 'Usage: RESTORE_VALIDATION_DATABASE=local_backup_legacy_validation_<unique> restore-database.sh <backup.sql[.gz]>' >&2
  exit 1
fi
BACKUP_FILE="$1"
RESTORE_VALIDATION_DATABASE="${RESTORE_VALIDATION_DATABASE:?explicit NEW isolated database name required}"
[[ "$RESTORE_VALIDATION_DATABASE" =~ ^local_backup_legacy_validation_[a-z0-9_]{1,24}$ ]] || { echo 'Unsafe isolated database name' >&2; exit 1; }
DB_HOST="${DB_HOST:-postgres.production}"
DB_PORT="${DB_PORT:-5432}"
DB_USERNAME="${DB_USERNAME:-commerce_platform}"
export PGPASSWORD="${DB_PASSWORD:?DB_PASSWORD is required}"
POSTGRES_ARGUMENTS=(-h "$DB_HOST" -p "$DB_PORT" -U "$DB_USERNAME")
if [[ "$BACKUP_FILE" == *.gz ]]; then gzip -t "$BACKUP_FILE"; fi
# createdb fails if the name exists; no live database is dropped or reused.
createdb "${POSTGRES_ARGUMENTS[@]}" --template=template0 "$RESTORE_VALIDATION_DATABASE"
if [[ "$BACKUP_FILE" == *.gz ]]; then
  gunzip -c "$BACKUP_FILE" | psql "${POSTGRES_ARGUMENTS[@]}" -X --set=ON_ERROR_STOP=1 --single-transaction -d "$RESTORE_VALIDATION_DATABASE"
else
  psql "${POSTGRES_ARGUMENTS[@]}" -X --set=ON_ERROR_STOP=1 --single-transaction -d "$RESTORE_VALIDATION_DATABASE" < "$BACKUP_FILE"
fi
psql "${POSTGRES_ARGUMENTS[@]}" -X --set=ON_ERROR_STOP=1 -d "$RESTORE_VALIDATION_DATABASE" -c 'ANALYZE;'
echo "SQL import completed in isolated database: $RESTORE_VALIDATION_DATABASE"
echo 'Business references, attachments, keys and complete application recovery still need independent validation.'
