-- Restore the pre-337 storage shape only for an explicit rollback. Values are
-- reconstructed from the canonical decimal column at the legacy precision.
ALTER TABLE currency_exchange_rates
    ADD COLUMN IF NOT EXISTS rate NUMERIC(20,10);

UPDATE currency_exchange_rates
SET rate = rate_decimal::NUMERIC(20,10)
WHERE rate IS NULL;

ALTER TABLE currency_exchange_rates
    ALTER COLUMN rate SET NOT NULL,
    ALTER COLUMN rate SET DEFAULT 0,
    DROP CONSTRAINT IF EXISTS chk_currency_exchange_rates_rate_decimal_positive,
    ADD CONSTRAINT chk_currency_exchange_rates_positive
        CHECK (rate > 0),
    DROP COLUMN IF EXISTS rate_decimal;
