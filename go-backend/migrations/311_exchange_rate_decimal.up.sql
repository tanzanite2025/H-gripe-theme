ALTER TABLE currency_exchange_rates
    ADD COLUMN IF NOT EXISTS rate_decimal NUMERIC(30,15);

UPDATE currency_exchange_rates
SET rate_decimal = rate::NUMERIC(30,15)
WHERE rate_decimal IS NULL OR rate_decimal = 0;

ALTER TABLE currency_exchange_rates
    ALTER COLUMN rate_decimal SET NOT NULL,
    ADD CONSTRAINT chk_currency_exchange_rates_rate_decimal_positive
        CHECK (rate_decimal > 0);
