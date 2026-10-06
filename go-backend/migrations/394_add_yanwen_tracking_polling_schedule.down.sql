DROP INDEX IF EXISTS idx_yanwen_tracking_snapshots_has_official_result;
DROP INDEX IF EXISTS idx_yanwen_tracking_snapshots_polling_due;

ALTER TABLE shipping_yanwen_tracking_snapshots
    DROP COLUMN IF EXISTS last_polling_error,
    DROP COLUMN IF EXISTS polling_failure_count,
    DROP COLUMN IF EXISTS polling_lease_until,
    DROP COLUMN IF EXISTS polling_lease_token,
    DROP COLUMN IF EXISTS last_polling_attempt_at,
    DROP COLUMN IF EXISTS next_sync_at,
    DROP COLUMN IF EXISTS polling_state,
    DROP COLUMN IF EXISTS has_official_result;
