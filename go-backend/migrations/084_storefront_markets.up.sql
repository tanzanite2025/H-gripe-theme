CREATE TABLE IF NOT EXISTS storefront_markets (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL,
    name VARCHAR(120) NOT NULL,
    default_locale VARCHAR(32) NOT NULL,
    supported_locales JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_currency VARCHAR(3) NOT NULL,
    display_currencies JSONB NOT NULL DEFAULT '[]'::jsonb,
    payment_method_policy VARCHAR(80) NOT NULL DEFAULT '',
    logistics_policy VARCHAR(80) NOT NULL DEFAULT '',
    tax_policy VARCHAR(80) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_storefront_markets_code_active
    ON storefront_markets(UPPER(code))
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_storefront_markets_enabled_priority
    ON storefront_markets(enabled, priority, code)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS storefront_market_countries (
    id BIGSERIAL PRIMARY KEY,
    market_id BIGINT NOT NULL REFERENCES storefront_markets(id) ON DELETE CASCADE,
    code VARCHAR(2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_storefront_market_countries_code_active
    ON storefront_market_countries(UPPER(code))
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_storefront_market_countries_market_id
    ON storefront_market_countries(market_id)
    WHERE deleted_at IS NULL;

INSERT INTO storefront_markets (
    code, name, default_locale, supported_locales, default_currency,
    display_currencies, priority, enabled, created_at, updated_at
) VALUES
    ('US', 'United States', 'en', '["en", "es"]'::jsonb, 'USD', '["USD"]'::jsonb, 10, TRUE, NOW(), NOW()),
    ('EU', 'European Union', 'en', '["en", "de", "fr", "es", "it", "nl"]'::jsonb, 'EUR', '["EUR", "USD", "GBP"]'::jsonb, 20, TRUE, NOW(), NOW()),
    ('UK', 'United Kingdom', 'en', '["en"]'::jsonb, 'GBP', '["GBP", "USD", "EUR"]'::jsonb, 30, TRUE, NOW(), NOW()),
    ('CA', 'Canada', 'en', '["en", "fr"]'::jsonb, 'CAD', '["CAD", "USD", "EUR"]'::jsonb, 40, TRUE, NOW(), NOW()),
    ('CN', 'China Mainland', 'zh_cn', '["zh_cn", "en"]'::jsonb, 'CNY', '["CNY", "USD"]'::jsonb, 50, TRUE, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO storefront_market_countries (market_id, code, created_at, updated_at)
SELECT m.id, c.code, NOW(), NOW()
FROM storefront_markets m
JOIN (VALUES
    ('US', 'US'),
    ('EU', 'AT'), ('EU', 'BE'), ('EU', 'DE'), ('EU', 'ES'), ('EU', 'FI'),
    ('EU', 'FR'), ('EU', 'IE'), ('EU', 'IT'), ('EU', 'LU'), ('EU', 'NL'), ('EU', 'PT'),
    ('UK', 'GB'),
    ('CA', 'CA'),
    ('CN', 'CN')
) AS c(market_code, code) ON c.market_code = m.code
ON CONFLICT DO NOTHING;
