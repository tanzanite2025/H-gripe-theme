CREATE TABLE IF NOT EXISTS payment_operation_idempotencies (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 0,
    content_type VARCHAR(255) NOT NULL DEFAULT '',
    response_body TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_payment_operation_idempotencies_scope_key UNIQUE (user_id, scope, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_payment_operation_idempotencies_created_at
    ON payment_operation_idempotencies(created_at);
