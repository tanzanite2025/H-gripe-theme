ALTER TABLE tax_rates
    DROP CONSTRAINT IF EXISTS chk_tax_rates_rate_decimal;

ALTER TABLE tax_rates
    DROP COLUMN IF EXISTS rate_decimal;

