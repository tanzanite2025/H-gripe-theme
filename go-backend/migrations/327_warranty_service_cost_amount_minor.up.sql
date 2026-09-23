ALTER TABLE warranty_service_records
    ADD COLUMN IF NOT EXISTS cost_amount_minor BIGINT;

UPDATE warranty_service_records
SET cost_amount_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(cost_amount, 0))::BIGINT
    ELSE ROUND(COALESCE(cost_amount, 0) * 100)::BIGINT
END
WHERE cost_amount_minor IS NULL;

ALTER TABLE warranty_service_records
    ALTER COLUMN cost_amount_minor SET NOT NULL,
    ALTER COLUMN cost_amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_warranty_service_records_cost_amount_minor_non_negative
        CHECK (cost_amount_minor >= 0),
    DROP COLUMN IF EXISTS cost_amount;
