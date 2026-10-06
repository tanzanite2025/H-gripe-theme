ALTER TABLE shipping_carrier_services
    DROP CONSTRAINT IF EXISTS chk_shipping_carrier_services_single_published_collection,
    DROP CONSTRAINT IF EXISTS fk_shipping_carrier_services_fpx_channel,
    DROP CONSTRAINT IF EXISTS fk_shipping_carrier_services_yanwen_channel;

DROP INDEX IF EXISTS idx_shipping_carrier_services_fpx_channel_id;
DROP INDEX IF EXISTS idx_shipping_carrier_services_yanwen_channel_id;

ALTER TABLE shipping_carrier_services
    DROP COLUMN IF EXISTS fpx_channel_id,
    DROP COLUMN IF EXISTS yanwen_published_channel_id;
