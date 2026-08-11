CREATE TABLE IF NOT EXISTS paypal_disputes (
    id BIGSERIAL PRIMARY KEY,
    paypal_dispute_id VARCHAR(255) NOT NULL UNIQUE,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    transaction_id BIGINT REFERENCES transactions(id) ON DELETE SET NULL,
    provider_payment_id VARCHAR(255) NOT NULL DEFAULT '',
    amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL,
    reason VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(64) NOT NULL DEFAULT '',
    dispute_state VARCHAR(80) NOT NULL DEFAULT '',
    dispute_life_cycle_stage VARCHAR(80) NOT NULL DEFAULT '',
    raw_payload TEXT NOT NULL DEFAULT '',
    evidence_submitted_at TIMESTAMPTZ,
    evidence_submission_payload TEXT NOT NULL DEFAULT '',
    evidence_submission_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_paypal_disputes_status_state
    ON paypal_disputes(status, dispute_state);

CREATE INDEX IF NOT EXISTS idx_paypal_disputes_provider_payment
    ON paypal_disputes(provider_payment_id);

CREATE INDEX IF NOT EXISTS idx_paypal_disputes_evidence_submitted_at
    ON paypal_disputes(evidence_submitted_at);
