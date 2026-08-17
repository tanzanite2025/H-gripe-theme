CREATE TABLE IF NOT EXISTS site_quality_targets (
    id BIGSERIAL PRIMARY KEY,
    route_entry_id BIGINT UNIQUE,
    canonical_url TEXT NOT NULL UNIQUE,
    locale VARCHAR(20) NOT NULL DEFAULT '',
    source_type VARCHAR(32) NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    sampling_tier VARCHAR(16) NOT NULL DEFAULT 'standard',
    sampling_interval_seconds INTEGER NOT NULL DEFAULT 604800,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    next_scheduled_at TIMESTAMPTZ,
    last_scheduled_at TIMESTAMPTZ,
    last_completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_site_quality_targets_route_entry
        FOREIGN KEY (route_entry_id)
        REFERENCES storefront_route_catalog_entries(id)
        ON DELETE SET NULL,
    CONSTRAINT ck_site_quality_targets_tier
        CHECK (sampling_tier IN ('critical', 'standard', 'background')),
    CONSTRAINT ck_site_quality_targets_interval
        CHECK (sampling_interval_seconds > 0)
);

CREATE INDEX IF NOT EXISTS idx_site_quality_targets_due
    ON site_quality_targets (enabled, next_scheduled_at);
CREATE INDEX IF NOT EXISTS idx_site_quality_targets_tier_locale
    ON site_quality_targets (sampling_tier, locale);

CREATE TABLE IF NOT EXISTS site_quality_jobs (
    id BIGSERIAL PRIMARY KEY,
    target_id BIGINT NOT NULL,
    strategy VARCHAR(16) NOT NULL,
    kind VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'queued',
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    sample_count INTEGER NOT NULL DEFAULT 3,
    required_confirmations INTEGER NOT NULL DEFAULT 2,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 4,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(128) NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    initiated_by_user_id BIGINT NOT NULL DEFAULT 0,
    release_id VARCHAR(128) NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_site_quality_jobs_target
        FOREIGN KEY (target_id)
        REFERENCES site_quality_targets(id)
        ON DELETE CASCADE,
    CONSTRAINT ck_site_quality_jobs_strategy
        CHECK (strategy IN ('mobile', 'desktop')),
    CONSTRAINT ck_site_quality_jobs_kind
        CHECK (kind IN ('scheduled', 'manual', 'recheck')),
    CONSTRAINT ck_site_quality_jobs_status
        CHECK (status IN ('queued', 'processing', 'succeeded', 'failed', 'dead_letter')),
    CONSTRAINT ck_site_quality_jobs_samples
        CHECK (
            sample_count > 0
            AND required_confirmations > 0
            AND required_confirmations <= sample_count
            AND max_attempts > 0
        )
);

CREATE INDEX IF NOT EXISTS idx_site_quality_jobs_claim
    ON site_quality_jobs (status, available_at, id);
CREATE INDEX IF NOT EXISTS idx_site_quality_jobs_target_strategy
    ON site_quality_jobs (target_id, strategy, created_at DESC);

CREATE TABLE IF NOT EXISTS site_quality_provider_slots (
    provider VARCHAR(64) NOT NULL,
    slot INTEGER NOT NULL,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (provider, slot),
    CONSTRAINT ck_site_quality_provider_slots_slot
        CHECK (slot > 0)
);

CREATE INDEX IF NOT EXISTS idx_site_quality_provider_slots_claim
    ON site_quality_provider_slots (provider, available_at, locked_at);

CREATE TABLE IF NOT EXISTS site_quality_evaluations (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL UNIQUE,
    target_id BIGINT NOT NULL,
    strategy VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL,
    sample_count INTEGER NOT NULL,
    successful_samples INTEGER NOT NULL,
    required_confirmations INTEGER NOT NULL,
    confirmed_audit_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    clean_audit_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    decision_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    decided_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_site_quality_evaluations_job
        FOREIGN KEY (job_id)
        REFERENCES site_quality_jobs(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_site_quality_evaluations_target
        FOREIGN KEY (target_id)
        REFERENCES site_quality_targets(id)
        ON DELETE CASCADE,
    CONSTRAINT ck_site_quality_evaluations_strategy
        CHECK (strategy IN ('mobile', 'desktop')),
    CONSTRAINT ck_site_quality_evaluations_status
        CHECK (status IN ('completed', 'insufficient_samples', 'failed')),
    CONSTRAINT ck_site_quality_evaluations_samples
        CHECK (
            sample_count > 0
            AND successful_samples >= 0
            AND successful_samples <= sample_count
            AND required_confirmations > 0
            AND required_confirmations <= sample_count
        )
);

CREATE INDEX IF NOT EXISTS idx_site_quality_evaluations_target_strategy_decided
    ON site_quality_evaluations (target_id, strategy, decided_at DESC);

INSERT INTO site_quality_targets (
    canonical_url,
    sampling_tier,
    sampling_interval_seconds,
    enabled,
    next_scheduled_at
)
SELECT DISTINCT candidate_url, 'standard', 604800, TRUE, NOW()
FROM (
    SELECT NULLIF(canonical_url, '') AS candidate_url
    FROM site_quality_runs
    UNION
    SELECT NULLIF(target_url, '') AS candidate_url
    FROM site_quality_runs
    UNION
    SELECT NULLIF(target_url, '') AS candidate_url
    FROM site_quality_findings
) AS legacy_targets
WHERE candidate_url IS NOT NULL
ON CONFLICT (canonical_url) DO NOTHING;

UPDATE site_quality_runs AS run
SET
    target_id = target.id,
    canonical_url = target.canonical_url
FROM site_quality_targets AS target
WHERE run.target_id IS NULL
  AND target.canonical_url = CASE
      WHEN run.canonical_url <> '' THEN run.canonical_url
      ELSE run.target_url
  END;

UPDATE site_quality_findings AS finding
SET target_id = target.id
FROM site_quality_targets AS target
WHERE finding.target_id IS NULL
  AND target.canonical_url = finding.target_url;

ALTER TABLE site_quality_runs
    ADD CONSTRAINT fk_site_quality_runs_target
        FOREIGN KEY (target_id)
        REFERENCES site_quality_targets(id)
        ON DELETE SET NULL;
ALTER TABLE site_quality_runs
    ADD CONSTRAINT fk_site_quality_runs_job
        FOREIGN KEY (job_id)
        REFERENCES site_quality_jobs(id)
        ON DELETE SET NULL;
ALTER TABLE site_quality_findings
    ADD CONSTRAINT fk_site_quality_findings_target
        FOREIGN KEY (target_id)
        REFERENCES site_quality_targets(id)
        ON DELETE SET NULL;

ALTER TABLE site_quality_findings
    DROP CONSTRAINT IF EXISTS uq_site_quality_findings_url_strategy_audit;
CREATE UNIQUE INDEX IF NOT EXISTS uq_site_quality_findings_target_strategy_audit
    ON site_quality_findings (target_id, strategy, audit_id);

CREATE INDEX IF NOT EXISTS idx_site_quality_runs_target_strategy_created_v2
    ON site_quality_runs (target_id, strategy, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_site_quality_findings_target_strategy_v2
    ON site_quality_findings (target_id, strategy);
