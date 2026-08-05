CREATE TABLE IF NOT EXISTS payment_refund_executions (
    id BIGSERIAL PRIMARY KEY,
    refund_id BIGINT NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    transaction_id BIGINT NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    provider_payment_id VARCHAR(255) NOT NULL,
    amount NUMERIC(18, 2) NOT NULL,
    currency VARCHAR(8) NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'processing',
    idempotency_key VARCHAR(128) NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 1,
    requested_by_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    provider_refund_id VARCHAR(255) NOT NULL DEFAULT '',
    provider_status VARCHAR(64) NOT NULL DEFAULT '',
    gateway_response_json TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_payment_refund_executions_refund UNIQUE (refund_id),
    CONSTRAINT uq_payment_refund_executions_idempotency UNIQUE (idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_payment_refund_executions_status_created
    ON payment_refund_executions(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_refund_executions_provider_payment
    ON payment_refund_executions(provider, provider_payment_id);

CREATE INDEX IF NOT EXISTS idx_payment_refund_executions_provider_refund
    ON payment_refund_executions(provider_refund_id);
