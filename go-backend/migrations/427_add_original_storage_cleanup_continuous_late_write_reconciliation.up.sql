-- Surviving original cleanup authorizations are the durable ledger. Do not
-- reset status/attempts, synthesize receipts, or touch consumed/deleted uploads.
ALTER TABLE outbox_events
    ADD COLUMN IF NOT EXISTS storage_cleanup_last_reconciled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS storage_cleanup_reconciliation_error TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_outbox_original_storage_cleanup_reconciliation_due
    ON outbox_events (COALESCE(reconcile_after, processed_at, updated_at, created_at), id)
    WHERE deleted_at IS NULL
      AND event_type IN ('storage.object_cleanup_requested', 'customer_service.avatar_cleanup')
      AND status IN ('processed', 'unknown', 'dead_letter');
