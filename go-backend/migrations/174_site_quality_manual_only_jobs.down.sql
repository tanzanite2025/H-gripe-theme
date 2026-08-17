ALTER TABLE site_quality_jobs
    DROP CONSTRAINT IF EXISTS ck_site_quality_jobs_status;

UPDATE site_quality_jobs
SET
    status = 'queued',
    finished_at = NULL,
    last_error = '',
    available_at = COALESCE(available_at, NOW()),
    updated_at = NOW()
WHERE kind = 'scheduled'
  AND status = 'cancelled';

UPDATE site_quality_jobs
SET
    status = 'dead_letter',
    last_error = 'manual-only Site Quality migration rollback',
    updated_at = NOW()
WHERE status = 'cancelled';

ALTER TABLE site_quality_jobs
    ADD CONSTRAINT ck_site_quality_jobs_status
    CHECK (status IN ('queued', 'processing', 'succeeded', 'failed', 'dead_letter'));
