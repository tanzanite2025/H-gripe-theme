ALTER TABLE google_merchant_offers
    ADD COLUMN IF NOT EXISTS price_override_minor BIGINT,
    ADD COLUMN IF NOT EXISTS sale_price_override_minor BIGINT;

UPDATE google_merchant_offers
SET price_override_minor = CASE
        WHEN price_override IS NULL THEN NULL
        WHEN UPPER(COALESCE(currency_code, 'USD')) IN ('JPY','KRW','CLP') THEN ROUND(price_override)::BIGINT
        ELSE ROUND(price_override * 100)::BIGINT
    END,
    sale_price_override_minor = CASE
        WHEN sale_price_override IS NULL THEN NULL
        WHEN UPPER(COALESCE(currency_code, 'USD')) IN ('JPY','KRW','CLP') THEN ROUND(sale_price_override)::BIGINT
        ELSE ROUND(sale_price_override * 100)::BIGINT
    END
WHERE price_override_minor IS NULL OR sale_price_override_minor IS NULL;

ALTER TABLE google_merchant_offers
    ADD CONSTRAINT chk_google_merchant_offers_price_override_minor_positive
        CHECK (price_override_minor IS NULL OR price_override_minor > 0),
    ADD CONSTRAINT chk_google_merchant_offers_sale_price_override_minor_positive
        CHECK (sale_price_override_minor IS NULL OR sale_price_override_minor > 0),
    DROP COLUMN IF EXISTS price_override,
    DROP COLUMN IF EXISTS sale_price_override;
