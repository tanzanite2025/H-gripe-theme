-- Stores the most recently evaluated operational level for each provider.
-- This state makes level-transition alerts safe across repeated scheduler runs.

CREATE TABLE IF NOT EXISTS payment_risk_alert_states (
    provider VARCHAR(32) PRIMARY KEY,
    current_level VARCHAR(16) NOT NULL DEFAULT 'normal',
    current_snapshot_id BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
