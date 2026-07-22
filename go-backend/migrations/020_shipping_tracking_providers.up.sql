ALTER TABLE carriers DROP COLUMN IF EXISTS api_endpoint;
ALTER TABLE carriers DROP COLUMN IF EXISTS api_key;
ALTER TABLE carriers DROP COLUMN IF EXISTS api_secret;

CREATE TABLE IF NOT EXISTS shipping_tracking_providers (
    id BIGSERIAL PRIMARY KEY,
    provider_code VARCHAR(80) NOT NULL,
    provider_name VARCHAR(160) NOT NULL,
    environment VARCHAR(40) NOT NULL DEFAULT 'production',
    base_url TEXT,
    api_key TEXT,
    webhook_secret TEXT,
    webhook_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    auto_register BOOLEAN NOT NULL DEFAULT FALSE,
    polling_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    polling_interval_minutes INTEGER NOT NULL DEFAULT 60,
    request_timeout_seconds INTEGER NOT NULL DEFAULT 15,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS provider_code VARCHAR(80);
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS provider_name VARCHAR(160);
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS environment VARCHAR(40) DEFAULT 'production';
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS base_url TEXT;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS api_key TEXT;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS webhook_secret TEXT;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS webhook_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS auto_register BOOLEAN DEFAULT FALSE;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS polling_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS polling_interval_minutes INTEGER DEFAULT 60;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS request_timeout_seconds INTEGER DEFAULT 15;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS sort_order INTEGER DEFAULT 0;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_tracking_providers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_tracking_providers_code
    ON shipping_tracking_providers (provider_code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_providers_enabled ON shipping_tracking_providers (enabled);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_providers_deleted_at ON shipping_tracking_providers (deleted_at);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_providers_sort_order ON shipping_tracking_providers (sort_order);
