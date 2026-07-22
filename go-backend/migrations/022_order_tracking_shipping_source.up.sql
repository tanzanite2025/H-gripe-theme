ALTER TABLE orders DROP COLUMN IF EXISTS carrier_code;

ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_provider_id BIGINT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS carrier_id BIGINT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS carrier_service_id BIGINT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_carrier_mapping_id BIGINT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS provider_carrier_code TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS provider_carrier_name TEXT;

CREATE INDEX IF NOT EXISTS idx_orders_tracking_provider_id ON orders (tracking_provider_id);
CREATE INDEX IF NOT EXISTS idx_orders_carrier_id ON orders (carrier_id);
CREATE INDEX IF NOT EXISTS idx_orders_carrier_service_id ON orders (carrier_service_id);
CREATE INDEX IF NOT EXISTS idx_orders_tracking_carrier_mapping_id ON orders (tracking_carrier_mapping_id);
CREATE INDEX IF NOT EXISTS idx_orders_provider_carrier_code ON orders (provider_carrier_code);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_orders_tracking_provider') THEN
        ALTER TABLE orders
            ADD CONSTRAINT fk_orders_tracking_provider
            FOREIGN KEY (tracking_provider_id) REFERENCES shipping_tracking_providers(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_orders_carrier') THEN
        ALTER TABLE orders
            ADD CONSTRAINT fk_orders_carrier
            FOREIGN KEY (carrier_id) REFERENCES carriers(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_orders_carrier_service') THEN
        ALTER TABLE orders
            ADD CONSTRAINT fk_orders_carrier_service
            FOREIGN KEY (carrier_service_id) REFERENCES shipping_carrier_services(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_orders_tracking_carrier_mapping') THEN
        ALTER TABLE orders
            ADD CONSTRAINT fk_orders_tracking_carrier_mapping
            FOREIGN KEY (tracking_carrier_mapping_id) REFERENCES shipping_tracking_carrier_mappings(id) ON DELETE SET NULL;
    END IF;
END $$;
