CREATE TABLE IF NOT EXISTS refund_idempotencies (
    id BIGSERIAL PRIMARY KEY,
    admin_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    refund_id BIGINT REFERENCES refunds(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_refund_idempotencies_scope_key UNIQUE (scope, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_refund_idempotencies_refund_id
    ON refund_idempotencies(refund_id);

CREATE INDEX IF NOT EXISTS idx_refund_idempotencies_created_at
    ON refund_idempotencies(created_at);
