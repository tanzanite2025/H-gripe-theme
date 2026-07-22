CREATE TABLE IF NOT EXISTS shipping_tracking_carrier_mappings (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL REFERENCES shipping_tracking_providers(id) ON DELETE CASCADE,
    scope VARCHAR(40) NOT NULL DEFAULT 'carrier',
    carrier_id BIGINT REFERENCES carriers(id) ON DELETE SET NULL,
    carrier_service_id BIGINT REFERENCES shipping_carrier_services(id) ON DELETE SET NULL,
    provider_carrier_code VARCHAR(120) NOT NULL,
    provider_carrier_name VARCHAR(160),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS provider_id BIGINT;
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS scope VARCHAR(40) DEFAULT 'carrier';
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS carrier_id BIGINT;
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS carrier_service_id BIGINT;
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS provider_carrier_code VARCHAR(120);
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS provider_carrier_name VARCHAR(160);
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_tracking_carrier_mappings ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_provider_id
    ON shipping_tracking_carrier_mappings (provider_id);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_scope
    ON shipping_tracking_carrier_mappings (scope);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_carrier_id
    ON shipping_tracking_carrier_mappings (carrier_id);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_service_id
    ON shipping_tracking_carrier_mappings (carrier_service_id);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_provider_code
    ON shipping_tracking_carrier_mappings (provider_carrier_code);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_enabled
    ON shipping_tracking_carrier_mappings (enabled);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_deleted_at
    ON shipping_tracking_carrier_mappings (deleted_at);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_priority
    ON shipping_tracking_carrier_mappings (priority);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_provider_carrier
    ON shipping_tracking_carrier_mappings (provider_id, carrier_id)
    WHERE scope = 'carrier' AND carrier_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_tracking_carrier_mappings_provider_service
    ON shipping_tracking_carrier_mappings (provider_id, carrier_service_id)
    WHERE scope = 'carrier_service' AND carrier_service_id IS NOT NULL AND deleted_at IS NULL;
