ALTER TABLE tax_rates
    ADD COLUMN IF NOT EXISTS rate DOUBLE PRECISION;

UPDATE tax_rates
SET rate = rate_decimal::DOUBLE PRECISION
WHERE rate IS NULL;

ALTER TABLE tax_rates
    ALTER COLUMN rate SET NOT NULL,
    ALTER COLUMN rate SET DEFAULT 0,
    DROP CONSTRAINT IF EXISTS chk_tax_rates_rate_decimal_non_negative,
    DROP CONSTRAINT IF EXISTS chk_tax_rates_rate_decimal,
    DROP COLUMN IF EXISTS rate_decimal;
