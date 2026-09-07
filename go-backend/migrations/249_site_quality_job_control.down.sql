ALTER TABLE site_quality_jobs
    DROP CONSTRAINT IF EXISTS ck_site_quality_jobs_progress;

UPDATE site_quality_jobs
SET
    status = 'failed',
    finished_at = COALESCE(finished_at, NOW()),
    locked_at = NULL,
    locked_by = '',
    lease_expires_at = NULL,
    heartbeat_at = NULL,
    last_error = CASE
        WHEN status = 'cancelled' THEN 'cancelled job restored during migration rollback'
        ELSE last_error
    END,
    updated_at = NOW()
WHERE status = 'cancelled';

ALTER TABLE site_quality_jobs
    DROP CONSTRAINT IF EXISTS ck_site_quality_jobs_status;

ALTER TABLE site_quality_jobs
    ADD CONSTRAINT ck_site_quality_jobs_status
        CHECK (status IN ('queued', 'processing', 'succeeded', 'failed', 'dead_letter'));

DROP INDEX IF EXISTS idx_site_quality_jobs_progress;

ALTER TABLE site_quality_jobs
    DROP COLUMN IF EXISTS progress_stage,
    DROP COLUMN IF EXISTS completed_samples,
    DROP COLUMN IF EXISTS progress_total;
