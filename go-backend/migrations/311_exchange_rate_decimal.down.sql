ALTER TABLE currency_exchange_rates
    DROP CONSTRAINT IF EXISTS chk_currency_exchange_rates_rate_decimal_positive;

ALTER TABLE currency_exchange_rates
    DROP COLUMN IF EXISTS rate_decimal;
