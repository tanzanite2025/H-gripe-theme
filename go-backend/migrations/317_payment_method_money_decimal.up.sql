ALTER TABLE payment_methods
    ADD COLUMN IF NOT EXISTS fee_rate_decimal NUMERIC(30,15),
    ADD COLUMN IF NOT EXISTS fee_value_minor BIGINT;

UPDATE payment_methods
-- A legacy fixed fee has no currency, so it cannot be migrated faithfully.
-- This project has not launched: reset it and require a currency-aware
-- configuration before release. Percentage fees are unitless and survive.
SET fee_rate_decimal = CASE
        WHEN LOWER(COALESCE(fee_type, 'fixed')) = 'percentage'
            THEN ROUND(COALESCE(fee_value, 0)::NUMERIC, 15)
        ELSE 0
    END,
    fee_value_minor = 0
WHERE fee_rate_decimal IS NULL OR fee_value_minor IS NULL;

ALTER TABLE payment_methods
    ALTER COLUMN fee_rate_decimal SET NOT NULL,
    ALTER COLUMN fee_rate_decimal SET DEFAULT 0,
    ALTER COLUMN fee_value_minor SET NOT NULL,
    ALTER COLUMN fee_value_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_payment_methods_fee_rate_decimal
        CHECK (fee_rate_decimal >= 0 AND fee_rate_decimal <= 100),
    ADD CONSTRAINT chk_payment_methods_fee_value_minor_non_negative
        CHECK (fee_value_minor >= 0);
