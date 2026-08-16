ALTER TABLE tickets
    DROP CONSTRAINT IF EXISTS ck_tickets_status_version_positive;

ALTER TABLE tickets
    DROP COLUMN IF EXISTS status_version;
