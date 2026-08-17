ALTER TABLE site_quality_jobs
    DROP CONSTRAINT IF EXISTS ck_site_quality_jobs_status;

UPDATE site_quality_jobs
SET
    status = 'cancelled',
    locked_at = NULL,
    locked_by = '',
    lease_expires_at = NULL,
    heartbeat_at = NULL,
    finished_at = COALESCE(finished_at, NOW()),
    last_error = 'automatic Site Quality scheduling retired',
    updated_at = NOW()
WHERE kind = 'scheduled'
  AND status IN ('queued', 'processing', 'failed');

ALTER TABLE site_quality_jobs
    ADD CONSTRAINT ck_site_quality_jobs_status
    CHECK (status IN ('queued', 'processing', 'succeeded', 'failed', 'dead_letter', 'cancelled'));
