-- Provider-neutral payment-risk event ledger and rolling metric snapshots.
-- Threshold interpretation stays in application configuration, not in schema.

CREATE TABLE IF NOT EXISTS payment_risk_events (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    kind VARCHAR(64) NOT NULL,
    external_reference VARCHAR(255) NOT NULL,
    webhook_event_id VARCHAR(255) NOT NULL DEFAULT '',
    payment_intent_id VARCHAR(255) NOT NULL DEFAULT '',
    charge_id VARCHAR(255) NOT NULL DEFAULT '',
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    transaction_id BIGINT REFERENCES transactions(id) ON DELETE SET NULL,
    amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    payload TEXT NOT NULL DEFAULT '',
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_payment_risk_event_reference UNIQUE (provider, kind, external_reference)
);

CREATE INDEX IF NOT EXISTS idx_payment_risk_events_provider_kind_occurred
    ON payment_risk_events(provider, kind, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_risk_events_payment_intent
    ON payment_risk_events(payment_intent_id);

CREATE TABLE IF NOT EXISTS payment_risk_snapshots (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    window_days INTEGER NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    successful_payment_count BIGINT NOT NULL DEFAULT 0,
    successful_payment_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    dispute_count BIGINT NOT NULL DEFAULT 0,
    dispute_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    early_fraud_warning_count BIGINT NOT NULL DEFAULT 0,
    refund_count BIGINT NOT NULL DEFAULT 0,
    refund_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    dispute_activity_rate NUMERIC(12, 8) NOT NULL DEFAULT 0,
    early_fraud_warning_rate NUMERIC(12, 8) NOT NULL DEFAULT 0,
    refund_rate NUMERIC(12, 8) NOT NULL DEFAULT 0,
    level VARCHAR(16) NOT NULL DEFAULT 'normal',
    recommended_action VARCHAR(160) NOT NULL DEFAULT '',
    reasons_json TEXT NOT NULL DEFAULT '[]',
    computed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_risk_snapshots_provider_computed
    ON payment_risk_snapshots(provider, computed_at DESC);
