-- Fulfillment tracking is package-scoped. Orders no longer carry a primary
-- tracking projection; all package facts live in shipping_tracking_shipments.
ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS fk_orders_tracking_provider,
    DROP CONSTRAINT IF EXISTS fk_orders_carrier,
    DROP CONSTRAINT IF EXISTS fk_orders_carrier_service,
    DROP CONSTRAINT IF EXISTS fk_orders_tracking_carrier_mapping;

DROP INDEX IF EXISTS idx_orders_tracking_provider_id;
DROP INDEX IF EXISTS idx_orders_carrier_id;
DROP INDEX IF EXISTS idx_orders_carrier_service_id;
DROP INDEX IF EXISTS idx_orders_tracking_carrier_mapping_id;
DROP INDEX IF EXISTS idx_orders_provider_carrier_code;

ALTER TABLE orders
    DROP COLUMN IF EXISTS tracking_number,
    DROP COLUMN IF EXISTS tracking_provider_id,
    DROP COLUMN IF EXISTS carrier_id,
    DROP COLUMN IF EXISTS carrier_service_id,
    DROP COLUMN IF EXISTS tracking_carrier_mapping_id,
    DROP COLUMN IF EXISTS provider_carrier_code,
    DROP COLUMN IF EXISTS provider_carrier_name;
