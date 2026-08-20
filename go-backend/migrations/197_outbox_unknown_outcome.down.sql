DROP INDEX IF EXISTS idx_outbox_events_unknown_reconcile;

ALTER TABLE outbox_events
    DROP COLUMN IF EXISTS reconcile_after,
    DROP COLUMN IF EXISTS uncertain_at,
    DROP COLUMN IF EXISTS last_attempt_at;
