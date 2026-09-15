CREATE TABLE IF NOT EXISTS shipping_quote_snapshots (
    id VARCHAR(36) PRIMARY KEY,
    request_hash CHAR(64) NOT NULL,
    rate_version CHAR(64) NOT NULL,
    quote_data JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shipping_quote_snapshots_request_hash
    ON shipping_quote_snapshots (request_hash);
CREATE INDEX IF NOT EXISTS idx_shipping_quote_snapshots_expires_at
    ON shipping_quote_snapshots (expires_at);

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS shipping_quote_id VARCHAR(36),
    ADD COLUMN IF NOT EXISTS shipping_quote_plan_id VARCHAR(36),
    ADD COLUMN IF NOT EXISTS shipping_plan_snapshot JSONB NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_orders_shipping_quote_id ON orders (shipping_quote_id);
