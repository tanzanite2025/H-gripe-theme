CREATE TABLE IF NOT EXISTS payment_risk_checkout_decisions (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    provider_payment_id VARCHAR(255) NOT NULL,
    mode VARCHAR(32) NOT NULL,
    strategy VARCHAR(80) NOT NULL,
    exemption_candidate BOOLEAN NOT NULL DEFAULT FALSE,
    risk_level VARCHAR(32) NOT NULL DEFAULT 'normal',
    risk_score INTEGER NOT NULL DEFAULT 0,
    portfolio_risk_level VARCHAR(32) NOT NULL DEFAULT 'normal',
    reasons_json TEXT NOT NULL DEFAULT '[]',
    amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_payment_risk_checkout_decision_provider_payment
        UNIQUE (provider, provider_payment_id)
);

CREATE INDEX IF NOT EXISTS idx_payment_risk_checkout_decisions_provider_occurred
    ON payment_risk_checkout_decisions(provider, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_risk_checkout_decisions_strategy
    ON payment_risk_checkout_decisions(provider, strategy, occurred_at DESC);

ALTER TABLE payment_risk_snapshots
    ADD COLUMN IF NOT EXISTS checkout_attempt_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS three_ds_upgrade_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS three_ds_challenge_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS three_ds_exemption_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS three_ds_upgrade_rate NUMERIC(12, 8) NOT NULL DEFAULT 0;
