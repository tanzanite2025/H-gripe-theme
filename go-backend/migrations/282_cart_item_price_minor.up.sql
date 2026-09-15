-- Cart snapshots are transactional money values. Convert the last remaining
-- cart-item major-unit column to integer minor units before removing it.
ALTER TABLE cart_items
    ADD COLUMN IF NOT EXISTS price_minor BIGINT;

UPDATE cart_items
SET price_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) = 'JPY'
        THEN ROUND(price)::BIGINT
    ELSE ROUND(price * 100)::BIGINT
END
WHERE price_minor IS NULL;

ALTER TABLE cart_items
    ALTER COLUMN price_minor SET NOT NULL,
    ADD CONSTRAINT chk_cart_items_price_minor_non_negative CHECK (price_minor >= 0);

ALTER TABLE cart_items
    DROP COLUMN IF EXISTS price;

CREATE INDEX IF NOT EXISTS idx_cart_items_price_currency
    ON cart_items(currency, price_minor);
