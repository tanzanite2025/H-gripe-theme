CREATE TABLE IF NOT EXISTS shipping_yanwen_published_channels (
    id BIGSERIAL PRIMARY KEY,
    product_code VARCHAR(80) NOT NULL,
    display_name VARCHAR(160) NOT NULL,
    package_type VARCHAR(80) NOT NULL DEFAULT '',
    max_weight_grams INTEGER NOT NULL DEFAULT 0,
    volumetric_divisor INTEGER NOT NULL DEFAULT 8000,
    notes TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_yanwen_channel_product_code
    ON shipping_yanwen_published_channels (product_code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_shipping_yanwen_channels_enabled
    ON shipping_yanwen_published_channels (enabled);
