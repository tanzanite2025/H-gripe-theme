-- Stripe event inbox, dispute monitoring, and manual payment review workflow.

CREATE TABLE IF NOT EXISTS stripe_webhook_events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL UNIQUE,
    event_type VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'processing',
    payload TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stripe_webhook_events_status_created
    ON stripe_webhook_events(status, created_at DESC);

CREATE TABLE IF NOT EXISTS stripe_disputes (
    id BIGSERIAL PRIMARY KEY,
    stripe_dispute_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_charge_id VARCHAR(255) NOT NULL DEFAULT '',
    payment_intent_id VARCHAR(255) NOT NULL DEFAULT '',
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    transaction_id BIGINT REFERENCES transactions(id) ON DELETE SET NULL,
    amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL,
    reason VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(64) NOT NULL DEFAULT 'needs_response',
    evidence_due_at TIMESTAMPTZ,
    raw_payload TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_stripe_disputes_status_due
    ON stripe_disputes(status, evidence_due_at);
CREATE INDEX IF NOT EXISTS idx_stripe_disputes_payment_intent
    ON stripe_disputes(payment_intent_id);

CREATE TABLE IF NOT EXISTS payment_reviews (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    transaction_id BIGINT REFERENCES transactions(id) ON DELETE SET NULL,
    dispute_id BIGINT REFERENCES stripe_disputes(id) ON DELETE SET NULL,
    payment_intent_id VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    reason VARCHAR(100) NOT NULL,
    source VARCHAR(32) NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    assigned_to_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_by_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_reviews_status_created
    ON payment_reviews(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_reviews_payment_intent
    ON payment_reviews(payment_intent_id);
