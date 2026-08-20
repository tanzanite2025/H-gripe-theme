ALTER TABLE outbox_events
    ADD COLUMN IF NOT EXISTS last_attempt_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS uncertain_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reconcile_after TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_outbox_events_unknown_reconcile
    ON outbox_events(status, reconcile_after, id)
    WHERE status = 'unknown' AND deleted_at IS NULL;
