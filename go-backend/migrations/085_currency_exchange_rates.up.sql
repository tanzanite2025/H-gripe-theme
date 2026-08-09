CREATE TABLE IF NOT EXISTS currency_exchange_rates (
    id BIGSERIAL PRIMARY KEY,
    base_currency VARCHAR(3) NOT NULL,
    quote_currency VARCHAR(3) NOT NULL,
    rate DOUBLE PRECISION NOT NULL,
    source VARCHAR(80) NOT NULL DEFAULT '',
    fetched_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_currency_exchange_rates_base_iso CHECK (base_currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_currency_exchange_rates_quote_iso CHECK (quote_currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_currency_exchange_rates_positive CHECK (rate > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_currency_exchange_rate_pair
    ON currency_exchange_rates(base_currency, quote_currency)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_currency_exchange_rates_base
    ON currency_exchange_rates(base_currency, expires_at)
    WHERE deleted_at IS NULL;
