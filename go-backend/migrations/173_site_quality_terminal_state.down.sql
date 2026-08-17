DROP INDEX IF EXISTS idx_site_quality_jobs_lease;
DROP INDEX IF EXISTS idx_site_quality_jobs_finding;

ALTER TABLE site_quality_jobs
    DROP CONSTRAINT IF EXISTS fk_site_quality_jobs_finding;

ALTER TABLE site_quality_jobs
    DROP COLUMN IF EXISTS heartbeat_at,
    DROP COLUMN IF EXISTS lease_expires_at,
    DROP COLUMN IF EXISTS lease_generation,
    DROP COLUMN IF EXISTS finding_id;

DROP INDEX IF EXISTS idx_site_quality_targets_source_ledger;
ALTER TABLE site_quality_targets
    DROP CONSTRAINT IF EXISTS ck_site_quality_targets_source;

ALTER TABLE site_quality_targets
    DROP COLUMN IF EXISTS disable_reason,
    DROP COLUMN IF EXISTS disabled_at,
    DROP COLUMN IF EXISTS ledger_synced_at,
    DROP COLUMN IF EXISTS ledger_sync_marker,
    DROP COLUMN IF EXISTS ledger_synced,
    DROP COLUMN IF EXISTS source;
