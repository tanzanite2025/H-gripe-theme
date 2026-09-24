-- Fuel surcharge percentages are transactional pricing inputs. Persist the
-- exact decimal value before removing the legacy low-precision numeric field.
ALTER TABLE shipping_carrier_services
    ADD COLUMN IF NOT EXISTS fuel_surcharge_percent_decimal NUMERIC(30,15);

UPDATE shipping_carrier_services
SET fuel_surcharge_percent_decimal = ROUND(COALESCE(fuel_surcharge_percent, 0)::NUMERIC, 15)
WHERE fuel_surcharge_percent_decimal IS NULL;

ALTER TABLE shipping_carrier_services
    ALTER COLUMN fuel_surcharge_percent_decimal SET NOT NULL,
    ALTER COLUMN fuel_surcharge_percent_decimal SET DEFAULT 0,
    DROP CONSTRAINT IF EXISTS chk_shipping_carrier_services_fuel_surcharge_percent_decimal_range,
    ADD CONSTRAINT chk_shipping_carrier_services_fuel_surcharge_percent_decimal_range
        CHECK (fuel_surcharge_percent_decimal >= 0 AND fuel_surcharge_percent_decimal <= 100),
    DROP COLUMN IF EXISTS fuel_surcharge_percent;
