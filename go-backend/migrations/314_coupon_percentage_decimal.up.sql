ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS value_rate_decimal NUMERIC(30,15);

UPDATE coupons
SET value_rate_decimal = ROUND(COALESCE(value, 0)::NUMERIC, 15)
WHERE value_rate_decimal IS NULL;

ALTER TABLE coupons
    ALTER COLUMN value_rate_decimal SET NOT NULL,
    ALTER COLUMN value_rate_decimal SET DEFAULT 0,
    ADD CONSTRAINT chk_coupons_value_rate_decimal
        CHECK (value_rate_decimal >= 0 AND value_rate_decimal <= 100);

