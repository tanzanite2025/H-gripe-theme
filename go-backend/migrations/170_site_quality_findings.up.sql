CREATE TABLE IF NOT EXISTS site_quality_findings (
    id BIGSERIAL PRIMARY KEY,
    target_id BIGINT,
    target_url TEXT NOT NULL,
    strategy VARCHAR(16) NOT NULL,
    audit_id VARCHAR(128) NOT NULL,
    finding_kind VARCHAR(32) NOT NULL DEFAULT 'opportunity',
    rule_version VARCHAR(64) NOT NULL DEFAULT '',
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    sample_count INTEGER NOT NULL DEFAULT 1,
    confirmations INTEGER NOT NULL DEFAULT 1,
    consecutive_clean INTEGER NOT NULL DEFAULT 0,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    severity VARCHAR(16) NOT NULL,
    state VARCHAR(32) NOT NULL DEFAULT 'open',
    first_detected_at TIMESTAMPTZ NOT NULL,
    last_detected_at TIMESTAMPTZ NOT NULL,
    latest_run_id BIGINT NOT NULL,
    latest_savings_ms DOUBLE PRECISION,
    latest_savings_bytes BIGINT,
    resource_count INTEGER NOT NULL DEFAULT 0,
    latest_evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    resolution_note TEXT NOT NULL DEFAULT '',
    resolved_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_site_quality_findings_url_strategy_audit
        UNIQUE (target_url, strategy, audit_id),
    CONSTRAINT fk_site_quality_findings_latest_run
        FOREIGN KEY (latest_run_id)
        REFERENCES site_quality_runs(id)
        ON DELETE RESTRICT,
    CONSTRAINT ck_site_quality_findings_strategy
        CHECK (strategy IN ('mobile', 'desktop')),
    CONSTRAINT ck_site_quality_findings_severity
        CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT ck_site_quality_findings_state
        CHECK (state IN ('open', 'acknowledged', 'resolved', 'verified'))
);

CREATE INDEX IF NOT EXISTS idx_site_quality_findings_state_detected
    ON site_quality_findings (state, last_detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_site_quality_findings_target_strategy
    ON site_quality_findings (target_url, strategy);
CREATE INDEX IF NOT EXISTS idx_site_quality_findings_latest_run
    ON site_quality_findings (latest_run_id);

CREATE TABLE IF NOT EXISTS site_quality_finding_events (
    id BIGSERIAL PRIMARY KEY,
    finding_id BIGINT NOT NULL,
    run_id BIGINT,
    event_type VARCHAR(64) NOT NULL,
    actor_user_id BIGINT NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_site_quality_finding_events_finding
        FOREIGN KEY (finding_id)
        REFERENCES site_quality_findings(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_site_quality_finding_events_run
        FOREIGN KEY (run_id)
        REFERENCES site_quality_runs(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_site_quality_finding_events_finding_created
    ON site_quality_finding_events (finding_id, created_at DESC);
