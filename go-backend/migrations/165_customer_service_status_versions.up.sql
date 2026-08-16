-- A monotonic per-conversation version makes status-only realtime events
-- idempotent and replayable. Existing rows start at the first version.
ALTER TABLE tickets
    ADD COLUMN IF NOT EXISTS status_version INTEGER NOT NULL DEFAULT 1;

UPDATE tickets
SET status_version = 1
WHERE status_version IS NULL OR status_version < 1;

ALTER TABLE tickets
    DROP CONSTRAINT IF EXISTS ck_tickets_status_version_positive;

ALTER TABLE tickets
    ADD CONSTRAINT ck_tickets_status_version_positive
        CHECK (status_version > 0);
