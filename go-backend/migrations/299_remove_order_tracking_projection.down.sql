ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(255),
    ADD COLUMN IF NOT EXISTS tracking_provider_id BIGINT,
    ADD COLUMN IF NOT EXISTS carrier_id BIGINT,
    ADD COLUMN IF NOT EXISTS carrier_service_id BIGINT,
    ADD COLUMN IF NOT EXISTS tracking_carrier_mapping_id BIGINT,
    ADD COLUMN IF NOT EXISTS provider_carrier_code TEXT,
    ADD COLUMN IF NOT EXISTS provider_carrier_name TEXT;

CREATE INDEX IF NOT EXISTS idx_orders_tracking_provider_id ON orders (tracking_provider_id);
CREATE INDEX IF NOT EXISTS idx_orders_carrier_id ON orders (carrier_id);
CREATE INDEX IF NOT EXISTS idx_orders_carrier_service_id ON orders (carrier_service_id);
CREATE INDEX IF NOT EXISTS idx_orders_tracking_carrier_mapping_id ON orders (tracking_carrier_mapping_id);
CREATE INDEX IF NOT EXISTS idx_orders_provider_carrier_code ON orders (provider_carrier_code);
