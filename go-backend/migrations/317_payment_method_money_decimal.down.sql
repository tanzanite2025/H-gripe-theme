ALTER TABLE payment_methods
    DROP CONSTRAINT IF EXISTS chk_payment_methods_fee_rate_decimal,
    DROP CONSTRAINT IF EXISTS chk_payment_methods_fee_value_minor_non_negative,
    DROP COLUMN IF EXISTS fee_rate_decimal,
    DROP COLUMN IF EXISTS fee_value_minor;

