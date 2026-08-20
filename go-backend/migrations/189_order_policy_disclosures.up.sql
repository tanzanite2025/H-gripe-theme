CREATE TABLE IF NOT EXISTS order_policy_disclosures (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    policy_key VARCHAR(128) NOT NULL,
    locale VARCHAR(32) NOT NULL,
    requested_locale VARCHAR(32) NOT NULL,
    fallback BOOLEAN NOT NULL DEFAULT FALSE,
    policy_version VARCHAR(128) NOT NULL,
    policy_hash VARCHAR(128) NOT NULL,
    policy_json TEXT NOT NULL,
    policy_url TEXT NOT NULL,
    disclosed_at TIMESTAMPTZ NOT NULL,
    consented_at TIMESTAMPTZ,
    source VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_order_policy_disclosures_order_policy UNIQUE (order_id, policy_key)
);

CREATE INDEX IF NOT EXISTS idx_order_policy_disclosures_order_id
    ON order_policy_disclosures(order_id);

CREATE INDEX IF NOT EXISTS idx_order_policy_disclosures_policy_hash
    ON order_policy_disclosures(policy_key, policy_hash);
