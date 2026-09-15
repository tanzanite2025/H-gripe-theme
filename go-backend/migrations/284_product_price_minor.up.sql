-- Product source prices are monetary values, so persist them as integer minor
-- units. The legacy major-unit columns remain untouched in this migration;
-- they are removed after all read/write paths use the Money boundary.
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS sale_price_minor BIGINT;

UPDATE products
SET price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(price)::BIGINT
        ELSE ROUND(price * 100)::BIGINT
    END
WHERE price_minor IS NULL;

UPDATE products
SET sale_price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(sale_price)::BIGINT
        ELSE ROUND(sale_price * 100)::BIGINT
    END
WHERE sale_price IS NOT NULL AND sale_price_minor IS NULL;

ALTER TABLE products
    ALTER COLUMN price_minor SET NOT NULL,
    ADD CONSTRAINT chk_products_price_minor_non_negative CHECK (price_minor >= 0),
    ADD CONSTRAINT chk_products_sale_price_minor_non_negative CHECK (sale_price_minor IS NULL OR sale_price_minor >= 0);

CREATE INDEX IF NOT EXISTS idx_products_price_currency ON products(currency, price_minor);

ALTER TABLE product_variants
    ADD COLUMN IF NOT EXISTS price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS sale_price_minor BIGINT;

UPDATE product_variants
SET price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(price)::BIGINT
        ELSE ROUND(price * 100)::BIGINT
    END
WHERE price_minor IS NULL;

UPDATE product_variants
SET sale_price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(sale_price)::BIGINT
        ELSE ROUND(sale_price * 100)::BIGINT
    END
WHERE sale_price IS NOT NULL AND sale_price_minor IS NULL;

ALTER TABLE product_variants
    ALTER COLUMN price_minor SET NOT NULL,
    ADD CONSTRAINT chk_product_variants_price_minor_non_negative CHECK (price_minor >= 0),
    ADD CONSTRAINT chk_product_variants_sale_price_minor_non_negative CHECK (sale_price_minor IS NULL OR sale_price_minor >= 0);

CREATE INDEX IF NOT EXISTS idx_product_variants_price_currency ON product_variants(currency, price_minor);
