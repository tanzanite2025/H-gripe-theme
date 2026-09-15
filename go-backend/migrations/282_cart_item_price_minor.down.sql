ALTER TABLE cart_items
    ADD COLUMN IF NOT EXISTS price NUMERIC(18,2);

UPDATE cart_items
SET price = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) = 'JPY'
        THEN price_minor::NUMERIC
    ELSE price_minor::NUMERIC / 100
END
WHERE price IS NULL;

ALTER TABLE cart_items
    DROP CONSTRAINT IF EXISTS chk_cart_items_price_minor_non_negative;

ALTER TABLE cart_items
    DROP COLUMN IF EXISTS price_minor;

DROP INDEX IF EXISTS idx_cart_items_price_currency;
