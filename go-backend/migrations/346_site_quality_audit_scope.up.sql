ALTER TABLE site_quality_jobs
    ADD COLUMN IF NOT EXISTS audit_scope VARCHAR(16) NOT NULL DEFAULT 'full';

ALTER TABLE site_quality_jobs
    ADD CONSTRAINT ck_site_quality_jobs_audit_scope
        CHECK (audit_scope IN ('full', 'headings', 'schema', 'link_text'));

CREATE INDEX IF NOT EXISTS idx_site_quality_jobs_audit_scope
    ON site_quality_jobs (audit_scope, created_at DESC);
