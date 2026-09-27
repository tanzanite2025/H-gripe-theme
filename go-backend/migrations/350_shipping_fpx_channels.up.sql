CREATE TABLE IF NOT EXISTS shipping_fpx_channels (
    id BIGSERIAL PRIMARY KEY,
    service_code VARCHAR(80) NOT NULL,
    display_name VARCHAR(160) NOT NULL,
    max_length_cm NUMERIC(10,2) NOT NULL DEFAULT 0,
    max_width_cm NUMERIC(10,2) NOT NULL DEFAULT 0,
    max_height_cm NUMERIC(10,2) NOT NULL DEFAULT 0,
    volumetric_divisor INTEGER NOT NULL DEFAULT 6000,
    max_weight_grams INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_fpx_channel_service_code
    ON shipping_fpx_channels (service_code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_shipping_fpx_channels_enabled
    ON shipping_fpx_channels (enabled);

CREATE INDEX IF NOT EXISTS idx_shipping_fpx_channels_sort_order
    ON shipping_fpx_channels (sort_order);
