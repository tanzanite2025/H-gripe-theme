ALTER TABLE shipping_carrier_services
    ADD COLUMN IF NOT EXISTS remote_surcharge DOUBLE PRECISION;

ALTER TABLE shipping_carrier_services
    DROP CONSTRAINT IF EXISTS chk_shipping_carrier_services_remote_surcharge_minor_non_negative;

ALTER TABLE shipping_carrier_services
    DROP COLUMN IF EXISTS remote_surcharge_minor;
