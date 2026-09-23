ALTER TABLE shipping_carrier_services
    ADD COLUMN IF NOT EXISTS remote_surcharge_minor BIGINT;

UPDATE shipping_carrier_services
SET remote_surcharge_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(remote_surcharge, 0))::BIGINT
    ELSE ROUND(COALESCE(remote_surcharge, 0) * 100)::BIGINT
END
WHERE remote_surcharge_minor IS NULL;

ALTER TABLE shipping_carrier_services
    ALTER COLUMN remote_surcharge_minor SET NOT NULL,
    ALTER COLUMN remote_surcharge_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_shipping_carrier_services_remote_surcharge_minor_non_negative
        CHECK (remote_surcharge_minor >= 0);

ALTER TABLE shipping_carrier_services
    DROP COLUMN IF EXISTS remote_surcharge;
