-- Exchange rates are transactional decimal values. The legacy floating-point
-- projection is removed before launch so conversions cannot silently re-enter
-- IEEE-754 arithmetic through the persistence model.
ALTER TABLE currency_exchange_rates
    ALTER COLUMN rate_decimal TYPE NUMERIC(30,15),
    ALTER COLUMN rate_decimal SET DEFAULT 0;

UPDATE currency_exchange_rates
SET rate_decimal = ROUND(COALESCE(rate_decimal, rate)::NUMERIC, 15)
WHERE rate_decimal IS NULL OR rate_decimal = 0;

ALTER TABLE currency_exchange_rates
    ALTER COLUMN rate_decimal SET NOT NULL,
    DROP CONSTRAINT IF EXISTS chk_currency_exchange_rates_positive,
    DROP CONSTRAINT IF EXISTS chk_currency_exchange_rates_rate_decimal_positive,
    ADD CONSTRAINT chk_currency_exchange_rates_rate_decimal_positive
        CHECK (rate_decimal > 0),
    DROP COLUMN IF EXISTS rate;
