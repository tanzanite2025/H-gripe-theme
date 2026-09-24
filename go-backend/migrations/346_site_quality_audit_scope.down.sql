DROP INDEX IF EXISTS idx_site_quality_jobs_audit_scope;
ALTER TABLE site_quality_jobs
    DROP CONSTRAINT IF EXISTS ck_site_quality_jobs_audit_scope;
ALTER TABLE site_quality_jobs
    DROP COLUMN IF EXISTS audit_scope;
