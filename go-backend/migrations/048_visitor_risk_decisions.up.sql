-- Manual visitor-risk decisions are historical records. They do not directly
-- enforce blocking or rate limiting; enforcement can consume them later.

CREATE TABLE IF NOT EXISTS visitor_risk_decisions (
    id BIGSERIAL PRIMARY KEY,
    scope VARCHAR(24) NOT NULL,
    value_hash VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL,
    reason VARCHAR(500) NOT NULL,
    expires_at TIMESTAMPTZ,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_decisions_scope_value
    ON visitor_risk_decisions(scope, value_hash, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_decisions_expiry
    ON visitor_risk_decisions(expires_at);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_decisions_created_by
    ON visitor_risk_decisions(created_by);
