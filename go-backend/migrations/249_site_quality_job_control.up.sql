ALTER TABLE site_quality_jobs
    ADD COLUMN IF NOT EXISTS progress_total INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS completed_samples INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS progress_stage VARCHAR(32) NOT NULL DEFAULT 'queued';

UPDATE site_quality_jobs
SET
    progress_total = CASE
        WHEN progress_total > 0 THEN progress_total
        ELSE sample_count
    END,
    completed_samples = CASE
        WHEN status = 'succeeded' THEN sample_count
        ELSE LEAST(completed_samples, sample_count)
    END,
    progress_stage = CASE status
        WHEN 'queued' THEN 'queued'
        WHEN 'processing' THEN 'starting'
        WHEN 'succeeded' THEN 'completed'
        WHEN 'cancelled' THEN 'cancelled'
        ELSE 'failed'
    END;

ALTER TABLE site_quality_jobs
    DROP CONSTRAINT IF EXISTS ck_site_quality_jobs_status;

ALTER TABLE site_quality_jobs
    ADD CONSTRAINT ck_site_quality_jobs_status
        CHECK (status IN ('queued', 'processing', 'succeeded', 'failed', 'dead_letter', 'cancelled'));

ALTER TABLE site_quality_jobs
    ADD CONSTRAINT ck_site_quality_jobs_progress
        CHECK (
            progress_total >= 0
            AND completed_samples >= 0
            AND completed_samples <= progress_total
        );

CREATE INDEX IF NOT EXISTS idx_site_quality_jobs_progress
    ON site_quality_jobs (status, progress_stage, updated_at DESC);
