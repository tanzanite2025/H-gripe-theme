-- Tagged-product snapshots are transactional catalog amounts. Persist only the
-- exact smallest-unit integer representation and remove the floating major
-- column before launch.
ALTER TABLE workbench_feed_tagged_products
    ADD COLUMN IF NOT EXISTS price_minor BIGINT;

UPDATE workbench_feed_tagged_products
SET price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(COALESCE(price, 0))::BIGINT
        ELSE ROUND(COALESCE(price, 0) * 100)::BIGINT
    END
WHERE price_minor IS NULL;

ALTER TABLE workbench_feed_tagged_products
    ALTER COLUMN price_minor SET NOT NULL,
    ALTER COLUMN price_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_workbench_feed_tagged_products_price_minor_non_negative
        CHECK (price_minor >= 0),
    DROP COLUMN IF EXISTS price;
