-- Operator review queue for refund recommendations created from provider risk
-- webhooks. The queue records recommendations only; it does not execute gateway
-- refunds.

CREATE TABLE IF NOT EXISTS payment_refund_recommendations (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    source_kind VARCHAR(64) NOT NULL,
    external_reference VARCHAR(255) NOT NULL,
    webhook_event_id VARCHAR(255) NOT NULL DEFAULT '',
    risk_event_id BIGINT REFERENCES payment_risk_events(id) ON DELETE SET NULL,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    transaction_id BIGINT REFERENCES transactions(id) ON DELETE SET NULL,
    provider_payment_id VARCHAR(255) NOT NULL DEFAULT '',
    payment_intent_id VARCHAR(255) NOT NULL DEFAULT '',
    charge_id VARCHAR(255) NOT NULL DEFAULT '',
    recommended_action VARCHAR(80) NOT NULL,
    recommended_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL DEFAULT '',
    priority VARCHAR(24) NOT NULL DEFAULT 'normal',
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    reason VARCHAR(160) NOT NULL,
    provider_reason VARCHAR(160) NOT NULL DEFAULT '',
    review_by TIMESTAMPTZ,
    decision_notes TEXT NOT NULL DEFAULT '',
    reviewed_by_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    source_metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_payment_refund_recommendation_source
        UNIQUE (provider, source_kind, external_reference)
);

CREATE INDEX IF NOT EXISTS idx_payment_refund_recommendations_status_created
    ON payment_refund_recommendations(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_refund_recommendations_provider_status
    ON payment_refund_recommendations(provider, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_refund_recommendations_order
    ON payment_refund_recommendations(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_refund_recommendations_provider_payment
    ON payment_refund_recommendations(provider_payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_refund_recommendations_review_by
    ON payment_refund_recommendations(review_by);
