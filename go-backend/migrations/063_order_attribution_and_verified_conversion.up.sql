CREATE TABLE IF NOT EXISTS order_attributions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    source VARCHAR(96) NOT NULL DEFAULT '',
    medium VARCHAR(96) NOT NULL DEFAULT '',
    campaign VARCHAR(160) NOT NULL DEFAULT '',
    term VARCHAR(160) NOT NULL DEFAULT '',
    content VARCHAR(160) NOT NULL DEFAULT '',
    click_id_kind VARCHAR(32) NOT NULL DEFAULT '',
    click_id VARCHAR(256) NOT NULL DEFAULT '',
    captured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_order_attributions_order_id
    ON order_attributions(order_id);

CREATE INDEX IF NOT EXISTS idx_order_attributions_captured_at
    ON order_attributions(captured_at);
