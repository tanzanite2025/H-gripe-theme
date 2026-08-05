-- Append-only storefront behavior facts for recommendation analytics.
-- This table deliberately stores no IP address, user-agent, payment data, or
-- customer-service transcript. Identity is limited to first-party IDs supplied
-- by the storefront and an optional authenticated user ID from the session.

CREATE TABLE IF NOT EXISTS recommendation_events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(128) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    anonymous_id VARCHAR(128),
    session_id VARCHAR(128),
    user_id BIGINT,
    product_id BIGINT,
    category_id BIGINT,
    locale VARCHAR(20),
    path VARCHAR(1024),
    referrer VARCHAR(1024),
    metadata_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    occurred_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_recommendation_events_event_id
    ON recommendation_events(event_id);

CREATE INDEX IF NOT EXISTS idx_recommendation_events_type_occurred_at
    ON recommendation_events(event_type, occurred_at);

CREATE INDEX IF NOT EXISTS idx_recommendation_events_anonymous_occurred_at
    ON recommendation_events(anonymous_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_recommendation_events_session_occurred_at
    ON recommendation_events(session_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_recommendation_events_user_occurred_at
    ON recommendation_events(user_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_recommendation_events_product_occurred_at
    ON recommendation_events(product_id, occurred_at);
