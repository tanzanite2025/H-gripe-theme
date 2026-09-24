ALTER TABLE tax_rates
    ADD COLUMN IF NOT EXISTS rate_decimal NUMERIC(30,15);

UPDATE tax_rates
SET rate_decimal = ROUND(COALESCE(rate, 0)::NUMERIC, 15)
WHERE rate_decimal IS NULL;

ALTER TABLE tax_rates
    ALTER COLUMN rate_decimal SET NOT NULL,
    ALTER COLUMN rate_decimal SET DEFAULT 0,
    ADD CONSTRAINT chk_tax_rates_rate_decimal
        CHECK (rate_decimal >= 0 AND rate_decimal <= 100);

