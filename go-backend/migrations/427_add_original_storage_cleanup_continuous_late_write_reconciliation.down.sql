-- Rollback removes only reconciliation metadata. Exact original authorizations,
-- writer fences and their next reconciliation time must survive rollback.
DROP INDEX IF EXISTS idx_outbox_original_storage_cleanup_reconciliation_due;
ALTER TABLE outbox_events
    DROP COLUMN IF EXISTS storage_cleanup_last_reconciled_at,
    DROP COLUMN IF EXISTS storage_cleanup_reconciliation_error;
