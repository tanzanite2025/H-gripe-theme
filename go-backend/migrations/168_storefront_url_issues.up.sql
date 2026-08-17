CREATE TABLE IF NOT EXISTS storefront_url_issues (
    id BIGSERIAL PRIMARY KEY,
    route_entry_id BIGINT NOT NULL,
    issue_type VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    state VARCHAR(32) NOT NULL DEFAULT 'open',
    assignee_id BIGINT,
    resolution_type VARCHAR(64) NOT NULL DEFAULT '',
    resolution_note TEXT NOT NULL DEFAULT '',
    linked_redirect_rule_id BIGINT,
    latest_check_result_id BIGINT,
    first_detected_at TIMESTAMPTZ NOT NULL,
    last_detected_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    suppressed_until TIMESTAMPTZ,
    suppression_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_storefront_url_issues_route_type
        UNIQUE (route_entry_id, issue_type),
    CONSTRAINT fk_storefront_url_issues_route_entry
        FOREIGN KEY (route_entry_id)
        REFERENCES storefront_route_catalog_entries(id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_storefront_url_issues_redirect_rule
        FOREIGN KEY (linked_redirect_rule_id)
        REFERENCES storefront_redirect_rules(id)
        ON DELETE SET NULL,
    CONSTRAINT fk_storefront_url_issues_latest_check_result
        FOREIGN KEY (latest_check_result_id)
        REFERENCES storefront_route_check_results(id)
        ON DELETE SET NULL,
    CONSTRAINT ck_storefront_url_issues_severity
        CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT ck_storefront_url_issues_state
        CHECK (state IN ('open', 'acknowledged', 'resolved', 'verified', 'suppressed'))
);

CREATE INDEX IF NOT EXISTS idx_storefront_url_issues_state_severity_detected
    ON storefront_url_issues (state, severity, last_detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_storefront_url_issues_assignee_state
    ON storefront_url_issues (assignee_id, state);
CREATE INDEX IF NOT EXISTS idx_storefront_url_issues_route_entry
    ON storefront_url_issues (route_entry_id);
CREATE INDEX IF NOT EXISTS idx_storefront_url_issues_redirect_rule
    ON storefront_url_issues (linked_redirect_rule_id)
    WHERE linked_redirect_rule_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS storefront_url_issue_events (
    id BIGSERIAL PRIMARY KEY,
    issue_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    actor_user_id BIGINT NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_storefront_url_issue_events_issue
        FOREIGN KEY (issue_id)
        REFERENCES storefront_url_issues(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_storefront_url_issue_events_issue_created
    ON storefront_url_issue_events (issue_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_storefront_url_issue_events_type_created
    ON storefront_url_issue_events (event_type, created_at DESC);
