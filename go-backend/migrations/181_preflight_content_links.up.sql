CREATE TABLE IF NOT EXISTS preflight_content_link_runs (
    id BIGSERIAL PRIMARY KEY,
    target_url TEXT NOT NULL,
    route_entry_id BIGINT,
    status VARCHAR(24) NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL,
    issue_count INTEGER NOT NULL DEFAULT 0,
    fixable_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_preflight_content_link_runs_route_entry
        FOREIGN KEY (route_entry_id)
        REFERENCES storefront_route_catalog_entries(id)
        ON DELETE SET NULL,
    CONSTRAINT ck_preflight_content_link_runs_status
        CHECK (status IN ('success', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_preflight_content_link_runs_checked
    ON preflight_content_link_runs(checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_preflight_content_link_runs_target
    ON preflight_content_link_runs(target_url, checked_at DESC);

CREATE TABLE IF NOT EXISTS preflight_content_link_issues (
    id BIGSERIAL PRIMARY KEY,
    route_entry_id BIGINT,
    run_id BIGINT NOT NULL,
    target_url TEXT NOT NULL,
    final_url TEXT NOT NULL DEFAULT '',
    link_url TEXT NOT NULL DEFAULT '',
    link_text TEXT NOT NULL DEFAULT '',
    selector TEXT NOT NULL DEFAULT '',
    snippet TEXT NOT NULL DEFAULT '',
    source_type VARCHAR(48) NOT NULL DEFAULT '',
    source_id BIGINT,
    source_key TEXT NOT NULL DEFAULT '',
    source_field VARCHAR(80) NOT NULL DEFAULT '',
    issue_key VARCHAR(128) NOT NULL,
    severity VARCHAR(16) NOT NULL DEFAULT 'medium',
    state VARCHAR(24) NOT NULL DEFAULT 'open',
    suggested_text TEXT NOT NULL DEFAULT '',
    fix_status VARCHAR(24) NOT NULL DEFAULT 'not_fixable',
    fix_error TEXT NOT NULL DEFAULT '',
    latest_evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    first_detected_at TIMESTAMPTZ NOT NULL,
    last_detected_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    fixed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_preflight_content_link_issues_issue_key
        UNIQUE (issue_key),
    CONSTRAINT fk_preflight_content_link_issues_route_entry
        FOREIGN KEY (route_entry_id)
        REFERENCES storefront_route_catalog_entries(id)
        ON DELETE SET NULL,
    CONSTRAINT fk_preflight_content_link_issues_run
        FOREIGN KEY (run_id)
        REFERENCES preflight_content_link_runs(id)
        ON DELETE RESTRICT,
    CONSTRAINT ck_preflight_content_link_issues_severity
        CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT ck_preflight_content_link_issues_state
        CHECK (state IN ('open', 'resolved', 'verified', 'ignored')),
    CONSTRAINT ck_preflight_content_link_issues_fix_status
        CHECK (fix_status IN ('not_fixable', 'pending', 'applied', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issues_state_detected
    ON preflight_content_link_issues(state, last_detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issues_target
    ON preflight_content_link_issues(target_url, state);
CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issues_route
    ON preflight_content_link_issues(route_entry_id, state);
CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issues_source
    ON preflight_content_link_issues(source_type, source_id, source_field);

CREATE TABLE IF NOT EXISTS preflight_content_link_issue_events (
    id BIGSERIAL PRIMARY KEY,
    issue_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    actor_user_id BIGINT NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_preflight_content_link_issue_events_issue
        FOREIGN KEY (issue_id)
        REFERENCES preflight_content_link_issues(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issue_events_issue_created
    ON preflight_content_link_issue_events(issue_id, created_at DESC);
