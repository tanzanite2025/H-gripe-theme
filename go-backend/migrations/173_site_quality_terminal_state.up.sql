ALTER TABLE site_quality_targets
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'operator',
    ADD COLUMN IF NOT EXISTS ledger_synced BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS ledger_sync_marker VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ledger_synced_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS disabled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS disable_reason TEXT NOT NULL DEFAULT '';

UPDATE site_quality_targets
SET
    source = 'route_catalog',
    ledger_synced = TRUE,
    ledger_synced_at = COALESCE(ledger_synced_at, updated_at)
WHERE route_entry_id IS NOT NULL;

ALTER TABLE site_quality_targets
    ADD CONSTRAINT ck_site_quality_targets_source
        CHECK (source IN ('route_catalog', 'operator'));

CREATE INDEX IF NOT EXISTS idx_site_quality_targets_source_ledger
    ON site_quality_targets (source, ledger_synced, ledger_sync_marker);

ALTER TABLE site_quality_jobs
    ADD COLUMN IF NOT EXISTS finding_id BIGINT,
    ADD COLUMN IF NOT EXISTS lease_generation BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lease_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS heartbeat_at TIMESTAMPTZ;

ALTER TABLE site_quality_jobs
    ADD CONSTRAINT fk_site_quality_jobs_finding
        FOREIGN KEY (finding_id)
        REFERENCES site_quality_findings(id)
        ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_site_quality_jobs_finding
    ON site_quality_jobs (finding_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_site_quality_jobs_lease
    ON site_quality_jobs (status, lease_expires_at, lease_generation);
