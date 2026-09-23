-- Restore the pre-339 field only for an explicit rollback. The canonical
-- decimal value is rounded to the legacy three-decimal precision.
ALTER TABLE shipping_carrier_services
    ADD COLUMN IF NOT EXISTS fuel_surcharge_percent NUMERIC(8,3);

UPDATE shipping_carrier_services
SET fuel_surcharge_percent = fuel_surcharge_percent_decimal::NUMERIC(8,3)
WHERE fuel_surcharge_percent IS NULL;

ALTER TABLE shipping_carrier_services
    ALTER COLUMN fuel_surcharge_percent SET NOT NULL,
    ALTER COLUMN fuel_surcharge_percent SET DEFAULT 0,
    DROP CONSTRAINT IF EXISTS chk_shipping_carrier_services_fuel_surcharge_percent_decimal_range,
    DROP COLUMN IF EXISTS fuel_surcharge_percent_decimal;
