CREATE TABLE IF NOT EXISTS shipping_tracking_shipments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    tracking_provider_id BIGINT NOT NULL REFERENCES shipping_tracking_providers(id) ON DELETE RESTRICT,
    tracking_number VARCHAR(120) NOT NULL,
    provider_carrier_code VARCHAR(120) NOT NULL,
    carrier_id BIGINT REFERENCES carriers(id) ON DELETE SET NULL,
    carrier_service_id BIGINT REFERENCES shipping_carrier_services(id) ON DELETE SET NULL,
    tracking_carrier_mapping_id BIGINT REFERENCES shipping_tracking_carrier_mappings(id) ON DELETE SET NULL,
    registration_status VARCHAR(40) NOT NULL DEFAULT 'pending',
    sync_status VARCHAR(40) NOT NULL DEFAULT 'pending',
    event_count INTEGER NOT NULL DEFAULT 0,
    last_event_at TIMESTAMPTZ,
    last_synced_at TIMESTAMPTZ,
    next_sync_at TIMESTAMPTZ,
    last_error TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS order_id BIGINT;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS tracking_provider_id BIGINT;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(120);
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS provider_carrier_code VARCHAR(120);
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS carrier_id BIGINT;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS carrier_service_id BIGINT;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS tracking_carrier_mapping_id BIGINT;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS registration_status VARCHAR(40) DEFAULT 'pending';
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS sync_status VARCHAR(40) DEFAULT 'pending';
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS event_count INTEGER DEFAULT 0;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS last_event_at TIMESTAMPTZ;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS next_sync_at TIMESTAMPTZ;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS last_error TEXT;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_tracking_shipments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_order
    ON shipping_tracking_shipments (order_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_provider
    ON shipping_tracking_shipments (tracking_provider_id);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_tracking_number
    ON shipping_tracking_shipments (tracking_number);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_provider_code
    ON shipping_tracking_shipments (provider_carrier_code);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_carrier
    ON shipping_tracking_shipments (carrier_id);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_service
    ON shipping_tracking_shipments (carrier_service_id);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_mapping
    ON shipping_tracking_shipments (tracking_carrier_mapping_id);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_registration
    ON shipping_tracking_shipments (registration_status);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_sync
    ON shipping_tracking_shipments (sync_status);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_next_sync
    ON shipping_tracking_shipments (next_sync_at);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_enabled
    ON shipping_tracking_shipments (enabled);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_deleted_at
    ON shipping_tracking_shipments (deleted_at);
